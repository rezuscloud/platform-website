package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SigNozClient fetches live data from the SigNoz Prometheus-compatible API.
// Results are cached for 30s to avoid querying on every page render.
type SigNozClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cached     LiveData
	cachedAt   time.Time
	cacheTTL   time.Duration
}

func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	return &SigNozClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		cacheTTL:   30 * time.Second,
	}
}

// NewSigNozClientFromEnv creates a client from SIGNOZ_URL and SIGNOZ_API_KEY.
// Returns nil if either is missing.
func NewSigNozClientFromEnv() *SigNozClient {
	u := os.Getenv("SIGNOZ_URL")
	k := os.Getenv("SIGNOZ_API_KEY")
	if u == "" || k == "" {
		return nil
	}
	return NewSigNozClient(u, k)
}

// Fetch returns the platform-website service tree with live metrics.
// Returns cached data if less than 30s old. Falls back to topology-only on error.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && c.cached.Root.Name != "" {
		return c.cached, nil
	}

	root := PlatformTopology()

	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	c.populateStatus(fetchCtx, &root)
	c.populateMetrics(fetchCtx, &root)
	stats := c.populateStats(fetchCtx)
	health := c.populateHealth(&root)

	c.cached = LiveData{
		Root:       root,
		Stats:      stats,
		Health:     health,
		HasMetrics: true,
	}
	c.cachedAt = time.Now()

	return c.cached, nil
}

// populateStatus checks which services are reachable via their scraped metrics.
func (c *SigNozClient) populateStatus(ctx context.Context, root *ServiceNode) {
	// The Dapr sidecar (daprd) is the only container with prometheus.io annotations.
	// All other services are inferred from what we can query.
	queries := map[string]string{
		// Daprd up = entire pod is running (app + sidecar)
		"daprd":              `up{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`,
		"platform-website":   `up{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`,
		"dapr-control-plane": `up{k8s_namespace_name="dapr-system"}`,
	}

	root.Walk(func(s *ServiceNode) {
		q := queries[s.Name]
		if q == "" {
			s.Status = "unknown"
			return
		}
		results, err := c.queryInstant(ctx, q)
		if err != nil || len(results) == 0 {
			s.Status = "unknown"
			return
		}
		s.Status = "healthy"
	})
}

// populateMetrics fetches compact metric values for services that have them.
func (c *SigNozClient) populateMetrics(ctx context.Context, root *ServiceNode) {
	// All metrics come from the daprd sidecar's :9090 endpoint since the app
	// doesn't expose its own /metrics. goroutines/heap are daprd's runtime.
	type mq struct {
		name   string
		label  string
		unit   string
		query  string
		format func(float64) string
	}

	queries := []mq{
		{
			name: "platform-website", label: "goroutines", unit: "",
			query:  `go_goroutines{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`,
			format: func(v float64) string { return fmt.Sprintf("%.0f", v) },
		},
		{
			name: "platform-website", label: "heap", unit: "MiB",
			query:  `go_memstats_alloc_bytes{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`,
			format: func(v float64) string { return fmt.Sprintf("%.1f", v/(1024*1024)) },
		},
		{
			name: "daprd", label: "components", unit: "loaded",
			query:  `dapr_runtime_component_loaded{k8s_namespace_name="platform-website"}`,
			format: func(v float64) string { return fmt.Sprintf("%.0f", v) },
		},
	}

	for _, q := range queries {
		node := root.Find(q.name)
		if node == nil {
			continue
		}
		results, err := c.queryInstant(ctx, q.query)
		if err != nil || len(results) == 0 {
			continue
		}
		valStr := ""
		if len(results[0].Value) >= 2 {
			if f, err := parseFloat(results[0].Value[1]); err == nil {
				valStr = q.format(f)
			}
		}
		if valStr != "" {
			node.Metrics = append(node.Metrics, MetricSeries{
				Label: q.label,
				Value: valStr,
				Unit:  q.unit,
			})
		}
	}
}

func (c *SigNozClient) populateStats(ctx context.Context) StatsStrip {
	stats := StatsStrip{NodeCount: 2}

	// Get uptime from daprd's process_start_time_seconds
	results, err := c.queryInstant(ctx,
		`process_start_time_seconds{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`)
	if err == nil && len(results) > 0 && len(results[0].Value) >= 2 {
		if ts, err := parseFloat(results[0].Value[1]); err == nil {
			uptime := time.Since(time.Unix(int64(ts), 0))
			stats.Uptime = formatDuration(uptime)
		}
	}

	// Go version from go_info
	results, err = c.queryInstant(ctx,
		`go_info{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`)
	if err == nil && len(results) > 0 {
		if v, ok := results[0].Metric["version"]; ok {
			stats.GoVersion = v
		}
	}

	return stats
}

func (c *SigNozClient) populateHealth(root *ServiceNode) []HealthCheck {
	var checks []HealthCheck
	root.Walk(func(s *ServiceNode) {
		h := HealthCheck{
			ServiceName: s.Label,
			Status:      s.Status,
		}
		if s.Status == "healthy" {
			h.LastCheck = "5s ago"
		}
		checks = append(checks, h)
	})
	return checks
}

func parseFloat(v interface{}) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
	return f, err
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	if h >= 24 {
		days := h / 24
		return fmt.Sprintf("%dd", days)
	}
	if h >= 1 {
		return fmt.Sprintf("%dh", h)
	}
	m := int(d.Minutes())
	return fmt.Sprintf("%dm", m)
}

func (c *SigNozClient) queryInstant(ctx context.Context, query string) ([]promResultItem, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/query"
	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("SIGNOZ-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pr promResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if pr.Status != "success" {
		return nil, fmt.Errorf("sigNoz: %s", string(body))
	}
	return pr.Data.Result, nil
}

// Prometheus API response types

type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string           `json:"resultType"`
		Result     []promResultItem `json:"result"`
	} `json:"data"`
}

type promResultItem struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value,omitempty"`
}
