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

// Fetch returns the platform-website service map with live metrics.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	data := PlatformTopology()

	// Check which services are up
	if err := c.populateStatus(ctx, &data); err != nil {
		return data, fmt.Errorf("status: %w", err)
	}

	// Fetch app metrics for sparklines
	if err := c.populateMetrics(ctx, &data); err != nil {
		return data, fmt.Errorf("metrics: %w", err)
	}

	return data, nil
}

// populateStatus checks the "up" metric for each known service.
func (c *SigNozClient) populateStatus(ctx context.Context, data *LiveData) error {
	// Map our service names to PromQL label matchers
	queries := map[string]string{
		"platform-website":   `up{k8s_namespace_name="platform-website"}`,
		"daprd":              `up{k8s_namespace_name="platform-website",k8s_container_name="daprd"}`,
		"signoz-collector":   `up{k8s_namespace_name="signoz",k8s_deployment_name="signoz-otel-collector"}`,
		"cilium-gateway":     `up{k8s_namespace_name="kube-system",k8s_container_name="cilium-envoy"}`,
		"dapr-control-plane": `up{k8s_namespace_name="dapr-system"}`,
	}

	for i := range data.Services {
		q := queries[data.Services[i].Name]
		if q == "" {
			continue
		}
		results, err := c.queryInstant(ctx, q)
		if err != nil {
			data.Services[i].Status = "unknown"
			continue
		}
		if len(results) > 0 {
			data.Services[i].Status = "healthy"
		} else {
			data.Services[i].Status = "unknown"
		}
	}
	return nil
}

// populateMetrics fetches sparkline data for the platform-website service.
func (c *SigNozClient) populateMetrics(ctx context.Context, data *LiveData) error {
	end := time.Now()
	start := end.Add(-15 * time.Minute)

	type metricQuery struct {
		serviceIdx int
		label      string
		unit       string
		query      string
		transform  func(float64) float64
		format     func(float64) string
	}

	queries := []metricQuery{
		{
			serviceIdx: 1, // platform-website
			label:      "Goroutines",
			unit:       "",
			query:      `go_goroutines{k8s_namespace_name="platform-website",k8s_container_name="platform-website"}`,
			transform:  nil,
			format:     func(v float64) string { return fmt.Sprintf("%.0f", v) },
		},
		{
			serviceIdx: 1,
			label:      "Heap",
			unit:       "MiB",
			query:      `go_memstats_alloc_bytes{k8s_namespace_name="platform-website",k8s_container_name="platform-website"}`,
			transform:  func(v float64) float64 { return v / (1024 * 1024) },
			format:     func(v float64) string { return fmt.Sprintf("%.1f", v) },
		},
		{
			serviceIdx: 2, // daprd
			label:      "Components",
			unit:       "loaded",
			query:      `dapr_runtime_component_loaded{k8s_namespace_name="platform-website"}`,
			transform:  nil,
			format:     func(v float64) string { return fmt.Sprintf("%.0f", v) },
		},
	}

	for _, mq := range queries {
		results, err := c.queryRange(ctx, mq.query, start, end, time.Minute)
		if err != nil || len(results) == 0 {
			continue
		}

		rawPts := extractPoints(results)
		if len(rawPts) == 0 {
			continue
		}

		// Transform points (e.g. bytes → MiB)
		pts := make([]float64, len(rawPts))
		for i, v := range rawPts {
			if mq.transform != nil {
				pts[i] = mq.transform(v)
			} else {
				pts[i] = v
			}
		}

		// Format the last value
		lastVal := pts[len(pts)-1]

		data.Services[mq.serviceIdx].Metrics = append(
			data.Services[mq.serviceIdx].Metrics,
			MetricSeries{
				Label:  mq.label,
				Value:  mq.format(lastVal),
				Unit:   mq.unit,
				Points: pts,
			},
		)
	}

	return nil
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
		return nil, fmt.Errorf("parse: %w", err)
	}
	if pr.Status != "success" {
		return nil, fmt.Errorf("sigNoz: %s", string(body))
	}
	return pr.Data.Result, nil
}

func extractPoints(results []promResultItem) []float64 {
	if len(results) == 0 {
		return nil
	}
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

// Prometheus API types

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
