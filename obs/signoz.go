package obs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// SigNozClient fetches live platform data from SigNoz.
// A single v3 query_range call returns all metrics + sparkline history.
type SigNozClient struct {
	baseURL   string
	apiKey    string
	namespace string
	http      *http.Client
	cached    LiveData
	cachedAt  time.Time
	cacheTTL  time.Duration
}

func NewSigNozClient(baseURL, apiKey string) *SigNozClient {
	ns := getNamespace()
	log.Printf("SigNoz client: namespace=%s", ns)
	return &SigNozClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		namespace: ns,
		http:      &http.Client{Timeout: 10 * time.Second},
		cacheTTL:  30 * time.Second,
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

// ── v3 API types ──

type v3Request struct {
	Start          int64       `json:"start"`
	End            int64       `json:"end"`
	Step           int64       `json:"step"`
	CompositeQuery v3Composite `json:"compositeQuery"`
}

type v3Composite struct {
	PanelType   string                 `json:"panelType"`
	QueryType   string                 `json:"queryType"`
	PromQueries map[string]v3PromQuery `json:"promQueries"`
}

type v3PromQuery struct {
	Query    string `json:"query"`
	Disabled bool   `json:"disabled"`
}

type v3Response struct {
	Status string     `json:"status"`
	Error  string     `json:"error,omitempty"`
	Data   v3DataRoot `json:"data"`
}

type v3DataRoot struct {
	Result []v3Result `json:"result"`
}

type v3Result struct {
	QueryName string     `json:"queryName"`
	Series    []v3Series `json:"series"`
}

type v3Series struct {
	Labels map[string]string `json:"labels"`
	Values []v3Point         `json:"values"`
}

type v3Point struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"`
}

// ── Fetch ──

func (c *SigNozClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Services) > 0 {
		return c.cached, nil
	}

	fctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	startMs := now.Add(-2 * time.Hour).UnixMilli()
	endMs := now.UnixMilli()

	// Single v3 query_range call for all metrics.
	// Uses 2h window for resilience against SigNoz data gaps after restarts.
	// step=300 over 2h gives ~24 data points per series (good sparklines).
	// No `up` query: service discovery uses the CPU metric directly,
	// which has k8s.deployment.name, k8s.node.name, k8s.pod.start_time labels.
	resp, err := c.queryV3(fctx, v3Request{
		Start: startMs,
		End:   endMs,
		Step:  300,
		CompositeQuery: v3Composite{
			PanelType: "graph",
			QueryType: "promql",
			PromQueries: map[string]v3PromQuery{
				"cpu":      {Query: `{__name__="k8s.pod.cpu.usage"}`, Disabled: false},
				"ram":      {Query: `{__name__="k8s.pod.memory.working_set"}`, Disabled: false},
				"disk":     {Query: `{__name__="k8s.pod.filesystem.usage"}`, Disabled: false},
				"net":      {Query: `rate({__name__="k8s.pod.network.io",direction="receive"}[5m])`, Disabled: false},
				"nodeCpu":  {Query: `{__name__="k8s.node.cpu.usage"}`, Disabled: false},
				"nodeRam":  {Query: `{__name__="k8s.node.memory.working_set"}`, Disabled: false},
				"nodeLoad": {Query: `{__name__="system.cpu.load_average.5m"}`, Disabled: false},
				"nodeUp":   {Query: `{__name__="k8s.node.uptime"}`, Disabled: false},
			},
		},
	})
	if err != nil {
		// Return static hosts on failure
		return LiveData{
			Hosts:     staticHosts(),
			Timestamp: now.Unix(),
		}, nil
	}

	// Parse results by query name
	results := map[string][]v3Series{}
	for _, r := range resp.Data.Result {
		results[r.QueryName] = r.Series
	}

	services := buildServices(results["cpu"], results, now)
	hosts := buildHosts(results)

	c.cached = LiveData{
		Hosts:      hosts,
		Services:   services,
		HasMetrics: true,
		Timestamp:  now.Unix(),
	}
	c.cachedAt = now
	return c.cached, nil
}

// ── Service builder ──

