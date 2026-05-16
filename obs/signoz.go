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
	"sort"
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
	startMs := now.Add(-6 * time.Hour).UnixMilli()
	endMs := now.UnixMilli()

	// Single v3 query_range call for all metrics.
	// Uses 6h window for resilience against SigNoz collection gaps.
	// step=300 over 6h gives ~72 data points per series (good sparklines).
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
				"cpu":     {Query: `{__name__="k8s.pod.cpu.usage"}`, Disabled: false},
				"ram":     {Query: `{__name__="k8s.pod.memory.working_set"}`, Disabled: false},
				"disk":    {Query: `{__name__="k8s.pod.filesystem.usage"}`, Disabled: false},
				"net":     {Query: `rate({__name__="k8s.pod.network.io",direction="receive"}[5m])`, Disabled: false},
				"nodeCpu": {Query: `{__name__="k8s.node.cpu.usage"}`, Disabled: false},
				"nodeRam": {Query: `{__name__="k8s.node.memory.working_set"}`, Disabled: false},
				"nodeUp":  {Query: `{__name__="k8s.node.uptime"}`, Disabled: false},
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
		Hosts:         hosts,
		Services:      services,
		HasMetrics:    true,
		Timestamp:     now.Unix(),
		SelfNamespace: c.namespace,
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

	// Build deployment metric lookups using k8s.deployment.name label
	deployCPU := latestByDeployment(allResults["cpu"])
	deployRAM := latestByDeployment(allResults["ram"])
	deployDisk := latestByDeployment(allResults["disk"])
	deployNet := latestByDeployment(allResults["net"])

	sparkCPU := sparkByDeployment(allResults["cpu"])
	sparkRAM := sparkByDeployment(allResults["ram"])
	sparkDisk := sparkByDeployment(allResults["disk"])
	sparkNet := sparkByDeployment(allResults["net"])

	services := make([]Service, 0, len(svcMap))
	for _, info := range svcMap {
		key := info.ns + "/" + info.deploy
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

		if v, ok := deployCPU[key]; ok && v <= 100 {
			svc.CPU = math.Round(v*1000) / 1000
		}
		if v, ok := deployRAM[key]; ok {
			svc.RAM = math.Round(v/1024/1024*10) / 10
		}
		if v, ok := deployNet[key]; ok {
			svc.NetKB = math.Round(v/1024*10) / 10
		}
		if v, ok := deployDisk[key]; ok {
			svc.DiskMB = math.Round(v/1024/1024*10) / 10
		}

		if pts, ok := sparkCPU[key]; ok {
			svc.CPUHist = pts
		}
		if pts, ok := sparkRAM[key]; ok {
			svc.RAMHist = pts
		}
		if pts, ok := sparkNet[key]; ok {
			svc.NetHist = pts
		}
		if pts, ok := sparkDisk[key]; ok {
			svc.DiskHist = pts
		}

		services = append(services, svc)
	}

	// Filter out stale services: pods from deleted namespaces still have
	// residual SigNoz data in the query window but no active metrics.
	active := make([]Service, 0, len(services))
	for _, svc := range services {
		if svc.CPU > 0 || svc.RAM > 0 || svc.NetKB > 0 || svc.DiskMB > 0 {
			active = append(active, svc)
		}
	}

	sortServices(active)
	return active
}

// latestByDeployment returns map[ns/deploy]latestValue using k8s.deployment.name label.
// Takes max across pods of the same deployment. Clamps corrupted values (>100 cores).
func latestByDeployment(series []v3Series) map[string]float64 {
	m := map[string]float64{}
	for _, s := range series {
		ns := labelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		deploy := labelStr(s.Labels, "k8s.deployment.name")
		if deploy == "" {
			deploy = labelStr(s.Labels, "k8s.statefulset.name")
		}
		if deploy == "" {
			deploy = labelStr(s.Labels, "k8s.daemonset.name")
		}
		if ns == "" || deploy == "" || len(s.Values) == 0 {
			continue
		}
		key := ns + "/" + deploy
		last := s.Values[len(s.Values)-1].Value
		v, err := parseFloat(last)
		if err != nil {
			continue
		}

		if existing, ok := m[key]; ok {
			if v > existing {
				m[key] = v
			}
		} else {
			m[key] = v
		}
	}
	return m
}

