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

// SigNozClient fetches live data from the SigNoz Prometheus-compatible API
// and ClickHouse for real node-level metrics.
type SigNozClient struct {
	baseURL        string
	apiKey         string
	clickhouseURL  string
	clickhouseAuth string
	namespace      string
	httpClient     *http.Client
	cached         LiveData
	cachedAt       time.Time
	cacheTTL       time.Duration
}

// nodeMetrics holds real node-level metrics from ClickHouse.
type nodeMetrics struct {
	cpuPct  float64 // k8s.node.cpu.usage
	ramMB   float64 // k8s.node.memory.working_set (MB)
	ioWait  float64 // system.cpu.time state=wait (% of total)
	loadAvg float64 // system.cpu.load_average.5m
	uptime  float64 // k8s.node.uptime (seconds)
}

func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	ns := getNamespace()
	log.Printf("SigNoz client: namespace=%s", ns)
	return &SigNozClient{
		baseURL:        strings.TrimRight(baseURL, "/"),
		apiKey:         apiKey,
		clickhouseURL:  os.Getenv("CLICKHOUSE_URL"),
		clickhouseAuth: os.Getenv("CLICKHOUSE_AUTH"),
		namespace:      ns,
		httpClient:     &http.Client{Timeout: 5 * time.Second},
		cacheTTL:       30 * time.Second,
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

// queryClickHouse runs a SQL query against the ClickHouse HTTP interface.
func (c *SigNozClient) queryClickHouse(ctx context.Context, sql string, fn func(row map[string]interface{})) {
	if c.clickhouseURL == "" || c.clickhouseAuth == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST",
		c.clickhouseURL,
		strings.NewReader(sql+" FORMAT JSONEachRow"))
	if err != nil {
		return
	}
	req.SetBasicAuth("admin", c.clickhouseAuth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	dec := json.NewDecoder(resp.Body)
	for {
		var row map[string]interface{}
		if err := dec.Decode(&row); err != nil {
			break
		}
		fn(row)
	}
}

// sparklinePoints returns SVG polyline points for a sparkline.
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
	r := max - min
	if r == 0 {
		r = 1
	}
	var pts []string
	for i, v := range values {
		x := float64(i) * float64(width) / float64(len(values)-1)
		y := float64(height) - ((v - min) / r * float64(height))
		pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	return strings.Join(pts, " ")
}

// sortServices sorts by category order then by name.
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

// promResponse and promResultItem model the SigNoz /api/v1/query JSON response.
type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string           `json:"resultType"`
		Result     []promResultItem `json:"result"`
	} `json:"data"`
}

type promResultItem struct {
	Metric map[string]string `json:"metric"`
	Value  interface{}       `json:"value,omitempty"`
	Values interface{}       `json:"values,omitempty"`
}

func getOrCreateNode(m map[string]*nodeMetrics, key string) *nodeMetrics {
	if _, ok := m[key]; !ok {
		m[key] = &nodeMetrics{}
	}
	return m[key]
}