func buildServices(cpuSeries []v3Series, allResults map[string][]v3Series, now time.Time) []Service {
	type svcInfo struct {
		ns, deploy, host, uptime string
	}
	svcMap := map[string]*svcInfo{}

	// Discover services from CPU metric labels.
	// k8s.pod.cpu.usage has: k8s.deployment.name, k8s.statefulset.name,
	// k8s.daemonset.name, k8s.node.name, k8s.pod.start_time.
	for _, s := range cpuSeries {
		ns := labelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		deploy := labelStr(s.Labels, "k8s.deployment.name")
		if deploy == "" {
			deploy = labelStr(s.Labels, "k8s.statefulset.name")
		}
		if deploy == "" {
			deploy = labelStr(s.Labels, "k8s.daemonset.name")
		}
		if ns == "" || deploy == "" {
			continue
		}
		key := ns + "/" + deploy
		if _, exists := svcMap[key]; exists {
			continue
		}
		info := &svcInfo{ns: ns, deploy: deploy}
		svcMap[key] = info
		if t := labelStr(s.Labels, "k8s.pod.start_time"); t != "" {
			info.uptime = t
		}
		if node := labelStr(s.Labels, "k8s_node_name", "k8s.node.name"); node != "" {
			info.host = node
		}
	}

	// Build pod metric lookups
	podCPU := latestByPod(allResults["cpu"])
	podRAM := latestByPod(allResults["ram"])
	podDisk := latestByPod(allResults["disk"])
	podNet := latestByPod(allResults["net"])

	sparkCPU := sparkByPod(allResults["cpu"])
	sparkRAM := sparkByPod(allResults["ram"])
	sparkDisk := sparkByPod(allResults["disk"])
	sparkNet := sparkByPod(allResults["net"])

	services := make([]Service, 0, len(svcMap))
	for _, info := range svcMap {
		svc := Service{
			Name:      info.deploy,
			Namespace: info.ns,
			Category:  CategoryForNamespace(info.ns),
			Host:      info.host,
		}
		// All discovered services are running (they have metric data)
		svc.Status = "running"
		if info.uptime != "" {
			if parsed, err := time.Parse(time.RFC3339, info.uptime); err == nil {
				svc.Uptime = FormatUptime(now.Sub(parsed))
			}
		}

		// Match pod metrics to deployment (prefix matching)
		var cpuMax, ramMax, netSum, diskSum float64
		for podKey, v := range podCPU {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				if v > cpuMax {
					cpuMax = v
				}
			}
		}
		for podKey, v := range podRAM {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				if v/1024/1024 > ramMax {
					ramMax = v / 1024 / 1024
				}
			}
		}
		for podKey, v := range podNet {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				netSum += v / 1024 // bytes/s to KB/s
			}
		}
		for podKey, v := range podDisk {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				diskSum += v / 1024 / 1024
			}
		}

		svc.CPU = math.Round(cpuMax*1000) / 1000
		svc.RAM = math.Round(ramMax*10) / 10
		svc.NetKB = math.Round(netSum*10) / 10
		svc.DiskMB = math.Round(diskSum*10) / 10

		// Sparklines: first matching pod's series
		for podKey, pts := range sparkCPU {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				svc.CPUHist = pts
				break
			}
		}
		for podKey, pts := range sparkRAM {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				svc.RAMHist = pts
				break
			}
		}
		for podKey, pts := range sparkNet {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				svc.NetHist = pts
				break
			}
		}
		for podKey, pts := range sparkDisk {
			if matchPod(podKey, svc.Namespace, svc.Name) {
				svc.DiskHist = pts
				break
			}
		}

		services = append(services, svc)
	}

	sortServices(services)
	return services
}

// latestByPod returns map[podKey]latestValue from a v3 series result.
// podKey = namespace/pod-name.
func latestByPod(series []v3Series) map[string]float64 {
	m := map[string]float64{}
	for _, s := range series {
		ns := labelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		pod := labelStr(s.Labels, "k8s.pod.name")
		if ns == "" || pod == "" || len(s.Values) == 0 {
			continue
		}
		last := s.Values[len(s.Values)-1].Value
		if v, err := parseFloat(last); err == nil {
			m[ns+"/"+pod] = v
		}
	}
	return m
}

