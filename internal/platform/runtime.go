package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Runtime interface {
	LoadSession(ctx context.Context, sessionID string) (SessionState, error)
	RunCommand(ctx context.Context, sessionID string, command string) (CommandResponse, error)
	Execute(ctx context.Context, request CommandRequest) (CommandResponse, error)
	ObserveEvent(ctx context.Context, event SessionEvent) (SessionState, error)
	PubsubName() string
	StateStoreName() string
	LockStoreName() string
}

type LocalRuntime struct {
	mu     sync.RWMutex
	states map[string]SessionState
}

type DaprRuntime struct {
	httpClient *http.Client
	baseURL    string
	stateStore string
	pubsubName string
	lockStore  string
	linuxAppID string
}

type lockRequest struct {
	ResourceID      string `json:"resourceId"`
	LockOwner       string `json:"lockOwner"`
	ExpiryInSeconds int    `json:"expiryInSeconds,omitempty"`
}

type lockResponse struct {
	Success any `json:"success"`
}

type unlockResponse struct {
	Status int `json:"status"`
}

func NewLocalRuntime() Runtime {
	return &LocalRuntime{states: make(map[string]SessionState)}
}

func NewRuntimeFromEnv() Runtime {
	port := strings.TrimSpace(os.Getenv("DAPR_HTTP_PORT"))
	if port == "" {
		return NewLocalRuntime()
	}

	return &DaprRuntime{
		httpClient: &http.Client{Timeout: 8 * time.Second},
		baseURL:    fmt.Sprintf("http://127.0.0.1:%s/v1.0", port),
		stateStore: envOrDefault("DAPR_STATE_STORE_NAME", DefaultStateStore),
		pubsubName: envOrDefault("DAPR_PUBSUB_NAME", DefaultPubsubName),
		lockStore:  envOrDefault("DAPR_LOCK_STORE_NAME", DefaultLockStore),
		linuxAppID: envOrDefault("PLATFORM_LINUX_APP_ID", LinuxAppID),
	}
}

func (r *LocalRuntime) LoadSession(_ context.Context, sessionID string) (SessionState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.ensureLocked(sessionID), nil
}

func (r *LocalRuntime) RunCommand(ctx context.Context, sessionID string, command string) (CommandResponse, error) {
	return r.Execute(ctx, CommandRequest{
		SessionID: sessionID,
		Command:   command,
		Origin:    TerminalAppID,
	})
}

func (r *LocalRuntime) Execute(_ context.Context, request CommandRequest) (CommandResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := r.ensureLocked(request.SessionID)
	response := applyCommand(state, request)
	if response.Accepted {
		response.State = observeShellEvent(response.State, response.Event)
	}
	r.states[request.SessionID] = response.State

	return response, nil
}

func (r *LocalRuntime) ObserveEvent(_ context.Context, event SessionEvent) (SessionState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state := r.ensureLocked(event.SessionID)
	state = observeShellEvent(state, event)
	r.states[event.SessionID] = state

	return state, nil
}

func (r *LocalRuntime) ensureLocked(sessionID string) SessionState {
	if state, ok := r.states[sessionID]; ok {
		return state
	}

	state := NewSessionState(sessionID)
	r.states[sessionID] = state
	return state
}

func (r *LocalRuntime) PubsubName() string {
	return DefaultPubsubName
}

func (r *LocalRuntime) StateStoreName() string {
	return DefaultStateStore
}

func (r *LocalRuntime) LockStoreName() string {
	return DefaultLockStore
}

func (r *DaprRuntime) LoadSession(ctx context.Context, sessionID string) (SessionState, error) {
	body, status, err := r.request(ctx, http.MethodGet, "/state/"+r.stateStore+"/"+url.PathEscape(SessionStateKey(sessionID)), nil)
	if err != nil {
		return SessionState{}, err
	}

	if status == http.StatusNoContent || len(bytes.TrimSpace(body)) == 0 || string(body) == "null" {
		state := NewSessionState(sessionID)
		if err := r.saveSession(ctx, state); err != nil {
			return SessionState{}, err
		}
		return state, nil
	}

	var state SessionState
	if err := json.Unmarshal(body, &state); err != nil {
		return SessionState{}, fmt.Errorf("decode state: %w", err)
	}
	if state.SessionID == "" {
		state = NewSessionState(sessionID)
	}

	return state, nil
}