// Fetch returns live services discovered from SigNoz metrics.
// Uses ClickHouse samples_v4_agg_5m (pre-aggregated 5-minute rollups)
// instead of raw samples_v4 to avoid scanning millions of rows.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Services) > 0 {
		return c.cached, nil
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	now := time.Now()

	// --- Service discovery via Prometheus v1 API ---

	discoveredMap := map[string]bool{}
	scrapeOKMap := map[string]bool{}
	uptimeMap := map[string]string{}
	hostMap := map[string]string{}
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
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				hostMap[key] = node
			}
		}
	}

	// --- Build service list ---

	seen := map[string]bool{}
	var services []Service

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
			Host:      hostMap[key],
		}
		if scrapeOKMap[key] {
			svc.Status = "healthy"
		} else {
			svc.Status = "running"
		}
		if t, ok := uptimeMap[key]; ok {
			if parsed, err := time.Parse(time.RFC3339, t); err == nil {
				svc.Uptime = FormatUptime(now.Sub(parsed))
			}
		}
		services = append(services, svc)
	}

	sortServices(services)

	// --- Pod-level metrics from ClickHouse agg table ---
	// Uses samples_v4_agg_5m (pre-aggregated 5-min rollups).
	// Single query per metric: returns both latest value AND 1h of bucketed
	// data, avoiding the previous pattern of separate snapshot + sparkline queries.

	type podData struct {
		latest  float64
		history []float64
	}

	fetchPodMetric := func(metricName string, expr func(float64) float64) map[string]*podData {
		result := map[string]*podData{}
		c.queryClickHouse(fetchCtx,
			"SELECT "+
				"JSONExtractString(t.labels, 'k8s.namespace.name') as ns, "+
				"JSONExtractString(t.labels, 'k8s.pod.name') as pod, "+
				"argMax(a.last, a.unix_milli) as latest, "+
				"groupArray(a.last) as vals "+
				"FROM signoz_metrics.samples_v4_agg_5m a "+
				"INNER JOIN signoz_metrics.time_series_v4 t ON a.fingerprint = t.fingerprint "+
				"WHERE t.metric_name = '"+metricName+"' "+
				"AND a.unix_milli >= toUnixTimestamp(now() - INTERVAL 1 HOUR) * 1000 "+
				"GROUP BY ns, pod "+
				"ORDER BY ns, pod",
			func(row map[string]interface{}) {
				if ns, ok := chString(row, "ns"); ok {
					if pod, ok := chString(row, "pod"); ok {
						latest, _ := chFloat(row, "latest")
						rawVals := parseFloatArray(row["vals"])
						vals := make([]float64, len(rawVals))
						for i, v := range rawVals {
							vals[i] = expr(v)
						}
						result[ns+"/"+pod] = &podData{latest: expr(latest), history: vals}
					}
				}
			})
		return result
	}

	// CPU: raw cores (no transform)
	podCPU := fetchPodMetric("k8s.pod.cpu.usage", func(v float64) float64 { return v })
	// RAM: bytes to MB
	podRAM := fetchPodMetric("k8s.pod.memory.working_set", func(v float64) float64 { return v / 1024 / 1024 })
	// Disk: bytes to MB
	podDisk := fetchPodMetric("k8s.pod.filesystem.usage", func(v float64) float64 { return v / 1024 / 1024 })

	// Network: cumulative counter, use sum/max/min from agg to compute rate
	// For each pod: latest rate = (max-min)/duration, per-bucket rate = (max-min)/300
	podNet := map[string]*podData{}
	c.queryClickHouse(fetchCtx,
		"SELECT "+
			"JSONExtractString(t.labels, 'k8s.namespace.name') as ns, "+
			"JSONExtractString(t.labels, 'k8s.pod.name') as pod, "+
			"(max(a.max) - min(a.min)) / 600 as avg_rate, "+
			"groupArray((a.max - a.min) / 300) as bucket_rates "+
			"FROM signoz_metrics.samples_v4_agg_5m a "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON a.fingerprint = t.fingerprint "+
			"WHERE t.metric_name = 'k8s.pod.network.io' "+
			"AND a.unix_milli >= toUnixTimestamp(now() - INTERVAL 1 HOUR) * 1000 "+
			"GROUP BY ns, pod "+
			"HAVING avg_rate > 0 "+
			"ORDER BY ns, pod",
		func(row map[string]interface{}) {
			if ns, ok := chString(row, "ns"); ok {
				if pod, ok := chString(row, "pod"); ok {
					latest, _ := chFloat(row, "avg_rate")
					rawRates := parseFloatArray(row["bucket_rates"])
					vals := make([]float64, len(rawRates))
					for i, v := range rawRates {
						vals[i] = v / 1024 // bytes/s to KB/s
					}
					podNet[ns+"/"+pod] = &podData{latest: latest / 1024, history: vals}
				}
			}
		})

	// Match pod data to services
	for i := range services {
		svcKey := services[i].Namespace + "/" + services[i].Name
		var cpuMax, ramMax, netSum, diskSum float64
		var netCount, diskCount int

		for podKey, d := range podCPU {
			if strings.HasPrefix(podKey, svcKey+"-") || podKey == svcKey {
				if d.latest > cpuMax {
					cpuMax = d.latest
				}
				if len(d.history) >= 2 {
					services[i].CPUHist = sparklinePoints(d.history, 48, 16)
				}
			}
		}
		for podKey, d := range podRAM {
			if strings.HasPrefix(podKey, svcKey+"-") || podKey == svcKey {
				if d.latest > ramMax {
					ramMax = d.latest
				}
				if len(d.history) >= 2 {
					services[i].RAMHist = sparklinePoints(d.history, 48, 16)
				}
			}
		}
		for podKey, d := range podNet {
			if strings.HasPrefix(podKey, svcKey+"-") || podKey == svcKey {
				netSum += d.latest
				netCount++
				if len(d.history) >= 2 {
					services[i].NetHist = sparklinePoints(d.history, 48, 16)
				}
			}
		}
		for podKey, d := range podDisk {
			if strings.HasPrefix(podKey, svcKey+"-") || podKey == svcKey {
				diskSum += d.latest
				diskCount++
				if len(d.history) >= 2 {
					services[i].DiskHist = sparklinePoints(d.history, 48, 16)
				}
			}
		}

		if cpuMax > 0 {
			services[i].CPU = math.Round(cpuMax*100) / 100
		}
		if ramMax > 0 {
			services[i].RAM = math.Round(ramMax*10) / 10
		}
		if netCount > 0 {
			services[i].NetKB = math.Round(netSum*10) / 10
		}
		if diskCount > 0 {
			services[i].DiskMB = math.Round(diskSum*10) / 10
		}
	}

	// --- Host metrics from ClickHouse ---
	// Node metrics use raw samples_v4 since they're few and simple.

	nodeMap := map[string]*nodeMetrics{}

	c.queryClickHouse(fetchCtx,
		"SELECT argMax(s.value, s.unix_milli) as v, "+
			"JSONExtractString(t.labels, 'k8s.node.name') as node "+
			"FROM signoz_metrics.samples_v4 s "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON s.fingerprint = t.fingerprint "+
			"WHERE s.metric_name = 'k8s.node.cpu.usage' "+
			"AND s.unix_milli >= toUnixTimestamp(now() - INTERVAL 5 MINUTE) * 1000 "+
			"GROUP BY node ORDER BY node",
		func(row map[string]interface{}) {
			if node, ok := chString(row, "node"); ok {
				if v, ok := chFloat(row, "v"); ok {
					getOrCreateNode(nodeMap, node).cpuPct = v
				}
			}
		})

	c.queryClickHouse(fetchCtx,
		"SELECT argMax(s.value, s.unix_milli) as v, "+
			"JSONExtractString(t.labels, 'k8s.node.name') as node "+
			"FROM signoz_metrics.samples_v4 s "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON s.fingerprint = t.fingerprint "+
			"WHERE s.metric_name = 'k8s.node.memory.working_set' "+
			"AND s.unix_milli >= toUnixTimestamp(now() - INTERVAL 5 MINUTE) * 1000 "+
			"GROUP BY node ORDER BY node",
		func(row map[string]interface{}) {
			if node, ok := chString(row, "node"); ok {
				if v, ok := chFloat(row, "v"); ok {
					getOrCreateNode(nodeMap, node).ramMB = v / 1024 / 1024
				}
			}
		})

	c.queryClickHouse(fetchCtx,
		"SELECT argMax(s.value, s.unix_milli) as v, "+
			"JSONExtractString(t.labels, 'k8s.node.name') as node "+
			"FROM signoz_metrics.samples_v4 s "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON s.fingerprint = t.fingerprint "+
			"WHERE s.metric_name = 'system.cpu.load_average.5m' "+
			"AND s.unix_milli >= toUnixTimestamp(now() - INTERVAL 5 MINUTE) * 1000 "+
			"GROUP BY node ORDER BY node",
		func(row map[string]interface{}) {
			if node, ok := chString(row, "node"); ok {
				if v, ok := chFloat(row, "v"); ok {
					getOrCreateNode(nodeMap, node).loadAvg = v
				}
			}
		})

	c.queryClickHouse(fetchCtx,
		"WITH rates AS ("+
			"SELECT JSONExtractString(t.labels, 'k8s.node.name') as node, "+
			"JSONExtractString(t.labels, 'state') as state, "+
			"(max(s.value)-min(s.value))/(max(s.unix_milli)-min(s.unix_milli)) as rate "+
			"FROM signoz_metrics.samples_v4 s "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON s.fingerprint = t.fingerprint "+
			"WHERE s.metric_name = 'system.cpu.time' "+
			"AND s.unix_milli >= toUnixTimestamp(now() - INTERVAL 10 MINUTE) * 1000 "+
			"GROUP BY node, state) "+
			"SELECT node, sumIf(rate, state = 'wait') / sum(rate) * 100 as iowait_pct "+
			"FROM rates GROUP BY node ORDER BY node",
		func(row map[string]interface{}) {
			if node, ok := chString(row, "node"); ok {
				if v, ok := chFloat(row, "iowait_pct"); ok {
					getOrCreateNode(nodeMap, node).ioWait = v
				}
			}
		})

	c.queryClickHouse(fetchCtx,
		"SELECT argMax(s.value, s.unix_milli) as v, "+
			"JSONExtractString(t.labels, 'k8s.node.name') as node "+
			"FROM signoz_metrics.samples_v4 s "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON s.fingerprint = t.fingerprint "+
			"WHERE s.metric_name = 'k8s.node.uptime' "+
			"AND s.unix_milli >= toUnixTimestamp(now() - INTERVAL 5 MINUTE) * 1000 "+
			"GROUP BY node ORDER BY node",
		func(row map[string]interface{}) {
			if node, ok := chString(row, "node"); ok {
				if v, ok := chFloat(row, "v"); ok {
					getOrCreateNode(nodeMap, node).uptime = v
				}
			}
		})

	// --- Build host list ---

	nodeCountMap := map[string]int{}
	if results, err := c.queryInstant(fetchCtx, "count(up) by (k8s_node_name)"); err == nil {
		for _, r := range results {
			if node := metricLabel(r, "k8s_node_name"); node != "" {
				if v, err := parseFloat(metricValue(r)); err == nil {
					nodeCountMap[node] = int(v)
				}
			}
		}
	}

	var hosts []Host
	hostDefs := []struct {
		name, label, detail string
	}{
		{"talosoci-control-plane-legal-poodle", "OCI Cloud", "ARM64 \u00b7 Ampere A1"},
		{"talosedge-genmachiche-flowing-bluejay", "Edge Node", "AMD64 \u00b7 Intel NUC"},
	}
	for _, h := range hostDefs {
		host := Host{
			Name:   h.name,
			Label:  h.label,
			Detail: h.detail,
		}
		if m, ok := nodeMap[h.name]; ok {
			host.CPU = math.Round(m.cpuPct*100) / 100
			host.RAM = math.Round(m.ramMB*10) / 10
			host.IOWait = math.Round(m.ioWait*10) / 10
			host.LoadAvg = math.Round(m.loadAvg*100) / 100
			if m.uptime > 0 {
				host.Uptime = FormatUptime(time.Duration(m.uptime) * time.Second)
			}
		}
		if count, ok := nodeCountMap[h.name]; ok {
			host.SvcCount = count
		}
		hosts = append(hosts, host)
	}

	c.cached = LiveData{
		Hosts:      hosts,
		Services:   services,
		HasMetrics: true,
		Timestamp:  now.Unix(),
	}
	c.cachedAt = now
	return c.cached, nil
}

