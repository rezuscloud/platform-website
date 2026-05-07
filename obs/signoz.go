package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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

// Fetch returns live services discovered from SigNoz metrics.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Services) > 0 {
		return c.cached, nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Query 1: Discovery — which deployments exist as scrape targets
	discoveredMap := map[string]bool{} // "ns/deploy" → seen in up metric
	scrapeOKMap := map[string]bool{}   // "ns/deploy" → up=1 (scrape succeeded)
	uptimeMap := map[string]string{}   // "ns/deploy" → RFC3339 timestamp
	upResults, err := c.queryInstant(fetchCtx, "up")
	if err == nil {
		for _, r := range upResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if deploy == "" {
				deploy = metricLabel(r, "k8s.statefulset.name")
			}
			if ns == "" || deploy == "" {
				continue
			}
			key := ns + "/" + deploy
			discoveredMap[key] = true
			if metricValue(r) == "1" {
				scrapeOKMap[key] = true
			}
			if t := metricLabel(r, "k8s.pod.start_time"); t != "" {
				uptimeMap[key] = t
			}
		}
	}

	// Query 2: CPU% (rate over 5m)
	cpuMap := map[string]float64{} // "ns/deploy" → CPU%
	cpuResults, err := c.queryInstant(fetchCtx, "rate(process_cpu_seconds_total[5m])*100")
	if err == nil {
		for _, r := range cpuResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if deploy == "" {
				deploy = metricLabel(r, "k8s.statefulset.name")
			}
			if ns != "" && deploy != "" {
				if v, err := parseFloat(metricValue(r)); err == nil {
					cpuMap[ns+"/"+deploy] = v
				}
			}
		}
	}

	// Query 3: RAM
	ramMap := map[string]float64{} // "ns/deploy" → bytes
	ramResults, err := c.queryInstant(fetchCtx, "process_resident_memory_bytes")
	if err == nil {
		for _, r := range ramResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if deploy == "" {
				deploy = metricLabel(r, "k8s.statefulset.name")
			}
			if ns != "" && deploy != "" {
				if v, err := parseFloat(metricValue(r)); err == nil {
					ramMap[ns+"/"+deploy] = v
				}
			}
		}
	}

	// Query 4: CPU histogram (12 points, 5min step = 1h lookback)
	now := time.Now()
	cpuHistMap := map[string]string{} // "ns/deploy" → SVG polyline points
	cpuHistResults, err := c.queryRange(fetchCtx,
		"rate(process_cpu_seconds_total[5m])*100",
		now.Add(-60*time.Minute), now, 300)
	if err == nil {
		for _, r := range cpuHistResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if deploy == "" {
				deploy = metricLabel(r, "k8s.statefulset.name")
			}
			if ns == "" || deploy == "" {
				continue
			}
			key := ns + "/" + deploy
			values := extractValues(r)
			if len(values) > 0 {
				cpuHistMap[key] = sparklinePoints(values, 48, 16)
			}
		}
	}

	// Query 5: RAM histogram
	ramHistMap := map[string]string{}
	ramHistResults, err := c.queryRange(fetchCtx,
		"process_resident_memory_bytes",
		now.Add(-60*time.Minute), now, 300)
	if err == nil {
		for _, r := range ramHistResults {
			ns := metricLabel(r, "k8s_namespace_name")
			deploy := metricLabel(r, "k8s.deployment.name")
			if deploy == "" {
				deploy = metricLabel(r, "k8s.statefulset.name")
			}
			if ns == "" || deploy == "" {
				continue
			}
			key := ns + "/" + deploy
			values := extractValues(r)
			if len(values) > 0 {
				ramHistMap[key] = sparklinePoints(values, 48, 16)
			}
		}
	}

	// Query 6: Per-node aggregates (for hosts column)
	nodeCPUMap := map[string]float64{} // node_name → total CPU%
	nodeRAMMap := map[string]float64{} // node_name → total RAM bytes
	nodeCountMap := map[string]int{}   // node_name → service count

	// CPU per node
	nodeCPUResults, err := c.queryInstant(fetchCtx, "sum(rate(process_cpu_seconds_total[5m])*100) by (k8s_node_name)")
	if err == nil {
		for _, r := range nodeCPUResults {
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				if v, err := parseFloat(metricValue(r)); err == nil {
					nodeCPUMap[node] = v
				}
			}
		}
	}

	// RAM per node
	nodeRAMResults, err := c.queryInstant(fetchCtx, "sum(process_resident_memory_bytes) by (k8s_node_name)")
	if err == nil {
		for _, r := range nodeRAMResults {
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				if v, err := parseFloat(metricValue(r)); err == nil {
					nodeRAMMap[node] = v
				}
			}
		}
	}

	// Service count per node
	nodeCountResults, err := c.queryInstant(fetchCtx, "count(up) by (k8s_node_name)")
	if err == nil {
		for _, r := range nodeCountResults {
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				if v, err := parseFloat(metricValue(r)); err == nil {
					nodeCountMap[node] = int(v)
				}
			}
		}
	}

	// Per-node CPU histogram
	nodeCPUHistMap := map[string]string{}
	nodeCPUHistResults, err := c.queryRange(fetchCtx,
		"sum(rate(process_cpu_seconds_total[5m])*100) by (k8s_node_name)",
		now.Add(-60*time.Minute), now, 300)
	if err == nil {
		for _, r := range nodeCPUHistResults {
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				values := extractValues(r)
				if len(values) > 0 {
					nodeCPUHistMap[node] = sparklinePoints(values, 48, 16)
				}
			}
		}
	}

	// Per-node RAM histogram
	nodeRAMHistMap := map[string]string{}
	nodeRAMHistResults, err := c.queryRange(fetchCtx,
		"sum(process_resident_memory_bytes) by (k8s_node_name)",
		now.Add(-60*time.Minute), now, 300)
	if err == nil {
		for _, r := range nodeRAMHistResults {
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				values := extractValues(r)
				if len(values) > 0 {
					nodeRAMHistMap[node] = sparklinePoints(values, 48, 16)
				}
			}
		}
	}

	// Build service list from all discovered deployments
	seen := map[string]bool{}
	var services []Service

	// Collect all unique ns/deploy keys from discovery map
	for key := range discoveredMap {
		parts := strings.SplitN(key, "/", 2)
		if len(parts) != 2 {
			continue
		}
		ns, deploy := parts[0], parts[1]
		if seen[key] {
			continue
		}
		seen[key] = true

		svc := Service{
			Name:      deploy,
			Namespace: ns,
			Category:  CategoryForNamespace(ns),
		}

		// Service discovered as scrape target = pod is running.
		// up=0 means scrape failed (wrong port, timeout, TLS), not service down.
		if scrapeOKMap[key] {
			svc.Status = "healthy" // scrape OK, CPU/RAM metrics available
		} else {
			svc.Status = "running" // discovered, no metrics (scrape config issue)
		}

		if cpu, ok := cpuMap[key]; ok {
			svc.CPU = math.Round(cpu*100) / 100
		}
		if ram, ok := ramMap[key]; ok {
			svc.RAM = math.Round(ram/1024/1024*10) / 10
		}
		if t, ok := uptimeMap[key]; ok {
			if parsed, err := time.Parse(time.RFC3339, t); err == nil {
				svc.Uptime = FormatUptime(now.Sub(parsed))
			}
		}
		if hist, ok := cpuHistMap[key]; ok {
			svc.CPUHist = hist
		}
		if hist, ok := ramHistMap[key]; ok {
			svc.RAMHist = hist
		}

		services = append(services, svc)
	}

	// Sort services: by category order, then by name
	sortServices(services)

	// Prepend static host entries populated with node-level aggregates
	hosts := buildHosts(nodeCPUMap, nodeRAMMap, nodeCountMap, nodeCPUHistMap, nodeRAMHistMap)
	services = append(hosts, services...)

	c.cached = LiveData{
		Services:   services,
		HasMetrics: true,
		Timestamp:  now.Unix(),
	}
	c.cachedAt = now
	return c.cached, nil
}

