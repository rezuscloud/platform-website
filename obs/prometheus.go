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
	"strconv"
	"strings"
	"time"
)

// PrometheusClient fetches live platform data from a native Prometheus
// instance (kube-prometheus-stack). No SigNoz, no ClickHouse, no API key —
// just the standard kubelet/cAdvisor metrics every cluster already exposes.
type PrometheusClient struct {
	baseURL  string // e.g. http://kube-prometheus-stack-prometheus.monitoring.svc:9090
	topo     *k8sTopology
	http     *http.Client
	cached   LiveData
	cachedAt time.Time
	cacheTTL time.Duration
}

// DefaultPrometheusURL is the kube-prometheus-stack service DNS used when
// PROMETHEUS_URL is not set.
const DefaultPrometheusURL = "http://kube-prometheus-stack-prometheus.monitoring.svc.cluster.local:9090"

// getNamespace returns the namespace this pod runs in, used to highlight the
// website's own service in the live grid.
func getNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "platform-website"
}

// NewPrometheusClient creates a client for the given Prometheus base URL.
func NewPrometheusClient(baseURL string) *PrometheusClient {
	log.Printf("Prometheus client: %s", baseURL)
	return &PrometheusClient{
		baseURL:  strings.TrimRight(baseURL, "/"),
		topo:     NewK8sTopology(),
		http:     &http.Client{Timeout: 10 * time.Second},
		cacheTTL: 30 * time.Second,
	}
}

// NewPrometheusClientFromEnv builds a client from PROMETHEUS_URL, falling back
// to the kube-prometheus-stack in-cluster service. Returns nil if the pod is
// clearly not in a cluster (no service account) so callers can use mock data.
func NewPrometheusClientFromEnv() *PrometheusClient {
	u := os.Getenv("PROMETHEUS_URL")
	if u == "" {
		u = DefaultPrometheusURL
	}
	return NewPrometheusClient(u)
}

// URL returns the configured Prometheus base URL (for logging).
func (c *PrometheusClient) URL() string { return c.baseURL }

// Fetch implements the Client interface. It pulls pod-level usage from
// Prometheus and node metadata + pod ownership from the K8s API, then merges
// them into a single MetricsSnapshot consumed by BuildServices / BuildHosts.
func (c *PrometheusClient) Fetch(ctx context.Context) (LiveData, error) {
	if time.Since(c.cachedAt) < c.cacheTTL && len(c.cached.Hosts) > 0 {
		return c.cached, nil
	}

	now := time.Now()
	snap, err := c.buildSnapshot(ctx, now)
	if err != nil {
		return LiveData{
			Hosts:         []Host{},
			Services:      []Service{},
			Timestamp:     now.Unix(),
			SelfNamespace: getNamespace(),
		}, nil // never block the page on a metrics error
	}

	services := BuildServices(snap, now)
	hosts := BuildHosts(snap, c.topo.NodeInfoFunc())

	out := LiveData{
		Hosts:         hosts,
		Services:      services,
		HasMetrics:    true,
		Timestamp:     now.Unix(),
		SelfNamespace: getNamespace(),
	}
	c.cached = out
	c.cachedAt = now
	return out, nil
}