func (r *DaprRuntime) RunCommand(ctx context.Context, sessionID string, command string) (CommandResponse, error) {
	request := CommandRequest{
		SessionID: sessionID,
		Command:   command,
		Origin:    TerminalAppID,
	}

	body, _, err := r.request(ctx, http.MethodPost, "/invoke/"+url.PathEscape(r.linuxAppID)+"/method/internal/execute", request)
	if err != nil {
		return CommandResponse{}, err
	}

	var response CommandResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return CommandResponse{}, fmt.Errorf("decode invoke response: %w", err)
	}

	return response, nil
}

func (r *DaprRuntime) Execute(ctx context.Context, request CommandRequest) (CommandResponse, error) {
	unlock, err := r.lockSession(ctx, request.SessionID)
	if err != nil {
		return CommandResponse{}, err
	}
	defer unlock()

	state, err := r.LoadSession(ctx, request.SessionID)
	if err != nil {
		return CommandResponse{}, err
	}

	response := applyCommand(state, request)
	if err := r.saveSession(ctx, response.State); err != nil {
		return CommandResponse{}, err
	}
	if response.Accepted {
		if err := r.publishEvent(ctx, response.Event); err != nil {
			return CommandResponse{}, err
		}
	}

	return response, nil
}

func (r *DaprRuntime) ObserveEvent(ctx context.Context, event SessionEvent) (SessionState, error) {
	unlock, err := r.lockSession(ctx, event.SessionID)
	if err != nil {
		return SessionState{}, err
	}
	defer unlock()

	state, err := r.LoadSession(ctx, event.SessionID)
	if err != nil {
		return SessionState{}, err
	}

	state = observeShellEvent(state, event)
	if err := r.saveSession(ctx, state); err != nil {
		return SessionState{}, err
	}

	return state, nil
}

func (r *DaprRuntime) PubsubName() string {
	return r.pubsubName
}

func (r *DaprRuntime) StateStoreName() string {
	return r.stateStore
}

func (r *DaprRuntime) LockStoreName() string {
	return r.lockStore
}

func (r *DaprRuntime) saveSession(ctx context.Context, state SessionState) error {
	payload := []map[string]any{{
		"key":   SessionStateKey(state.SessionID),
		"value": state,
	}}

	_, _, err := r.request(ctx, http.MethodPost, "/state/"+r.stateStore, payload)
	return err
}

func (r *DaprRuntime) publishEvent(ctx context.Context, event SessionEvent) error {
	_, _, err := r.request(ctx, http.MethodPost, "/publish/"+r.pubsubName+"/"+SessionEventsTopic, event)
	return err
}

func (r *DaprRuntime) lockSession(ctx context.Context, sessionID string) (func(), error) {
	if sessionID == "" {
		return func() {}, nil
	}

	owner := fmt.Sprintf("%s/%d", sessionID, time.Now().UnixNano())
	body, _, err := r.request(ctx, http.MethodPost, "/-alpha1/lock/"+r.lockStore, lockRequest{
		ResourceID:      SessionStateKey(sessionID),
		LockOwner:       owner,
		ExpiryInSeconds: 20,
	})
	if err != nil {
		return nil, err
	}

	var response lockResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode lock response: %w", err)
	}
	if !lockSucceeded(response.Success) {
		return nil, fmt.Errorf("lock session %s: dapr lock store rejected request", sessionID)
	}

	return func() {
		_, _, _ = r.request(context.Background(), http.MethodPost, "/-alpha1/unlock/"+r.lockStore, lockRequest{
			ResourceID: SessionStateKey(sessionID),
			LockOwner:  owner,
		})
	}, nil
}

func lockSucceeded(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(typed, "true")
	default:
		return false
	}
}

func (r *DaprRuntime) request(ctx context.Context, method string, path string, payload any) ([]byte, int, error) {
	ctx = ensureContext(ctx)

	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, r.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, resp.StatusCode, fmt.Errorf("request %s %s failed with %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, resp.StatusCode, nil
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
