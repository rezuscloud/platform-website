package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SigNozClient fetches live data from the SigNoz Prometheus-compatible API.
type SigNozClient struct {
	baseURL    string
	apiKey     string
	namespace  string
	httpClient *http.Client
	cached     LiveData
	cachedAt   time.Time
	cacheTTL   time.Duration
}

func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	ns := getNamespace()
	log.Printf("SigNoz client: namespace=%s", ns)
	return &SigNozClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		namespace:  ns,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		cacheTTL:   30 * time.Second,
	}
}

func NewSigNozClientFromEnv() *SigNozClient {
	u := os.Getenv("SIGNOZ_URL")
	k := os.Getenv("SIGNOZ_API_KEY")
	if u == "" || k == "" {
		return nil
	}
	return NewSigNozClient(u, k)
}

func getNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "platform-website"
}

// Fetch returns the platform dashboard with live health data.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Categories) > 0 {
		return c.cached, nil
	}

	cats := PlatformCategories()
	monitored := MonitoredNamespaces()

	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Query health for each monitored namespace
	nsHealth := map[string]bool{}
	for ns := range monitored {
		results, err := c.queryInstant(fetchCtx, fmt.Sprintf(`up{k8s_namespace_name="%s"}`, ns))
		if err == nil && len(results) > 0 {
			nsHealth[ns] = true
		}
	}

	// Apply health status
	for i := range cats {
		for j := range cats[i].Services {
			svc := &cats[i].Services[j]
			if svc.Namespace == "" {
				svc.Status = "running"
			} else if nsHealth[svc.Namespace] {
				svc.Status = "healthy"
			} else if monitored[svc.Namespace] {
				svc.Status = "unknown"
			} else {
				svc.Status = "unmonitored"
			}
		}
	}

	// Fetch compact metrics for platform-website
	c.populateMetrics(fetchCtx, &cats)

	stats := c.populateStats(fetchCtx)

	c.cached = LiveData{
		Categories: cats,
		Stats:      stats,
		HasMetrics: true,
	}
	c.cachedAt = time.Now()
	return c.cached, nil
}

func (c *SigNozClient) populateMetrics(ctx context.Context, cats *[]PlatformCategory) {
	// Goroutines for the runtime column
	goroutines, err := c.queryInstant(ctx,
		fmt.Sprintf(`go_goroutines{k8s_namespace_name="%s"}`, c.namespace))
	if err == nil && len(goroutines) > 0 && len(goroutines[0].Value) >= 2 {
		if v, err := parseFloat(goroutines[0].Value[1]); err == nil {
			for i := range *cats {
				if (*cats)[i].ID == "runtime" {
					for j := range (*cats)[i].Services {
						if (*cats)[i].Services[j].Name == "platform-website" {
							(*cats)[i].Services[j].Metric = fmt.Sprintf("%.0f goroutines", v)
						}
						if (*cats)[i].Services[j].Name == "daprd" {
							(*cats)[i].Services[j].Metric = "2 components"
						}
					}
				}
			}
		}
	}
}

func (c *SigNozClient) populateStats(ctx context.Context) StatsStrip {
	stats := StatsStrip{NodeCount: 2}

	results, err := c.queryInstant(ctx,
		fmt.Sprintf(`process_start_time_seconds{k8s_namespace_name="%s"}`, c.namespace))
	if err == nil && len(results) > 0 && len(results[0].Value) >= 2 {
		if ts, err := parseFloat(results[0].Value[1]); err == nil {
			stats.Uptime = formatDuration(time.Since(time.Unix(int64(ts), 0)))
		}
	}

	results, err = c.queryInstant(ctx,
		fmt.Sprintf(`go_info{k8s_namespace_name="%s"}`, c.namespace))
	if err == nil && len(results) > 0 {
		if v, ok := results[0].Metric["version"]; ok {
			stats.GoVersion = v
		}
	}

	return stats
}

func parseFloat(v interface{}) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
	return f, err
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	if h >= 24 {
		return fmt.Sprintf("%dd", h/24)
	}
	if h >= 1 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
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