// sparkByDeployment returns map[ns/deploy]svgPolyline using k8s.deployment.name label.
// Picks the first pod's series for each deployment.
func sparkByDeployment(series []v3Series) map[string]string {
	m := map[string]string{}
	for _, s := range series {
		ns := labelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		deploy := labelStr(s.Labels, "k8s.deployment.name")
		if deploy == "" {
			deploy = labelStr(s.Labels, "k8s.statefulset.name")
		}
		if deploy == "" {
			deploy = labelStr(s.Labels, "k8s.daemonset.name")
		}
		if ns == "" || deploy == "" || len(s.Values) < 1 {
			continue
		}
		key := ns + "/" + deploy
		if _, exists := m[key]; exists {
			continue // first pod wins
		}
		vals := make([]float64, len(s.Values))
		for i, p := range s.Values {
			vals[i], _ = parseFloat(p.Value)
		}
		// Filter corrupted spikes: clamp to 100 max (CPU cores)
		for i, v := range vals {
			if v > 100 {
				vals[i] = 0
				vals[i] = 0
			}
		}
		if len(vals) == 1 {
			vals = []float64{vals[0], vals[0]}
		}
		m[key] = sparklinePoints(vals, 48, 16)
	}
	return m
}

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

// ── Host builder ──

func buildHosts(results map[string][]v3Series) []Host {
	nodeCPU := latestByNode(results["nodeCpu"])
	nodeRAM := latestByNode(results["nodeRam"])

	nodeUp := latestByNode(results["nodeUp"])

	// Count services per node from CPU results
	nodeCount := map[string]int{}
	for _, s := range results["cpu"] {
		if node := labelStr(s.Labels, "k8s_node_name", "k8s.node.name"); node != "" {
			nodeCount[node]++
		}
	}

	// Discover node names dynamically from all available sources
	nodeNames := discoverNodeNames(nodeCPU, nodeRAM, nodeUp, nodeCount)

	hosts := make([]Host, 0, len(nodeNames))
	for _, name := range nodeNames {
		h := Host{Name: name}
		if strings.Contains(name, "control-plane") {
			h.Label = "Cloud"
			h.Detail = "Control plane"
		} else {
			h.Label = "Edge"
			h.Detail = "Worker node"
		}
		if v, ok := nodeCPU[name]; ok {
			h.CPU = math.Round(v*100) / 100
		}
		if v, ok := nodeRAM[name]; ok {
			h.RAM = math.Round(v/1024/1024*10) / 10
		}
		if v, ok := nodeUp[name]; ok {
			h.Uptime = FormatUptime(time.Duration(v) * time.Second)
		}
		if n, ok := nodeCount[name]; ok {
			h.SvcCount = n
		}
		hosts = append(hosts, h)
	}
	return hosts
}

// discoverNodeNames returns sorted unique node names from multiple maps.
// Control-plane nodes come first.
func discoverNodeNames(sources ...any) []string {
	seen := map[string]bool{}
	var names []string
	for _, src := range sources {
		switch m := src.(type) {
		case map[string]float64:
			for k := range m {
				if !seen[k] {
					seen[k] = true
					names = append(names, k)
				}
			}
		case map[string]int:
			for k := range m {
				if !seen[k] {
					seen[k] = true
					names = append(names, k)
				}
			}
		}
	}
	sort.Slice(names, func(i, j int) bool {
		ci := strings.Contains(names[i], "control-plane")
		cj := strings.Contains(names[j], "control-plane")
		if ci != cj {
			return ci
		}
		return names[i] < names[j]
	})
	return names
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
	return []Host{}
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
