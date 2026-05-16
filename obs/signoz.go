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

// Fetch returns live platform data, using a 30s cache to avoid
// hitting SigNoz on every SSE tick.
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
		return LiveData{
			Hosts:     staticHosts(),
			Timestamp: now.Unix(),
		}, nil
	}

	results := map[string][]v3Series{}
	for _, r := range resp.Data.Result {
		results[r.QueryName] = r.Series
	}

	services := BuildServices(results["cpu"], results, now)
	hosts := BuildHosts(results)

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

func staticHosts() []Host {
	return []Host{}
}

// latestByPod is kept for backward compatibility with tests.
func latestByPod(series []v3Series) map[string]float64 {
	m := map[string]float64{}
	for _, s := range series {
		ns := LabelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		pod := LabelStr(s.Labels, "k8s.pod.name")
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

// Deprecated: Use DiscoverNodeNames instead.
func discoverNodeNames(sources ...any) []string {
	return DiscoverNodeNames(sources...)
}

// Deprecated: Use LabelStr instead.
func labelStr(labels map[string]string, keys ...string) string {
	return LabelStr(labels, keys...)
}

// Deprecated: Use SortServices instead.
func sortServices(services []Service) {
	SortServices(services)
}

// ServicesByCategory filters services by category.
func ServicesByCategory(services []Service, cat string) []Service {
	var result []Service
	for _, s := range services {
		if s.Category == cat {
			result = append(result, s)
		}
	}
	return result
}

// HostNames returns an ordered list of host names.
func HostNames(hosts []Host) []string {
	names := make([]string, len(hosts))
	for i, h := range hosts {
		names[i] = h.Name
	}
	return names
}

// FormatUptime returns a human-readable duration.
func FormatUptime(d time.Duration) string {
	h := int(d.Hours())
	if h >= 24 {
		return fmt.Sprintf("%dd", h/24)
	}
	if h >= 1 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// Ensure unused imports are referenced
var _ = math.Round
