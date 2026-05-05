package obs

// LiveData holds everything the live section template needs.
type LiveData struct {
	Nodes        []Node
	Metrics      []MetricSeries
	DashboardURL string
}

// Node is a Kubernetes node with its running pods.
type Node struct {
	Name   string // e.g. "talos-oci-cp-0"
	Tier   string // "oci-cloud" or "edge"
	Status string // "Ready", "NotReady"
	CPU    string // formatted usage
	Mem    string // formatted usage
	Pods   []Pod
}

// Pod is a running pod on a node.
type Pod struct {
	Name     string // "platform-website-shell-abc123"
	Status   string // "Running", "Pending"
	Restarts int
}

// MetricSeries holds a named metric with sparkline data.
type MetricSeries struct {
	Label     string
	Value     string
	Unit      string
	Points    []float64
	Sparkline string // pre-computed SVG polyline points
	Area      string // pre-computed SVG area polygon points
}