// --- ClickHouse JSON helpers ---

func chFloat(row map[string]interface{}, key string) (float64, bool) {
	v, ok := row[key]
	if !ok {
		return 0, false
	}
	f, err := parseFloat(fmt.Sprintf("%v", v))
	return f, err == nil
}

func chString(row map[string]interface{}, key string) (string, bool) {
	v, ok := row[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// parseFloatArray extracts []float64 from a ClickHouse groupArray result.
func parseFloatArray(v interface{}) []float64 {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]float64, 0, len(arr))
	for _, item := range arr {
		if f, err := parseFloat(fmt.Sprintf("%v", item)); err == nil {
			result = append(result, f)
		}
	}
	return result
}

// --- Prometheus v1 HTTP helpers ---

func (c *SigNozClient) queryInstant(ctx context.Context, query string) ([]map[string]interface{}, error) {
	u := c.baseURL + "/api/v1/query?query=" + url.QueryEscape(query)
	return c.doQuery(ctx, u)
}

func (c *SigNozClient) queryRange(ctx context.Context, query string, start, end time.Time, step int64) ([]map[string]interface{}, error) {
	u := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%d",
		c.baseURL, url.QueryEscape(query), start.Unix(), end.Unix(), step)
	return c.doQuery(ctx, u)
}

