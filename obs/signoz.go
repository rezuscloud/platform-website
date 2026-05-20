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

// ── v3 API types (private — never escape this file) ──

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

// ── Query definitions ──

// queryName constants tie the PromQL queries to snapshot builder logic.
// Changing these requires updating groupPodSeries / groupNodeSeries.
const (
	queryCPU     = "cpu"
	queryRAM     = "ram"
	queryDisk    = "disk"
	queryNet     = "net"
	queryNodeCPU = "nodeCpu"
	queryNodeRAM = "nodeRam"
	queryNodeUp  = "nodeUp"
)

// podQueries are the PromQL queries that produce per-pod metrics.
var podQueries = map[string]string{
	queryCPU:  `{__name__="k8s.pod.cpu.usage"}`,
	queryRAM:  `{__name__="k8s.pod.memory.working_set"}`,
	queryDisk: `{__name__="k8s.pod.filesystem.usage"}`,
	queryNet:  `rate({__name__="k8s.pod.network.io",direction="receive"}[5m])`,
}

// nodeQueries are the PromQL queries that produce per-node metrics.
var nodeQueries = map[string]string{
	queryNodeCPU: `{__name__="k8s.node.cpu.usage"}`,
	queryNodeRAM: `{__name__="k8s.node.memory.working_set"}`,
	queryNodeUp:  `{__name__="k8s.node.uptime"}`,
}

func allQueries() map[string]v3PromQuery {
	queries := make(map[string]v3PromQuery, len(podQueries)+len(nodeQueries))
	for name, q := range podQueries {
		queries[name] = v3PromQuery{Query: q}
	}
	for name, q := range nodeQueries {
		queries[name] = v3PromQuery{Query: q}
	}
	return queries
}

// ── Snapshot builder ──

// newSnapshot converts a v3 API response into a MetricsSnapshot.
// This is the single place that knows about v3Series label keys and value parsing.
func newSnapshot(resp *v3Response, now time.Time) MetricsSnapshot {
	// Group series by query name.
	results := make(map[string][]v3Series, len(resp.Data.Result))
	for _, r := range resp.Data.Result {
		results[r.QueryName] = r.Series
	}

	// Discover workloads from CPU series (the primary discovery metric).
	type svcInfo struct {
		ns, deploy, host, uptime string
	}
	svcMap := make(map[string]*svcInfo)
	for _, s := range results[queryCPU] {
		key := WorkloadKey(s.Labels)
		if key == "" || svcMap[key] != nil {
			continue
		}
		idx := strings.Index(key, "/")
		ns, deploy := key[:idx], key[idx+1:]
		info := &svcInfo{
			ns:     ns,
			deploy: deploy,
			host:   LabelStr(s.Labels, "k8s_node_name", "k8s.node.name"),
		}
		if t := LabelStr(s.Labels, "k8s.pod.start_time"); t != "" {
			info.uptime = t
		}
		svcMap[key] = info
	}

	// Extract per-deployment latest values and sparkline history.
	latest := make(map[string]map[string]float64, len(podQueries))
	spark := make(map[string]map[string][]float64, len(podQueries))
	for qName := range podQueries {
		latest[qName] = latestByWorkload(results[qName])
		spark[qName] = sparkByWorkload(results[qName])
	}

	// Build WorkloadMetrics.
	workloads := make(map[string]WorkloadMetrics, len(svcMap))
	for key, info := range svcMap {
		wm := WorkloadMetrics{
			Namespace: info.ns,
			Name:      info.deploy,
			Host:      info.host,
			Uptime:    info.uptime,
		}
		for qName, all := range latest {
			if v, ok := all[key]; ok {
				switch qName {
				case queryCPU:
					if v <= 100 {
						wm.CPU = math.Round(v*1000) / 1000
					}
				case queryRAM:
					wm.RAM = v // raw bytes — caller converts to MB
				case queryNet:
					wm.Net = v
				case queryDisk:
					wm.Disk = v
				}
			}
		}
		for qName, all := range spark {
			if pts, ok := all[key]; ok {
				switch qName {
				case queryCPU:
					wm.CPUHist = pts
				case queryRAM:
					wm.RAMHist = pts
				case queryNet:
					wm.NetHist = pts
				case queryDisk:
					wm.DiskHist = pts
				}
			}
		}
		workloads[key] = wm
	}

	// Build NodeMetrics.
	nodeLatest := make(map[string]map[string]float64, len(nodeQueries))
	for qName := range nodeQueries {
		nodeLatest[qName] = latestByNode(results[qName])
	}
	nodeNames := make(map[string]bool)
	for _, m := range nodeLatest {
		for name := range m {
			nodeNames[name] = true
		}
	}
	nodes := make(map[string]NodeMetrics, len(nodeNames))
	for name := range nodeNames {
		nm := NodeMetrics{}
		if v, ok := nodeLatest[queryNodeCPU][name]; ok {
			nm.CPU = v
		}
		if v, ok := nodeLatest[queryNodeRAM][name]; ok {
			nm.RAM = v
		}
		if v, ok := nodeLatest[queryNodeUp][name]; ok {
			nm.Uptime = v
		}
		nodes[name] = nm
	}

	// Count services per node from CPU series.
	nodeSvcCounts := make(map[string]int)
	for _, s := range results[queryCPU] {
		if node := LabelStr(s.Labels, "k8s_node_name", "k8s.node.name"); node != "" {
			nodeSvcCounts[node]++
		}
	}

	return MetricsSnapshot{
		Workloads:     workloads,
		Nodes:         nodes,
		NodeSvcCounts: nodeSvcCounts,
	}
}

