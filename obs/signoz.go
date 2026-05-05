package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// SigNozClient fetches live data from the SigNoz Prometheus-compatible API.
type SigNozClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	return &SigNozClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
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
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	root := PlatformTopology()

	c.populateStatus(ctx, &root)
	c.populateMetrics(ctx, &root)
	stats := c.populateStats(ctx)
	health := c.populateHealth(ctx, &root)

	return LiveData{
		Root:       root,
		Stats:      stats,
		Health:     health,
		HasMetrics: true,
	}, nil
}

func (c *SigNozClient) populateStatus(ctx context.Context, root *ServiceNode) {
	queries := map[string]string{
		"platform-website":   `up{k8s_namespace_name="platform-website"}`,
		"daprd":              `up{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`,
		"signoz-collector":   `up{k8s_namespace_name="signoz",k8s_deployment_name="signoz-otel-collector"}`,
		"cilium-gateway":     `up{k8s_namespace_name="kube-system",k8s_container_name="cilium-envoy"}`,
		"dapr-control-plane": `up{k8s_namespace_name="dapr-system"}`,
	}

	root.Walk(func(s *ServiceNode) {
		q := queries[s.Name]
		if q == "" {
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

func (c *SigNozClient) populateMetrics(ctx context.Context, root *ServiceNode) {
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
			query:  `go_goroutines{k8s_namespace_name="platform-website",k8s_container_name="platform-website"}`,
			format: func(v float64) string { return fmt.Sprintf("%.0f", v) },
		},
		{
			name: "platform-website", label: "heap", unit: "MiB",
			query:  `go_memstats_alloc_bytes{k8s_namespace_name="platform-website",k8s_container_name="platform-website"}`,
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
		valStr := results[0].Metric["value"]
		if len(results[0].Value) >= 2 {
			if f, err := strconv.ParseFloat(fmt.Sprintf("%v", results[0].Value[1]), 64); err == nil {
				valStr = q.format(f)
			}
		}
		node.Metrics = append(node.Metrics, MetricSeries{
			Label: q.label,
			Value: valStr,
			Unit:  q.unit,
		})
	}
}

func (c *SigNozClient) populateStats(ctx context.Context) StatsStrip {
	stats := StatsStrip{NodeCount: 2}

	// Uptime from process_start_time_seconds
	results, err := c.queryInstant(ctx,
		`process_start_time_seconds{k8s_namespace_name="platform-website",k8s_container_name="platform-website"}`)
	if err == nil && len(results) > 0 && len(results[0].Value) >= 2 {
		if ts, err := strconv.ParseFloat(fmt.Sprintf("%v", results[0].Value[1]), 64); err == nil {
			uptime := time.Since(time.Unix(int64(ts), 0))
			stats.Uptime = formatDuration(uptime)
		}
	}

	// Go version from go_info
	results, err = c.queryInstant(ctx,
		`go_info{k8s_namespace_name="platform-website",k8s_container_name="platform-website"}`)
	if err == nil && len(results) > 0 {
		if v, ok := results[0].Metric["version"]; ok {
			stats.GoVersion = v
		}
	}

	return stats
}

func (c *SigNozClient) populateHealth(ctx context.Context, root *ServiceNode) []HealthCheck {
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

// queryInstant runs a PromQL instant query.
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
