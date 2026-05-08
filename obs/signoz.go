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

// SigNozClient fetches live platform data from SigNoz.
// It runs 3 queries per fetch cycle (every 30s):
//  1. PromQL: "up" → discovers services, maps to hosts
//  2. ClickHouse: pod metrics → CPU, RAM, Disk + sparklines
//  3. ClickHouse: node metrics → host CPU, RAM, Load, Uptime
//
// Network rate is computed from rolling buffer deltas (no extra query).
type SigNozClient struct {
	baseURL   string
	apiKey    string
	chURL     string
	chAuth    string
	namespace string
	http      *http.Client
	cached    LiveData
	cachedAt  time.Time
	cacheTTL  time.Duration
	netBuf    map[string][]netEntry // rolling buffer for network rate
}

// netEntry stores a timestamped cumulative network byte count.
type netEntry struct {
	ts    time.Time
	bytes float64
}

const sparklineLen = 12

func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	ns := getNamespace()
	log.Printf("SigNoz client: namespace=%s", ns)
	return &SigNozClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		chURL:     os.Getenv("CLICKHOUSE_URL"),
		chAuth:    os.Getenv("CLICKHOUSE_AUTH"),
		namespace: ns,
		http:      &http.Client{Timeout: 10 * time.Second},
		cacheTTL:  30 * time.Second,
		netBuf:    map[string][]netEntry{},
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

// Fetch returns live platform data. Uses 30s cache.
func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Services) > 0 {
		return c.cached, nil
	}

	fctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	now := time.Now()

	// ── Query 1: Service discovery via PromQL "up" ──
	services, nodeCountMap := c.discoverServices(fctx, now)

	// ── Query 2: Pod metrics (CPU, RAM, Disk + sparklines) ──
	c.fetchPodMetrics(fctx, services, now)

	// ── Query 3: Node metrics (CPU, RAM, Load, Uptime) ──
	nodeMap := c.fetchNodeMetrics(fctx)

	// ── Build result ──
	hosts := buildHosts(nodeMap, nodeCountMap)

	c.cached = LiveData{
		Hosts:      hosts,
		Services:   services,
		HasMetrics: true,
		Timestamp:  now.Unix(),
	}
	c.cachedAt = now
	return c.cached, nil
}

// discoverServices queries PromQL "up" to find all monitored deployments.
func (c *SigNozClient) discoverServices(ctx context.Context, now time.Time) ([]Service, map[string]int) {
	type svcInfo struct {
		ns, deploy, host, uptime string
		healthy                  bool
	}
	svcMap := map[string]*svcInfo{}

	results, err := c.queryProm(ctx, "up")
	if err != nil {
		return nil, nil
	}
	for _, r := range results {
		ns := label(r, "k8s_namespace_name")
		deploy := label(r, "k8s.deployment.name")
		if deploy == "" {
			deploy = label(r, "k8s.statefulset.name")
		}
		if ns == "" || deploy == "" {
			continue
		}
		key := ns + "/" + deploy
		info, ok := svcMap[key]
		if !ok {
			info = &svcInfo{ns: ns, deploy: deploy}
			svcMap[key] = info
		}
		if value(r) == "1" {
			info.healthy = true
		}
		if t := label(r, "k8s.pod.start_time"); t != "" {
			info.uptime = t
		}
		if node := label(r, "k8s_node_name"); node != "" {
			info.host = node
		}
	}

	services := make([]Service, 0, len(svcMap))
	for key, info := range svcMap {
		svc := Service{
			Name:      info.deploy,
			Namespace: info.ns,
			Category:  CategoryForNamespace(info.ns),
			Host:      info.host,
		}
		if info.healthy {
			svc.Status = "healthy"
		} else {
			svc.Status = "running"
		}
		if info.uptime != "" {
			if parsed, err := time.Parse(time.RFC3339, info.uptime); err == nil {
				svc.Uptime = FormatUptime(now.Sub(parsed))
			}
		}
		_ = key
		services = append(services, svc)
	}
	sortServices(services)

	// Per-node service counts
	nodeCountMap := map[string]int{}
	results2, err := c.queryProm(ctx, "count(up) by (k8s_node_name)")
	if err == nil {
		for _, r := range results2 {
			if node := label(r, "k8s_node_name"); node != "" {
				if v, err := parseFloat(value(r)); err == nil {
					nodeCountMap[node] = int(v)
				}
			}
		}
	}

	return services, nodeCountMap
}