func (c *SigNozClient) doQuery(ctx context.Context, u string) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
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

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string                   `json:"resultType"`
			Result     []map[string]interface{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data.Result, nil
}

// --- Metric parsing helpers ---

func serviceKey(r map[string]interface{}) string {
	ns := metricLabel(r, "k8s_namespace_name")
	deploy := metricLabel(r, "k8s.deployment.name")
	if deploy == "" {
		deploy = metricLabel(r, "k8s.statefulset.name")
	}
	if ns == "" || deploy == "" {
		return ""
	}
	return ns + "/" + deploy
}

func metricLabel(r map[string]interface{}, key string) string {
	labels, ok := r["metric"].(map[string]interface{})
	if !ok {
		return ""
	}
	v, ok := labels[key].(string)
	return v
}

func metricValue(r map[string]interface{}) string {
	v, ok := r["value"]
	if !ok {
		return ""
	}
	if val, ok := v.([]interface{}); ok && len(val) == 2 {
		return fmt.Sprintf("%v", val[1])
	}
	return ""
}

func extractValues(r map[string]interface{}) []float64 {
	raw, ok := r["values"]
	if !ok {
		return nil
	}
	pairs, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	values := make([]float64, 0, len(pairs))
	for _, p := range pairs {
		pair, ok := p.([]interface{})
		if !ok || len(pair) < 2 {
			continue
		}
		if f, err := parseFloat(fmt.Sprintf("%v", pair[1])); err == nil {
			values = append(values, f)
		}
	}
	return values
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
