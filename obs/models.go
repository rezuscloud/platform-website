package obs

// ServiceNode is a service in the platform-website dependency graph.
type ServiceNode struct {
	Name    string // unique ID: "cilium-gateway"
	Label   string // display: "Cilium Gateway"
	Kind    string // "ingress", "app", "sidecar", "infra"
	Status  string // "healthy" or "unknown"
	Detail  string // e.g. "Go/Fiber v2"
	Metrics []MetricSeries
}

// ServiceEdge is a directed connection between two services.
type ServiceEdge struct {
	From  string // source ServiceNode.Name
	To    string // target ServiceNode.Name
	Label string // "HTTP", "OTLP", "gRPC"
}

// LiveData holds everything the live section template needs.
type LiveData struct {
	Services []ServiceNode
	Edges    []ServiceEdge
}

// PlatformTopology returns the known service map for the platform-website.
// The topology is static. Only the status and metrics come from SigNoz.
func PlatformTopology() LiveData {
	return LiveData{
		Services: []ServiceNode{
			{Name: "cilium-gateway", Label: "Cilium Gateway", Kind: "ingress", Detail: "Gateway API + TLS"},
			{Name: "platform-website", Label: "platform-website", Kind: "app", Detail: "Go / Fiber v2"},
			{Name: "daprd", Label: "Dapr Sidecar", Kind: "sidecar", Detail: "daprd v1.15"},
			{Name: "signoz-collector", Label: "SigNoz Collector", Kind: "infra", Detail: "OTEL Receiver"},
			{Name: "dapr-control-plane", Label: "Dapr Control Plane", Kind: "infra", Detail: "Placement + Sentry"},
		},
		Edges: []ServiceEdge{
			{From: "cilium-gateway", To: "platform-website", Label: "HTTPS"},
			{From: "platform-website", To: "daprd", Label: "localhost"},
			{From: "daprd", To: "signoz-collector", Label: "OTLP"},
			{From: "daprd", To: "dapr-control-plane", Label: "gRPC"},
		},
	}
}

// MetricSeries holds a named metric with sparkline data.
type MetricSeries struct {
	Label  string
	Value  string
	Unit   string
	Points []float64
}

// TierGroup groups nodes by infrastructure tier.
type TierGroup struct {
	Name  string
	Tier  string
	Nodes []Node
}

// Node and Pod kept for backward compat with mock client.
// Will be removed once SigNoz client is wired in production.

type Node struct {
	Name   string
	Tier   string
	Status string
	CPU    string
	Mem    string
	Pods   []Pod
}

type Pod struct {
	Name      string
	Namespace string
	Status    string
	Restarts  int
}