// latestByWorkload returns map[workloadKey]latestValue across pods.
// Takes max across pods of the same deployment.
func latestByWorkload(series []v3Series) map[string]float64 {
	m := make(map[string]float64, len(series))
	for _, s := range series {
		key := WorkloadKey(s.Labels)
		if key == "" || len(s.Values) == 0 {
			continue
		}
		v, err := parseFloat(s.Values[len(s.Values)-1].Value)
		if err != nil {
			continue
		}
		if existing, ok := m[key]; ok && v <= existing {
			continue
		}
		m[key] = v
	}
	return m
}

// sparkByWorkload returns map[workloadKey]sparklineValues.
// First pod wins for duplicate workloads; values >100 clamped to 0.
func sparkByWorkload(series []v3Series) map[string][]float64 {
	m := make(map[string][]float64, len(series))
	for _, s := range series {
		key := WorkloadKey(s.Labels)
		if key == "" || len(s.Values) < 1 {
			continue
		}
		if _, exists := m[key]; exists {
			continue
		}
		vals := make([]float64, len(s.Values))
		for i, p := range s.Values {
			vals[i], _ = parseFloat(p.Value)
		}
		for i, v := range vals {
			if v > 100 {
				vals[i] = 0
			}
		}
		if len(vals) == 1 {
			vals = []float64{vals[0], vals[0]}
		}
		m[key] = vals
	}
	return m
}

// latestByNode returns map[nodeName]latestValue.
func latestByNode(series []v3Series) map[string]float64 {
	m := make(map[string]float64, len(series))
	for _, s := range series {
		node := s.Labels["k8s.node.name"]
		if node == "" || len(s.Values) == 0 {
			continue
		}
		if v, err := parseFloat(s.Values[len(s.Values)-1].Value); err == nil {
			m[node] = v
		}
	}
	return m
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
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

	resp, err := c.queryV3(fctx, v3Request{
		Start: startMs,
		End:   endMs,
		Step:  300,
		CompositeQuery: v3Composite{
			PanelType:   "graph",
			QueryType:   "promql",
			PromQueries: allQueries(),
		},
	})
	if err != nil {
		return LiveData{
			Hosts:     []Host{},
			Timestamp: now.Unix(),
		}, nil
	}

	snap := newSnapshot(resp, now)
	services := BuildServices(snap, now)
	hosts := BuildHosts(snap)

	c.cached = LiveData{
		Hosts:      hosts,
		Services:   services,
		HasMetrics: true,
		Timestamp:  now.Unix(),
	}
	c.cachedAt = now
	return c.cached, nil
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
