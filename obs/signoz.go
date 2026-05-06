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

// Fetch returns the platform dashboard with live per-deployment metrics.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Categories) > 0 {
		return c.cached, nil
	}

	cats := PlatformCategories()
	fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Batch query 1: all up targets (health check per deployment + namespace)
	healthMap := map[string]bool{}   // "namespace/deployment" → alive
	nsHealthMap := map[string]bool{} // "namespace" → any alive
	uptimeMap := map[string]string{}
	upResults, err := c.queryInstant(fetchCtx, "up")
	if err == nil {
		for _, r := range upResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if deploy == "" {
				deploy = metricLabel(r, "k8s.statefulset.name")
			}
			alive := metricValue(r) == "1"
			if ns != "" {
				nsHealthMap[ns] = nsHealthMap[ns] || alive
			}
			if ns != "" && deploy != "" {
				healthMap[deploymentKey(ns, deploy)] = alive
			}
			if ns == c.namespace {
				if t := metricLabel(r, "k8s.pod.start_time"); t != "" {
					uptimeMap[ns] = t
				}
			}
		}
	}

	// Batch query 2: goroutines per deployment + namespace
	goroutineMap := map[string]float64{}   // "namespace/deployment" → count
	nsGoroutineMap := map[string]float64{} // "namespace" → first count
	goroutineResults, err := c.queryInstant(fetchCtx, "go_goroutines")
	if err == nil {
		for _, r := range goroutineResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if v, err := parseFloat(metricValue(r)); err == nil {
				if ns != "" {
					if _, exists := nsGoroutineMap[ns]; !exists {
						nsGoroutineMap[ns] = v
					}
				}
				if ns != "" && deploy != "" {
					goroutineMap[deploymentKey(ns, deploy)] = v
				}
			}
		}
	}

	// Batch query 3: memory per deployment + namespace
	memoryMap := map[string]float64{}   // "namespace/deployment" → bytes
	nsMemoryMap := map[string]float64{} // "namespace" → first bytes
	memResults, err := c.queryInstant(fetchCtx, "process_resident_memory_bytes")
	if err == nil {
		for _, r := range memResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if v, err := parseFloat(metricValue(r)); err == nil {
				if ns != "" {
					if _, exists := nsMemoryMap[ns]; !exists {
						nsMemoryMap[ns] = v
					}
				}
				if ns != "" && deploy != "" {
					memoryMap[deploymentKey(ns, deploy)] = v
				}
			}
		}
	}

	// Batch query 4: Dapr component count
	componentCount := 0
	daprResults, err := c.queryInstant(fetchCtx, fmt.Sprintf(`dapr_runtime_component_loaded{k8s_namespace_name="%s"}`, c.namespace))
	if err == nil && len(daprResults) > 0 {
		if v, err := parseFloat(metricValue(daprResults[0])); err == nil {
			componentCount = int(v)
		}
	}

	// Apply live data to services
	now := time.Now()
	for i := range cats {
		for j := range cats[i].Services {
			svc := &cats[i].Services[j]
			svc.UpdatedAt = now.Unix()

			if svc.Namespace == "" {
				// Infrastructure nodes are always running
				svc.Status = "running"
				continue
			}

			// For services without deployment, use namespace-level health
			if svc.Deployment == "" {
				if alive, ok := nsHealthMap[svc.Namespace]; ok && alive {
					svc.Status = "healthy"
				} else if MonitoredNamespaces()[svc.Namespace] {
					svc.Status = "unknown"
				} else {
					svc.Status = "unmonitored"
				}

				// Namespace-level metrics
				if g, ok := nsGoroutineMap[svc.Namespace]; ok {
					svc.Metric = fmt.Sprintf("%.0f goroutines", g)
				}
				if m, ok := nsMemoryMap[svc.Namespace]; ok {
					svc.Memory = fmt.Sprintf("%.0f MB", m/1024/1024)
				}
				continue
			}

			key := deploymentKey(svc.Namespace, svc.Deployment)

			if alive, ok := healthMap[key]; ok {
				if alive {
					svc.Status = "healthy"
				} else {
					svc.Status = "unknown"
				}
			} else {
				svc.Status = "unknown"
			}

			// Goroutines
			if g, ok := goroutineMap[key]; ok {
				svc.Metric = fmt.Sprintf("%.0f goroutines", g)
			}

			// Memory
			if m, ok := memoryMap[key]; ok {
				svc.Memory = fmt.Sprintf("%.0f MB", m/1024/1024)
			}
		}
	}

	// Dapr sidecar: show component count instead of goroutines (shares deployment with app)
	// Clear memory since it shows the app's memory, not daprd's own process
	for i := range cats {
		for j := range cats[i].Services {
			if cats[i].Services[j].Name == "daprd" {
				if componentCount > 0 {
					cats[i].Services[j].Metric = fmt.Sprintf("%d components", componentCount)
				}
				cats[i].Services[j].Memory = ""
			}
		}
	}

	// Stats strip
	stats := StatsStrip{NodeCount: 2}
	if t, ok := uptimeMap[c.namespace]; ok {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			stats.Uptime = formatDuration(now.Sub(parsed))
		}
	}
	goResults, err := c.queryInstant(fetchCtx, fmt.Sprintf(`go_info{k8s_namespace_name="%s"}`, c.namespace))
	if err == nil && len(goResults) > 0 {
		stats.GoVersion = metricLabel(goResults[0], "version")
	}

	c.cached = LiveData{
		Categories: cats,
		Stats:      stats,
		HasMetrics: true,
		Timestamp:  now.Unix(),
	}
	c.cachedAt = now
	return c.cached, nil
}

func parseFloat(v interface{}) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(fmt.Sprintf("%v", v), "%f", &f)
	return f, err
}

func metricLabel(r promResultItem, key string) string {
	if v, ok := r.Metric[key]; ok {
		return v
	}
	return ""
}

func metricValue(r promResultItem) string {
	if len(r.Value) >= 2 {
		return fmt.Sprintf("%v", r.Value[1])
	}
	return ""
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
