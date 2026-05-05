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
	"time"
)

// SigNozClient fetches live infrastructure data from the SigNoz Prometheus-compatible API.
// Uses SIGNOZ_URL (e.g. "https://signoz.rezus.cloud") and SIGNOZ_API_KEY.
type SigNozClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSigNozClient creates a client from the given URL and API key.
func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	return &SigNozClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// NewSigNozClientFromEnv creates a client from SIGNOZ_URL and SIGNOZ_API_KEY env vars.
// Returns nil if either is empty.
func NewSigNozClientFromEnv() *SigNozClient {
	u := os.Getenv("SIGNOZ_URL")
	k := os.Getenv("SIGNOZ_API_KEY")
	if u == "" || k == "" {
		return nil
	}
	return NewSigNozClient(u, k)
}

// Fetch queries SigNoz for the current cluster topology and metrics.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	nodes, err := c.fetchNodes(ctx)
	if err != nil {
		return LiveData{}, fmt.Errorf("fetch nodes: %w", err)
	}

	metrics, err := c.fetchMetrics(ctx)
	if err != nil {
		return LiveData{Nodes: nodes}, fmt.Errorf("fetch metrics: %w", err)
	}

	return LiveData{Nodes: nodes, Metrics: metrics}, nil
}

func (c *SigNozClient) fetchNodes(ctx context.Context) ([]Node, error) {
	// Query "up" to get all pods with their node and namespace labels.
	// This gives us the full cluster topology: which pods are on which nodes.
	results, err := c.queryInstant(ctx, `up`)
	if err != nil {
		return nil, err
	}

	// Group pods by node
	type nodeInfo struct {
		pods    []Pod
		seenPod map[string]bool
	}
	nodeMap := make(map[string]*nodeInfo)
	var nodeOrder []string

	for _, r := range results {
		metric := r.Metric
		nodeName := metric["k8s_node_name"]
		podName := metric["k8s_pod_name"]
		namespace := metric["k8s_namespace_name"]

		if nodeName == "" || podName == "" {
			continue
		}

		if _, ok := nodeMap[nodeName]; !ok {
			nodeMap[nodeName] = &nodeInfo{seenPod: make(map[string]bool)}
			nodeOrder = append(nodeOrder, nodeName)
		}

		ni := nodeMap[nodeName]
		if !ni.seenPod[podName] {
			ni.seenPod[podName] = true
			ni.pods = append(ni.pods, Pod{
				Name:      podName,
				Namespace: namespace,
				Status:    "Running",
			})
		}
	}

	// Build Node slice
	nodes := make([]Node, 0, len(nodeOrder))
	for _, name := range nodeOrder {
		ni := nodeMap[name]
		nodes = append(nodes, Node{
			Name:   name,
			Tier:   tierFromNodeName(name),
			Status: "Ready",
			Pods:   ni.pods,
		})
	}

	return nodes, nil
}

func (c *SigNozClient) fetchMetrics(ctx context.Context) ([]MetricSeries, error) {
	end := time.Now()
	start := end.Add(-15 * time.Minute)

	var metrics []MetricSeries

	// Goroutines sparkline for platform-website
	goroutineValues, err := c.queryRange(ctx,
		`go_goroutines{k8s_namespace_name="platform-website"}`,
		start, end, time.Minute,
	)
	if err == nil && len(goroutineValues) > 0 {
		pts := extractPoints(goroutineValues)
		last := lastValue(goroutineValues)
		metrics = append(metrics, MetricSeries{
			Label:  "Goroutines",
			Value:  last,
			Unit:   "",
			Points: pts,
		})
	}

	// Memory sparkline for platform-website
	memValues, err := c.queryRange(ctx,
		`go_memstats_alloc_bytes{k8s_namespace_name="platform-website"}`,
		start, end, time.Minute,
	)
	if err == nil && len(memValues) > 0 {
		pts := extractPoints(memValues)
		// Convert bytes to MiB
		for i, v := range pts {
			pts[i] = v / (1024 * 1024)
		}
		lastRaw := lastValueRaw(memValues)
		lastMiB := lastRaw / (1024 * 1024)
		metrics = append(metrics, MetricSeries{
			Label:  "Heap Alloc",
			Value:  fmt.Sprintf("%.1f", lastMiB),
			Unit:   "MiB",
			Points: pts,
		})
	}

	// Dapr sidecar metrics
	daprValues, err := c.queryRange(ctx,
		`dapr_runtime_component_loaded{k8s_namespace_name="platform-website"}`,
		start, end, time.Minute,
	)
	if err == nil && len(daprValues) > 0 {
		pts := extractPoints(daprValues)
		last := lastValue(daprValues)
		metrics = append(metrics, MetricSeries{
			Label:  "Components",
			Value:  last,
			Unit:   "loaded",
			Points: pts,
		})
	}

	return metrics, nil
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
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if pr.Status != "success" {
		return nil, fmt.Errorf("sigNoz error: %s", string(body))
	}

	return pr.Data.Result, nil
}

// queryRange runs a PromQL range query for sparkline data.
func (c *SigNozClient) queryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]promResultItem, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/query_range"
	q := u.Query()
	q.Set("query", query)
	q.Set("start", strconv.FormatInt(start.Unix(), 10))
	q.Set("end", strconv.FormatInt(end.Unix(), 10))
	q.Set("step", strconv.FormatInt(int64(step.Seconds()), 10))
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

	var pr promRangeResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if pr.Status != "success" {
		return nil, fmt.Errorf("sigNoz error: %s", string(body))
	}

	return pr.Data.Result, nil
}

// tierFromNodeName determines the infrastructure tier from the node name.
func tierFromNodeName(name string) string {
	if contains(name, "talosoci") || contains(name, "oci") {
		return "oci-cloud"
	}
	return "edge"
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// extractPoints gets the float64 values from a range query result.
func extractPoints(results []promResultItem) []float64 {
	if len(results) == 0 {
		return nil
	}
	// Take the first series
	vals := results[0].Values
	points := make([]float64, len(vals))
	for i, v := range vals {
		if len(v) >= 2 {
			if f, err := strconv.ParseFloat(fmt.Sprintf("%v", v[1]), 64); err == nil {
				points[i] = f
			}
		}
	}
	return points
}

func lastValue(results []promResultItem) string {
	if len(results) == 0 {
		return "?"
	}
	v := results[0].Value
	if len(v) >= 2 {
		return fmt.Sprintf("%v", v[1])
	}
	// Fallback to range values
	vals := results[0].Values
	if len(vals) > 0 && len(vals[len(vals)-1]) >= 2 {
		return fmt.Sprintf("%v", vals[len(vals)-1][1])
	}
	return "?"
}

func lastValueRaw(results []promResultItem) float64 {
	if len(results) == 0 {
		return 0
	}
	v := results[0].Value
	if len(v) >= 2 {
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", v[1]), 64)
		return f
	}
	vals := results[0].Values
	if len(vals) > 0 && len(vals[len(vals)-1]) >= 2 {
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", vals[len(vals)-1][1]), 64)
		return f
	}
	return 0
}

// Prometheus API response types

type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string           `json:"resultType"`
		Result     []promResultItem `json:"result"`
	} `json:"data"`
}

type promRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string           `json:"resultType"`
		Result     []promResultItem `json:"result"`
	} `json:"data"`
}

type promResultItem struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value,omitempty"`
	Values [][]interface{}   `json:"values,omitempty"`
}
