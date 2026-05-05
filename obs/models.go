package obs

import "context"

// ServiceNode is a service in the platform-website dependency tree.
type ServiceNode struct {
	Name    string         // unique ID: "cilium-gateway"
	Label   string         // display: "Cilium Gateway"
	Kind    string         // "ingress", "app", "sidecar", "infra"
	Status  string         // "healthy" or "unknown"
	Detail  string         // e.g. "Go / Fiber v2"
	Metrics []MetricSeries // live sparkline metrics
	Out     []ServiceEdge  // outgoing edges to downstream services
}

// ServiceEdge is a directed connection to a downstream service.
type ServiceEdge struct {
	Label  string      // "HTTPS", "OTLP", "gRPC"
	Target ServiceNode // the downstream service
}

// MetricSeries holds a named metric with sparkline data.
type MetricSeries struct {
	Label  string
	Value  string
	Unit   string
	Points []float64
}

// LiveData wraps the service tree root.
type LiveData struct {
	Root ServiceNode
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

// PlatformTopology returns the known dependency tree for platform-website.
// The topology is static. Only status and metrics come from SigNoz.
func PlatformTopology() ServiceNode {
	return ServiceNode{
		Name:   "cilium-gateway",
		Label:  "Cilium Gateway",
		Kind:   "ingress",
		Detail: "Gateway API + TLS",
		Out: []ServiceEdge{
			{Label: "HTTPS", Target: ServiceNode{
				Name:   "platform-website",
				Label:  "platform-website",
				Kind:   "app",
				Detail: "Go / Fiber v2",
				Out: []ServiceEdge{
					{Label: "localhost", Target: ServiceNode{
						Name:   "daprd",
						Label:  "Dapr Sidecar",
						Kind:   "sidecar",
						Detail: "daprd v1.15",
						Out: []ServiceEdge{
							{Label: "OTLP", Target: ServiceNode{
								Name:   "signoz-collector",
								Label:  "SigNoz Collector",
								Kind:   "infra",
								Detail: "OTEL Receiver",
							}},
							{Label: "gRPC", Target: ServiceNode{
								Name:   "dapr-control-plane",
								Label:  "Dapr Control Plane",
								Kind:   "infra",
								Detail: "Placement + Sentry",
							}},
						},
					}},
				},
			}},
		},
	}
}