// fetchPodMetrics runs ONE ClickHouse query to get CPU, RAM, Disk
// for all pods, with 1h of sparkline history from the agg table.
// Network rate is computed from the rolling buffer.
func (c *SigNozClient) fetchPodMetrics(ctx context.Context, services []Service, now time.Time) {
	// Single query: latest value + 1h sparkline history for CPU, RAM, Disk
	type podRow struct {
		metric string
		ns     string
		pod    string
		latest float64
		vals   []float64
	}
	rows := []podRow{}

	c.queryCH(ctx,
		"SELECT a.metric_name, "+
			"JSONExtractString(t.labels, 'k8s.namespace.name') as ns, "+
			"JSONExtractString(t.labels, 'k8s.pod.name') as pod, "+
			"argMax(a.last, a.unix_milli) as latest, "+
			"groupArray(a.last) as vals "+
			"FROM signoz_metrics.samples_v4_agg_5m a "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON a.fingerprint = t.fingerprint "+
			"WHERE a.metric_name IN ('k8s.pod.cpu.usage', 'k8s.pod.memory.working_set', 'k8s.pod.filesystem.usage', 'k8s.pod.network.io') "+
			"AND a.unix_milli >= toUnixTimestamp(now() - INTERVAL 1 HOUR) * 1000 "+
			"GROUP BY a.metric_name, ns, pod",
		func(row map[string]interface{}) {
			r := podRow{}
			if s, ok := chStr(row, "metric_name"); ok {
				r.metric = s
			}
			if s, ok := chStr(row, "ns"); ok {
				r.ns = s
			}
			if s, ok := chStr(row, "pod"); ok {
				r.pod = s
			}
			if v, ok := chNum(row, "latest"); ok {
				r.latest = v
			}
			r.vals = chArr(row["vals"])
			rows = append(rows, r)
		})

	// Build per-pod lookup
	type podMetrics struct {
		cpuLatest, ramLatest, diskLatest, netLatest float64
		cpuHist, ramHist, diskHist                  []float64
	}
	pods := map[string]*podMetrics{}
	for _, r := range rows {
		key := r.ns + "/" + r.pod
		m, ok := pods[key]
		if !ok {
			m = &podMetrics{}
			pods[key] = m
		}
		switch r.metric {
		case "k8s.pod.cpu.usage":
			m.cpuLatest = r.latest
			m.cpuHist = r.vals
		case "k8s.pod.memory.working_set":
			m.ramLatest = r.latest / 1024 / 1024
			m.ramHist = make([]float64, len(r.vals))
			for i, v := range r.vals {
				m.ramHist[i] = v / 1024 / 1024
			}
		case "k8s.pod.filesystem.usage":
			m.diskLatest = r.latest / 1024 / 1024
			m.diskHist = make([]float64, len(r.vals))
			for i, v := range r.vals {
				m.diskHist[i] = v / 1024 / 1024
			}
		case "k8s.pod.network.io":
			// Cumulative counter: push to rolling buffer for rate computation
			m.netLatest = r.latest
			// Skip sparkline from raw counter — we compute rate from buffer
		}
	}

	// Push network bytes to rolling buffer
	for key, m := range pods {
		if m.netLatest > 0 {
			c.netBuf[key] = append(c.netBuf[key], netEntry{ts: now, bytes: m.netLatest})
			if len(c.netBuf[key]) > sparklineLen {
				c.netBuf[key] = c.netBuf[key][len(c.netBuf[key])-sparklineLen:]
			}
		}
	}

	// Compute network rates from rolling buffer
	netRates := map[string]float64{}     // latest rate in KB/s
	netSparklines := map[string]string{} // sparkline SVG
	for key, buf := range c.netBuf {
		if len(buf) >= 2 {
			last := buf[len(buf)-1]
			prev := buf[len(buf)-2]
			dt := last.ts.Sub(prev.ts).Seconds()
			if dt > 0 && last.bytes >= prev.bytes {
				netRates[key] = (last.bytes - prev.bytes) / dt / 1024
			}
			// Build sparkline from rate between consecutive entries
			if len(buf) >= 2 {
				rates := []float64{}
				for i := 1; i < len(buf); i++ {
					dt := buf[i].ts.Sub(buf[i-1].ts).Seconds()
					if dt > 0 && buf[i].bytes >= buf[i-1].bytes {
						rates = append(rates, (buf[i].bytes-buf[i-1].bytes)/dt/1024)
					}
				}
				netSparklines[key] = sparklinePoints(rates, 48, 16)
			}
		}
	}

	// Match pod metrics to services
	for i := range services {
		svcKey := services[i].Namespace + "/" + services[i].Name
		var cpuMax, ramMax, diskSum float64
		var diskCount int

		for podKey, m := range pods {
			if !matchPod(podKey, svcKey) {
				continue
			}
			if m.cpuLatest > cpuMax {
				cpuMax = m.cpuLatest
			}
			if m.ramLatest > ramMax {
				ramMax = m.ramLatest
			}
			diskSum += m.diskLatest
			if m.diskLatest > 0 {
				diskCount++
			}
			// Sparklines: use first matching pod's history
			if len(m.cpuHist) >= 2 && services[i].CPUHist == "" {
				services[i].CPUHist = sparklinePoints(m.cpuHist, 48, 16)
			}
			if len(m.ramHist) >= 2 && services[i].RAMHist == "" {
				services[i].RAMHist = sparklinePoints(m.ramHist, 48, 16)
			}
			if len(m.diskHist) >= 2 && services[i].DiskHist == "" {
				services[i].DiskHist = sparklinePoints(m.diskHist, 48, 16)
			}
		}

		// Network rate from rolling buffer (per-service, not per-pod)
		var netSum float64
		for podKey := range pods {
			if matchPod(podKey, svcKey) {
				if rate, ok := netRates[podKey]; ok {
					netSum += rate
				}
				if sl, ok := netSparklines[podKey]; ok && services[i].NetHist == "" {
					services[i].NetHist = sl
				}
			}
		}

		if cpuMax > 0 {
			services[i].CPU = math.Round(cpuMax*100) / 100
		}
		if ramMax > 0 {
			services[i].RAM = math.Round(ramMax*10) / 10
		}
		if netSum > 0 {
			services[i].NetKB = math.Round(netSum*10) / 10
		}
		if diskSum > 0 {
			services[i].DiskMB = math.Round(diskSum*10) / 10
		}
	}
}

