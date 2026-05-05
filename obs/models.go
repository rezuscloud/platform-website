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
	Name      string // "platform-website-shell-abc123"
	Namespace string // "platform-website"
	Status    string // "Running", "Pending"
	Restarts  int
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

// TierGroup groups nodes by infrastructure tier.
type TierGroup struct {
	Name  string // "OCI Cloud" or "Edge"
	Tier  string // "oci-cloud" or "edge"
	Nodes []Node
}

// GroupedByTier returns nodes grouped by tier in display order.
func (d LiveData) GroupedByTier() []TierGroup {
	var groups []TierGroup
	byTier := make(map[string]*TierGroup)
	var order []string

	for _, n := range d.Nodes {
		tier := n.Tier
		if _, ok := byTier[tier]; !ok {
			order = append(order, tier)
			groups = append(groups, TierGroup{Name: tierDisplayName(tier), Tier: tier})
			byTier[tier] = &groups[len(groups)-1]
		}
		byTier[tier].Nodes = append(byTier[tier].Nodes, n)
	}
	return groups
}

func tierDisplayName(tier string) string {
	switch tier {
	case "oci-cloud":
		return "OCI Cloud"
	case "edge":
		return "Edge"
	default:
		return tier
	}
}
