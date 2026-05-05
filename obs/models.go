package obs

import "context"

// ServiceNode is a service in the platform-website dependency tree.
type ServiceNode struct {
	Name    string // unique ID: "cilium-gateway"
	Label   string // display: "Cilium Gateway"
	Kind    string // "ingress", "app", "sidecar", "infra"
	Status  string // "healthy" or "unknown"
	Detail  string // e.g. "Go / Fiber v2"
	Metrics []MetricSeries
	Out     []ServiceEdge // outgoing edges to downstream services
}

// ServiceEdge is a directed connection to a downstream service.
type ServiceEdge struct {
	Label  string      // "HTTPS", "OTLP", "gRPC"
	Target ServiceNode // the downstream service
}

// MetricSeries holds a named metric value (compact: one key number per node).
type MetricSeries struct {
	Label string
	Value string
	Unit  string
}

// StatsStrip shows runtime metadata above the diagram.
type StatsStrip struct {
	Uptime    string // "47h" from process_start_time_seconds
	GoVersion string // "go1.26" from go_info metric
	NodeCount int    // physical cluster nodes
}

// HealthCheck is one line in the health strip below the diagram.
type HealthCheck struct {
	ServiceName string
	Status      string // "healthy" or "unknown"
	LastCheck   string // "5s ago"
}

// LiveData holds everything the live section template needs.
type LiveData struct {
	Root       ServiceNode
	Stats      StatsStrip
	Health     []HealthCheck
	HasMetrics bool // false when SigNoz not configured
}

// Client fetches live service tree data from SigNoz metrics.
type Client interface {
	Fetch(ctx context.Context) (LiveData, error)
}

// Walk traverses the service tree depth-first, calling fn on each node.
func (s *ServiceNode) Walk(fn func(*ServiceNode)) {
	fn(s)
	for i := range s.Out {
		s.Out[i].Target.Walk(fn)
	}
}

// ServiceCount returns the total number of services in the tree.
func (s *ServiceNode) ServiceCount() int {
	count := 1
	for i := range s.Out {
		count += s.Out[i].Target.ServiceCount()
	}
	return count
}

// Find returns the first node matching name, or nil.
func (s *ServiceNode) Find(name string) *ServiceNode {
	if s.Name == name {
		return s
	}
	for i := range s.Out {
		if found := s.Out[i].Target.Find(name); found != nil {
			return found
		}
	}
	return nil
}

// PlatformTopology returns the known dependency tree for platform-website.
func PlatformTopology() ServiceNode {
	return ServiceNode{
		Name: "cilium-gateway", Label: "Cilium Gateway", Kind: "ingress",
		Detail: "Gateway API + TLS", Status: "unknown",
		Out: []ServiceEdge{
			{Label: "HTTPS", Target: ServiceNode{
				Name: "platform-website", Label: "platform-website", Kind: "app",
				Detail: "Go / Fiber v2", Status: "unknown",
				Out: []ServiceEdge{
					{Label: "localhost", Target: ServiceNode{
						Name: "daprd", Label: "Dapr Sidecar", Kind: "sidecar",
						Detail: "daprd v1.15", Status: "unknown",
						Out: []ServiceEdge{
							{Label: "OTLP", Target: ServiceNode{
								Name: "signoz-collector", Label: "SigNoz Collector", Kind: "infra",
								Detail: "OTEL Receiver", Status: "unknown",
							}},
							{Label: "gRPC", Target: ServiceNode{
								Name: "dapr-control-plane", Label: "Dapr Control Plane", Kind: "infra",
								Detail: "Placement + Sentry", Status: "unknown",
							}},
						},
					}},
				},
			}},
		},
	}
}