// buildHosts creates host entries from per-node aggregate metrics.
func buildHosts(
	nodeCPUMap, nodeRAMMap map[string]float64,
	nodeCountMap map[string]int,
	nodeCPUHistMap, nodeRAMHistMap map[string]string,
) []Service {
	// Static host definitions (name, detail)
	hostDefs := []struct{ name, detail string }{
		{"talosoci-control-plane-legal-poodle", "ARM64 · Ampere A1"},
		{"talosedge-genmachiche-flowing-bluejay", "AMD64 · Intel NUC"},
	}

	var hosts []Service
	for _, h := range hostDefs {
		svc := Service{
			Name:     h.name,
			Category: "hosts",
			Status:   "running",
			Detail:   h.detail,
		}

		if cpu, ok := nodeCPUMap[h.name]; ok {
			svc.CPU = math.Round(cpu*100) / 100
		}
		if ram, ok := nodeRAMMap[h.name]; ok {
			svc.RAM = math.Round(ram/1024/1024*10) / 10
		}
		if count, ok := nodeCountMap[h.name]; ok {
			svc.Uptime = fmt.Sprintf("%d svcs", count)
		}
		if hist, ok := nodeCPUHistMap[h.name]; ok {
			svc.CPUHist = hist
		}
		if hist, ok := nodeRAMHistMap[h.name]; ok {
			svc.RAMHist = hist
		}

		hosts = append(hosts, svc)
	}
	return hosts
}

// Returns a string like "0,16 4,12 8,14 ..." fitting width x height.
func sparklinePoints(values []float64, width, height int) string {
	if len(values) < 2 {
		return ""
	}

	min, max := values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Avoid division by zero
	range_ := max - min
	if range_ == 0 {
		range_ = 1
	}

	var points []string
	for i, v := range values {
		x := float64(i) * float64(width) / float64(len(values)-1)
		y := float64(height) - ((v - min) / range_ * float64(height))
		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
	}

	return strings.Join(points, " ")
}

func sortServices(services []Service) {
	catOrder := map[string]int{}
	for i, c := range CategoryOrder {
		catOrder[c] = i
	}
	for i := 0; i < len(services); i++ {
		for j := i + 1; j < len(services); j++ {
			ci := catOrder[services[i].Category]
			cj := catOrder[services[j].Category]
			if ci > cj || (ci == cj && services[i].Name > services[j].Name) {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
}

func extractValues(r promResultItem) []float64 {
	raw, ok := r.Values.([]interface{})
	if !ok {
		return nil
	}
	var values []float64
	for _, v := range raw {
		if pair, ok := v.([]interface{}); ok && len(pair) >= 2 {
			if f, err := parseFloat(fmt.Sprintf("%v", pair[1])); err == nil {
				values = append(values, f)
			}
		}
	}
	return values
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
	return c.doQuery(ctx, u.String())
}

func (c *SigNozClient) queryRange(ctx context.Context, query string, start, end time.Time, step int) ([]promResultItem, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = "/api/v1/query_range"
	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", start.Unix()))
	q.Set("end", fmt.Sprintf("%d", end.Unix()))
	q.Set("step", fmt.Sprintf("%d", step))
	u.RawQuery = q.Encode()
	return c.doQuery(ctx, u.String())
}

func (c *SigNozClient) doQuery(ctx context.Context, fullURL string) ([]promResultItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
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
	Values interface{}       `json:"values,omitempty"`
}