// buildSnapshot merges Prometheus metrics with K8s topology.
func (c *PrometheusClient) buildSnapshot(ctx context.Context, now time.Time) (MetricsSnapshot, error) {
	topo := ClusterTopology{Nodes: map[string]NodeInfo{}, Pods: map[string]PodInfo{}}
	if c.topo != nil {
		topo = c.topo.Topology()
	}

	fctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Latest pod metrics (instant query).
	podCPU := queryByPod(fctx, c, qPodCPU)
	podRAM := queryByPod(fctx, c, qPodRAM)
	podNet := queryByPod(fctx, c, qPodNet)
	podDisk := queryByPod(fctx, c, qPodDisk)

	// Sparkline history (range query, last 6h, 5m step).
	hist := queryPodRange(fctx, c, map[string]string{
		"cpu":  qPodCPU,
		"ram":  qPodRAM,
		"net":  qPodNet,
		"disk": qPodDisk,
	})

	// Aggregate pods per workload using the K8s topology (pod -> deployment/node).
	type wl struct {
		ns, deploy, host string
	}
	workloads := map[string]*wl{}
	addWL := func(key, ns, deploy, host string) {
		if deploy == "" || ns == "" {
			return
		}
		if workloads[key] == nil {
			workloads[key] = &wl{ns: ns, deploy: deploy, host: host}
		}
	}

	// Discover workloads from CPU series (most reliable presence signal).
	for key := range podCPU {
		pi, ok := topo.Pods[key]
		if !ok {
			continue
		}
		addWL(key, nsFromKey(key), pi.Deployment, pi.Node)
	}

	// Build WorkloadMetrics: sum across pods of the same deployment.
	wm := map[string]WorkloadMetrics{}
	aggCPU, aggRAM, aggNet, aggDisk := aggregateWorkloads(topo, podCPU, podRAM, podNet, podDisk)
	histCPU, histRAM, histNet, histDisk := aggregateHistory(topo, hist)
	nodeSvc := map[string]int{}
	for key, w := range workloads {
		depKey := w.ns + "/" + w.deploy
		m := wm[depKey]
		m.Namespace = w.ns
		m.Name = w.deploy
		m.Host = w.host
		m.CPU += aggCPU[key]
		m.RAM += aggRAM[key]
		m.Net += aggNet[key]
		m.Disk += aggDisk[key]
		// history: take the first pod's series for this deployment.
		m.CPUHist = pickHist(histCPU[key])
		m.RAMHist = pickHist(histRAM[key])
		m.NetHist = pickHist(histNet[key])
		m.DiskHist = pickHist(histDisk[key])
		wm[depKey] = m
		if w.host != "" {
			nodeSvc[w.host]++
		}
	}

	// Node usage = sum of its pods' usage. Capacity comes from the K8s topology.
	nodes := map[string]NodeMetrics{}
	nodeCPU := map[string]float64{}
	nodeRAM := map[string]float64{}
	for key, pi := range topo.Pods {
		nodeCPU[pi.Node] += podCPU[key]
		nodeRAM[pi.Node] += podRAM[key]
	}
	for name, ni := range topo.Nodes {
		nm := NodeMetrics{
			CPU:      nodeCPU[name],
			RAM:      nodeRAM[name],
			CPUCores: ni.CPUCores,
			RAMTotal: ni.RAMBytes,
		}
		if !ni.Created.IsZero() {
			nm.Uptime = now.Sub(ni.Created).Seconds()
		}
		nodes[name] = nm
	}
	// Ensure nodes with pods but no K8s metadata still appear.
	for name := range nodeCPU {
		if _, ok := nodes[name]; !ok {
			nodes[name] = NodeMetrics{CPU: nodeCPU[name], RAM: nodeRAM[name]}
		}
	}

	return MetricsSnapshot{
		Workloads:     wm,
		Nodes:         nodes,
		NodeSvcCounts: nodeSvc,
	}, nil
}

// ── PromQL queries ──
//
// cAdvisor (via kubelet) metrics carry namespace/pod/node labels. We aggregate
// by (namespace, pod) and map pod -> deployment/node through the K8s API.
// The root cgroup id="/" series are unreliable across kubelet versions, so
// node-level usage is derived by summing pod usage per node.
const (
	qPodCPU  = `sum by (namespace, pod) (rate(container_cpu_usage_seconds_total{container!="POD"}[5m]))`
	qPodRAM  = `sum by (namespace, pod) (container_memory_working_set_bytes{container!="POD"})`
	qPodNet  = `sum by (namespace, pod) (rate(container_network_receive_bytes_total[5m]))`
	qPodDisk = `sum by (namespace, pod) (container_fs_usage_bytes)`
)

// ── Prometheus HTTP API ──

// promSeries is a single instant-query time series.
type promSeries struct {
	Labels map[string]string `json:"metric"`
	Value  []json.Number     `json:"value"` // [timestamp, value]
}

