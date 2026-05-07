package obs

// ServiceMap defines the static topology of the platform website's dependency chain.
// Services are discovered dynamically from SigNoz, but the dependency edges are
// known from the platform architecture.

// MapLayer is a horizontal row in the service map.
type MapLayer struct {
	ID    string
	Label string
}

// MapNode is a single node in the service map, matched to a live Service.
type MapNode struct {
	ID      string // unique identifier (used for edges)
	Label   string // display name
	LayerID string
	Key     string // "namespace/name" or "/hostname" to match Service
	Static  bool   // true if not matched from SigNoz
	Detail  string // static detail (e.g. "Browser", "Cilium · TLS")
}

// MapEdge connects two nodes in the service map.
type MapEdge struct {
	From string
	To   string
}

// PlatformLayers is the ordered list of layers (top to bottom).
var PlatformLayers = []MapLayer{
	{ID: "traffic", Label: "Traffic"},
	{ID: "app", Label: "Application"},
	{ID: "runtime", Label: "Runtime"},
	{ID: "delivery", Label: "Delivery"},
	{ID: "observability", Label: "Observability"},
	{ID: "storage", Label: "Storage"},
	{ID: "infra", Label: "Infrastructure"},
}

// PlatformNodes defines all nodes and their service key matches.
var PlatformNodes = []MapNode{
	// Traffic
	{ID: "visitor", LayerID: "traffic", Label: "Visitor", Static: true, Detail: "Browser"},
	{ID: "gateway", LayerID: "traffic", Label: "Gateway", Static: true, Detail: "Cilium · TLS"},
	// Application
	{ID: "website", LayerID: "app", Label: "Platform Website", Key: "platform-website/platform-website"},
	// Runtime
	{ID: "dapr", LayerID: "runtime", Label: "Dapr", Key: "dapr-system/dapr-operator"},
	{ID: "cilium", LayerID: "runtime", Label: "Cilium", Key: "kube-system/cilium-operator"},
	{ID: "cert", LayerID: "runtime", Label: "cert-manager", Key: "cert-manager/cert-manager"},
	{ID: "dns", LayerID: "runtime", Label: "external-dns", Key: "external-dns/external-dns"},
	// Delivery
	{ID: "flux", LayerID: "delivery", Label: "Flux CD", Key: "flux-system/source-controller"},
	{ID: "vela", LayerID: "delivery", Label: "KubeVela", Key: "vela-system/kubevela-vela-core"},
	// Observability
	{ID: "signoz", LayerID: "observability", Label: "SigNoz", Key: "signoz/chi-signoz-clickhouse-cluster-0-0"},
	// Storage
	{ID: "juicefs", LayerID: "storage", Label: "JuiceFS", Key: "juicefs-csi/juicefs-csi-controller"},
	{ID: "tikv", LayerID: "storage", Label: "TiKV", Key: "tikv-system/juicefs-tikv-pd"},
	// Infrastructure
	{ID: "oci", LayerID: "infra", Label: "OCI Cloud", Key: "/talosoci-control-plane-legal-poodle"},
	{ID: "edge", LayerID: "infra", Label: "Edge Node", Key: "/talosedge-genmachiche-flowing-bluejay"},
}

// PlatformEdges defines the dependency connections.
var PlatformEdges = []MapEdge{
	{From: "visitor", To: "gateway"},
	{From: "gateway", To: "website"},
	{From: "website", To: "dapr"},
	{From: "website", To: "cilium"},
	{From: "website", To: "cert"},
	{From: "website", To: "dns"},
	{From: "flux", To: "website"},
	{From: "vela", To: "website"},
	{From: "website", To: "signoz"},
	{From: "signoz", To: "juicefs"},
	{From: "signoz", To: "tikv"},
	{From: "juicefs", To: "oci"},
	{From: "tikv", To: "edge"},
	{From: "website", To: "oci"},
}

// FindServiceByKey matches a topology key to a live service.
func FindServiceByKey(services []Service, key string) (Service, bool) {
	for _, s := range services {
		svcKey := s.Namespace + "/" + s.Name
		if svcKey == key {
			return s, true
		}
	}
	return Service{}, false
}

// LayerNodes returns all MapNodes for a given layer ID.
func LayerNodes(layerID string) []MapNode {
	var result []MapNode
	for _, n := range PlatformNodes {
		if n.LayerID == layerID {
			result = append(result, n)
		}
	}
	return result
}

// NodeEdges returns all edges where the given node ID appears as From or To.
func NodeEdges(nodeID string) []MapEdge {
	var result []MapEdge
	for _, e := range PlatformEdges {
		if e.From == nodeID || e.To == nodeID {
			result = append(result, e)
		}
	}
	return result
}
