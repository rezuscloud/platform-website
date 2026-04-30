package platform

import (
	"fmt"
	"time"
)

const (
	SessionCookieName    = "rezus_session"
	SessionCookieMaxAge  = 60 * 60 * 24 * 30
	DefaultStateStore    = "statestore"
	DefaultPubsubName    = "pubsub"
	DefaultLockStore     = "lockstore"
	SessionEventsTopic   = "homepage.events"
	ShellAppID           = "platform-website-shell"
	TerminalAppID        = "platform-website-terminal"
	MacAppID             = "platform-website-mac"
	LinuxAppID           = "platform-website-linux"
	MaxTerminalHistory   = 18
	MaxSessionEventCount = 8
)

type Proof struct {
	Label  string `json:"label"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type SummaryState struct {
	Kicker   string  `json:"kicker"`
	Headline string  `json:"headline"`
	Detail   string  `json:"detail"`
	Proofs   []Proof `json:"proofs"`
}

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type Artifact struct {
	Title     string   `json:"title"`
	Lines     []string `json:"lines"`
	UpdatedAt string   `json:"updatedAt"`
}

type TerminalState struct {
	Prompt      string   `json:"prompt"`
	Suggestions []string `json:"suggestions"`
	History     []string `json:"history"`
	LastCommand string   `json:"lastCommand"`
}

type MacState struct {
	Artifact Artifact `json:"artifact"`
	Notes    []string `json:"notes"`
}

type LinuxState struct {
	Mode       string          `json:"mode"`
	LastAction string          `json:"lastAction"`
	Services   []ServiceStatus `json:"services"`
}

type SessionEvent struct {
	SessionID string   `json:"sessionId"`
	Type      string   `json:"type"`
	Source    string   `json:"source"`
	Message   string   `json:"message"`
	Details   []string `json:"details"`
	Timestamp string   `json:"timestamp"`
}

type DaprSubscription struct {
	PubsubName string `json:"pubsubname"`
	Topic      string `json:"topic"`
	Route      string `json:"route"`
}

type SessionState struct {
	SessionID string         `json:"sessionId"`
	Summary   SummaryState   `json:"summary"`
	Terminal  TerminalState  `json:"terminal"`
	Mac       MacState       `json:"mac"`
	Linux     LinuxState     `json:"linux"`
	Events    []SessionEvent `json:"events"`
	UpdatedAt string         `json:"updatedAt"`
}

type CommandRequest struct {
	SessionID string `json:"sessionId"`
	Command   string `json:"command"`
	Origin    string `json:"origin"`
}

type CommandResponse struct {
	Accepted bool         `json:"accepted"`
	Message  string       `json:"message"`
	Event    SessionEvent `json:"event"`
	State    SessionState `json:"state"`
}

func SessionStateKey(sessionID string) string {
	return fmt.Sprintf("homepage/sessions/%s/state", sessionID)
}

func NewSessionState(sessionID string) SessionState {
	now := time.Now().UTC().Format(time.RFC3339)

	return SessionState{
		SessionID: sessionID,
		Summary: SummaryState{
			Kicker:   "Brand shell, live demos",
			Headline: "One homepage, three cooperating application surfaces",
			Detail:   "Run a terminal flow. Linux persists topology in PostgreSQL V2 state, publishes over NATS JetStream, coordinates with Redis locking, and the shell updates the proof rail.",
			Proofs: []Proof{
				{Label: "Invoke", Status: "armed", Detail: "terminal-app is ready to call linux-app"},
				{Label: "Pub/Sub", Status: "idle", Detail: "linux-app is waiting to publish a session event on NATS JetStream"},
				{Label: "State", Status: "cold", Detail: "shell-app and mac-app have not observed PostgreSQL-backed shared state yet"},
				{Label: "Lock", Status: "ready", Detail: "Redis lockstore is ready to serialize session updates"},
			},
		},
		Terminal: TerminalState{
			Prompt:      "rezus@terminal",
			Suggestions: []string{"rezus sync demo", "rezus fanout edge", "rezus inspect dossier"},
			History: []string{
				"\x1b[32mBOOT:\x1b[0m shell, terminal, mac, linux surfaces attached",
				"\x1b[36mDAPR:\x1b[0m PostgreSQL v2 state + JetStream pubsub + Redis lock ready",
				"\x1b[2mTIP: run `rezus sync demo` to prove the cross-app path\x1b[0m",
			},
		},
		Mac: MacState{
			Artifact: Artifact{
				Title:     "Artifact drawer empty",
				Lines:     []string{"No topology dossier has been published yet.", "Run a terminal flow to materialize PostgreSQL-backed shared state."},
				UpdatedAt: now,
			},
			Notes: []string{
				"Mac is the inspection surface.",
				"It reads PostgreSQL-backed shared state instead of inventing local copy.",
			},
		},
		Linux: LinuxState{
			Mode:       "idle",
			LastAction: "Waiting for the first invoke from terminal-app through Dapr service invocation.",
			Services: []ServiceStatus{
				{Name: "shell-app", Status: "watching", Detail: "renders the proof rail from PostgreSQL-backed shared state"},
				{Name: "terminal-app", Status: "ready", Detail: "collects intent and dispatches flows"},
				{Name: "mac-app", Status: "standby", Detail: "opens artifacts once they exist"},
				{Name: "linux-app", Status: "idle", Detail: "waiting to reconcile topology with Redis-backed locking"},
			},
		},
		Events: []SessionEvent{
			{
				SessionID: sessionID,
				Type:      "session.ready",
				Source:    ShellAppID,
				Message:   "Session initialized. The machine is waiting for a real flow.",
				Details:   []string{"Shell has not seen an invoke yet.", "Mac and Linux are showing baseline state.", "Namespace-local PostgreSQL, JetStream, and Redis backends are available."},
				Timestamp: now,
			},
		},
		UpdatedAt: now,
	}
}