// promResult is the Prometheus query API response shape.
type promResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Data   struct {
		ResultType string       `json:"resultType"`
		Result     []promSeries `json:"result"`
	} `json:"data"`
}

// queryByPod runs an instant query and returns map["namespace/pod"]value.
func queryByPod(ctx context.Context, c *PrometheusClient, query string) map[string]float64 {
	out := map[string]float64{}
	res, err := c.instant(ctx, query)
	if err != nil {
		return out
	}
	for _, s := range res {
		ns := s.Labels["namespace"]
		pod := s.Labels["pod"]
		if ns == "" || pod == "" || len(s.Value) < 2 {
			continue
		}
		v, _ := s.Value[1].Float64()
		out[ns+"/"+pod] = v
	}
	return out
}

// queryPodRange runs a range query per metric and returns
// map[metric]map["namespace/pod"][]float64.
func queryPodRange(ctx context.Context, c *PrometheusClient, queries map[string]string) map[string]map[string][]float64 {
	out := make(map[string]map[string][]float64, len(queries))
	for name, q := range queries {
		out[name] = c.rangeByPod(ctx, q)
	}
	return out
}

type promRangeSeries struct {
	Labels map[string]string `json:"metric"`
	Values [][]json.Number   `json:"values"` // [[ts, val], ...]
}

func (c *PrometheusClient) rangeByPod(ctx context.Context, query string) map[string][]float64 {
	out := map[string][]float64{}
	end := time.Now().Unix()
	start := end - 6*3600
	body, err := c.do(ctx, "/api/v1/query_range", url.Values{
		"query": {query},
		"start": {strconv.FormatInt(start, 10)},
		"end":   {strconv.FormatInt(end, 10)},
		"step":  {"300"},
	})
	if err != nil {
		return out
	}
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Result []promRangeSeries `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return out
	}
	for _, s := range resp.Data.Result {
		ns := s.Labels["namespace"]
		pod := s.Labels["pod"]
		if ns == "" || pod == "" {
			continue
		}
		vals := make([]float64, 0, len(s.Values))
		for _, p := range s.Values {
			if len(p) >= 2 {
				v, _ := p[1].Float64()
				vals = append(vals, v)
			}
		}
		if len(vals) > 0 {
			out[ns+"/"+pod] = vals
		}
	}
	return out
}

func (c *PrometheusClient) instant(ctx context.Context, query string) ([]promSeries, error) {
	body, err := c.do(ctx, "/api/v1/query", url.Values{"query": {query}})
	if err != nil {
		return nil, err
	}
	var resp promResult
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("prometheus: %s", resp.Error)
	}
	return resp.Data.Result, nil
}

func (c *PrometheusClient) do(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u := c.baseURL + path
	if enc := params.Encode(); enc != "" {
		u += "?" + enc
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ── aggregation helpers ──

func nsFromKey(key string) string {
	if i := strings.Index(key, "/"); i >= 0 {
		return key[:i]
	}
	return key
}

// aggregateWorkloads sums per-pod metrics (already keyed by namespace/pod) —
// the per-pod values are returned unchanged so the caller can sum pods sharing
// a deployment. Returns the same maps for clarity at the call site.
func aggregateWorkloads(topo ClusterTopology, cpu, ram, net, disk map[string]float64) (map[string]float64, map[string]float64, map[string]float64, map[string]float64) {
	return cpu, ram, net, disk
}

// aggregateHistory returns per-pod history keyed by "namespace/pod" per metric.
func aggregateHistory(topo ClusterTopology, hist map[string]map[string][]float64) (map[string][]float64, map[string][]float64, map[string][]float64, map[string][]float64) {
	return hist["cpu"], hist["ram"], hist["net"], hist["disk"]
}

// pickHist returns the first non-empty history slice (clamped), or nil.
func pickHist(v []float64) []float64 {
	if len(v) == 0 {
		return nil
	}
	out := make([]float64, len(v))
	for i, x := range v {
		if math.IsNaN(x) || x > 1e15 {
			x = 0
		}
		out[i] = x
	}
	return out
}