// sparkByPod returns map[podKey]svgPolyline from a v3 series result.
func sparkByPod(series []v3Series) map[string]string {
	m := map[string]string{}
	for _, s := range series {
		ns := labelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		pod := labelStr(s.Labels, "k8s.pod.name")
		if ns == "" || pod == "" || len(s.Values) < 2 {
			continue
		}
		vals := make([]float64, len(s.Values))
		for i, p := range s.Values {
			vals[i], _ = parseFloat(p.Value)
		}
		m[ns+"/"+pod] = sparklinePoints(vals, 48, 16)
	}
	return m
}

// ── Host builder ──

func buildHosts(results map[string][]v3Series) []Host {
	nodeCPU := latestByNode(results["nodeCpu"])
	nodeRAM := latestByNode(results["nodeRam"])
	nodeLoad := latestByNode(results["nodeLoad"])
	nodeUp := latestByNode(results["nodeUp"])

	// Count services per node from CPU results
	nodeCount := map[string]int{}
	for _, s := range results["cpu"] {
		if node := labelStr(s.Labels, "k8s_node_name", "k8s.node.name"); node != "" {
			nodeCount[node]++
		}
	}

	defs := []struct{ name, label, detail string }{
		{"talosoci-control-plane-legal-poodle", "OCI Cloud", "ARM64 \u00b7 Ampere A1"},
		{"talosedge-genmachiche-flowing-bluejay", "Edge Node", "AMD64 \u00b7 Intel NUC"},
	}
	hosts := make([]Host, 0, len(defs))
	for _, d := range defs {
		h := Host{Name: d.name, Label: d.label, Detail: d.detail}
		if v, ok := nodeCPU[d.name]; ok {
			h.CPU = math.Round(v*100) / 100
		}
		if v, ok := nodeRAM[d.name]; ok {
			h.RAM = math.Round(v/1024/1024*10) / 10
		}
		if v, ok := nodeLoad[d.name]; ok {
			h.LoadAvg = math.Round(v*100) / 100
		}
		if v, ok := nodeUp[d.name]; ok {
			h.Uptime = FormatUptime(time.Duration(v) * time.Second)
		}
		if n, ok := nodeCount[d.name]; ok {
			h.SvcCount = n
		}
		hosts = append(hosts, h)
	}
	return hosts
}

func latestByNode(series []v3Series) map[string]float64 {
	m := map[string]float64{}
	for _, s := range series {
		node := s.Labels["k8s.node.name"]
		if node == "" || len(s.Values) == 0 {
			continue
		}
		last := s.Values[len(s.Values)-1].Value
		if v, err := parseFloat(last); err == nil {
			m[node] = v
		}
	}
	return m
}

func staticHosts() []Host {
	return []Host{
		{Name: "talosoci-control-plane-legal-poodle", Label: "OCI Cloud", Detail: "ARM64 \u00b7 Ampere A1"},
		{Name: "talosedge-genmachiche-flowing-bluejay", Label: "Edge Node", Detail: "AMD64 \u00b7 Intel NUC"},
	}
}

// labelStr returns the first non-empty value from the given label keys.
func labelStr(labels map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := labels[k]; ok && v != "" {
			return v
		}
	}
	return ""
}

// matchPod checks if a pod key (ns/pod-name-hash) belongs to a deployment (ns/deploy).
func matchPod(podKey, ns, deploy string) bool {
	prefix := ns + "/" + deploy + "-"
	return strings.HasPrefix(podKey, prefix) || podKey == ns+"/"+deploy
}

// ── SigNoz v3 API call ──

func (c *SigNozClient) queryV3(ctx context.Context, req v3Request) (*v3Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/v3/query_range", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("SIGNOZ-API-KEY", c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var v3Resp v3Response
	if err := json.Unmarshal(respBody, &v3Resp); err != nil {
		return nil, err
	}
	if v3Resp.Status == "error" {
		return nil, fmt.Errorf("signoz: %s", v3Resp.Error)
	}
	return &v3Resp, nil
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