// fetchNodeMetrics runs ONE ClickHouse query for all node metrics.
func (c *SigNozClient) fetchNodeMetrics(ctx context.Context) map[string]*nodeMetrics {
	nodeMap := map[string]*nodeMetrics{}

	c.queryCH(ctx,
		"SELECT a.metric_name, "+
			"JSONExtractString(t.labels, 'k8s.node.name') as node, "+
			"argMax(a.last, a.unix_milli) as val "+
			"FROM signoz_metrics.samples_v4_agg_5m a "+
			"INNER JOIN signoz_metrics.time_series_v4 t ON a.fingerprint = t.fingerprint "+
			"WHERE a.metric_name IN ('k8s.node.cpu.usage', 'k8s.node.memory.working_set', 'system.cpu.load_average.5m', 'k8s.node.uptime') "+
			"AND a.unix_milli >= toUnixTimestamp(now() - INTERVAL 10 MINUTE) * 1000 "+
			"GROUP BY a.metric_name, node",
		func(row map[string]interface{}) {
			node, ok := chStr(row, "node")
			if !ok {
				return
			}
			val, ok := chNum(row, "val")
			if !ok {
				return
			}
			metric, _ := chStr(row, "metric_name")
			nm := getNode(nodeMap, node)
			switch metric {
			case "k8s.node.cpu.usage":
				nm.cpuPct = val
			case "k8s.node.memory.working_set":
				nm.ramMB = val / 1024 / 1024
			case "system.cpu.load_average.5m":
				nm.loadAvg = val
			case "k8s.node.uptime":
				nm.uptime = val
			}
		})

	return nodeMap
}

type nodeMetrics struct {
	cpuPct, ramMB, loadAvg, uptime float64
}

func getNode(m map[string]*nodeMetrics, k string) *nodeMetrics {
	if _, ok := m[k]; !ok {
		m[k] = &nodeMetrics{}
	}
	return m[k]
}

// buildHosts creates the host list from node metrics.
func buildHosts(nodeMap map[string]*nodeMetrics, nodeCountMap map[string]int) []Host {
	defs := []struct{ name, label, detail string }{
		{"talosoci-control-plane-legal-poodle", "OCI Cloud", "ARM64 \u00b7 Ampere A1"},
		{"talosedge-genmachiche-flowing-bluejay", "Edge Node", "AMD64 \u00b7 Intel NUC"},
	}
	hosts := make([]Host, 0, len(defs))
	for _, d := range defs {
		h := Host{Name: d.name, Label: d.label, Detail: d.detail}
		if m, ok := nodeMap[d.name]; ok {
			h.CPU = math.Round(m.cpuPct*100) / 100
			h.RAM = math.Round(m.ramMB*10) / 10
			h.LoadAvg = math.Round(m.loadAvg*100) / 100
			if m.uptime > 0 {
				h.Uptime = FormatUptime(time.Duration(m.uptime) * time.Second)
			}
		}
		if n, ok := nodeCountMap[d.name]; ok {
			h.SvcCount = n
		}
		hosts = append(hosts, h)
	}
	return hosts
}

// matchPod checks if a pod key (ns/pod-name-hash) belongs to a service (ns/deploy).
func matchPod(podKey, svcKey string) bool {
	return strings.HasPrefix(podKey, svcKey+"-") || podKey == svcKey
}

// ── ClickHouse ──

func (c *SigNozClient) queryCH(ctx context.Context, sql string, fn func(map[string]interface{})) {
	if c.chURL == "" || c.chAuth == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.chURL,
		strings.NewReader(sql+" FORMAT JSONEachRow"))
	if err != nil {
		return
	}
	req.SetBasicAuth("admin", c.chAuth)
	resp, err := c.http.Do(req)
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

func chStr(row map[string]interface{}, key string) (string, bool) {
	v, ok := row[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func chNum(row map[string]interface{}, key string) (float64, bool) {
	v, ok := row[key]
	if !ok {
		return 0, false
	}
	f, err := parseFloat(fmt.Sprintf("%v", v))
	return f, err == nil
}

func chArr(v interface{}) []float64 {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]float64, 0, len(arr))
	for _, item := range arr {
		if f, err := parseFloat(fmt.Sprintf("%v", item)); err == nil {
			out = append(out, f)
		}
	}
	return out
}

// ── Prometheus v1 API ──

func (c *SigNozClient) queryProm(ctx context.Context, query string) ([]map[string]interface{}, error) {
	u := c.baseURL + "/api/v1/query?query=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("SIGNOZ-API-KEY", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data struct {
			Result []map[string]interface{} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data.Result, nil
}

func label(r map[string]interface{}, key string) string {
	labels, ok := r["metric"].(map[string]interface{})
	if !ok {
		return ""
	}
	s, _ := labels[key].(string)
	return s
}

func value(r map[string]interface{}) string {
	v, ok := r["value"]
	if !ok {
		return ""
	}
	if arr, ok := v.([]interface{}); ok && len(arr) == 2 {
		return fmt.Sprintf("%v", arr[1])
	}
	return ""
}

// ── Sparkline ──

func sparklinePoints(values []float64, w, h int) string {
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
	pts := make([]string, len(values))
	for i, v := range values {
		x := float64(i) * float64(w) / float64(len(values)-1)
		y := float64(h) - ((v - min) / r * float64(h))
		pts[i] = fmt.Sprintf("%.1f,%.1f", x, y)
	}
	return strings.Join(pts, " ")
}

// ── Sorting ──

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

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
