package obs

import (
	"context"
)

// PlatformCategory groups related services in the platform dashboard.
type PlatformCategory struct {
	Name     string
	ID       string // "infra", "dev", "delivery", "runtime", "observability"
	Services []PlatformService
}

// PlatformService is one service in the platform dashboard.
type PlatformService struct {
	Name      string // unique ID
	Label     string // display name
	Detail    string // e.g. "Self-hosted Git"
	Status    string // "healthy", "unknown", "unmonitored"
	Namespace string // K8s namespace for SigNoz queries
	Metric    string // compact metric value, e.g. "75 goroutines"
}

// StatsStrip shows runtime metadata above the dashboard.
type StatsStrip struct {
	Uptime    string
	GoVersion string
	NodeCount int
}

// HealthCheck is one line in the health strip below the dashboard.
type HealthCheck struct {
	ServiceName string
	Status      string
	LastCheck   string
}

// LiveData holds everything the live section template needs.
type LiveData struct {
	Categories []PlatformCategory
	Stats      StatsStrip
	HasMetrics bool
}

// Client fetches live service tree data from SigNoz metrics.
type Client interface {
	Fetch(ctx context.Context) (LiveData, error)
}

// PlatformCategories returns the 5-column E2E platform topology.
func PlatformCategories() []PlatformCategory {
	return []PlatformCategory{
		{
			Name: "Infrastructure",
			ID:   "infra",
			Services: []PlatformService{
				{Name: "oci-cloud", Label: "OCI Cloud Node", Detail: "ARM64 · Ampere A1", Namespace: ""},
				{Name: "edge-node", Label: "Edge Node", Detail: "AMD64 · NUC", Namespace: ""},
			},
		},
		{
			Name: "Development",
			ID:   "dev",
			Services: []PlatformService{
				{Name: "forgejo", Label: "Forgejo", Detail: "Self-hosted Git", Namespace: "forgejo"},
				{Name: "arc-controller", Label: "ARC Controller", Detail: "GitHub Actions runners", Namespace: "arc-systems"},
			},
		},
		{
			Name: "Delivery",
			ID:   "delivery",
			Services: []PlatformService{
				{Name: "flux-source", Label: "Flux Source", Detail: "Git repository sync", Namespace: "flux-system"},
				{Name: "flux-kustomize", Label: "Flux Kustomize", Detail: "Manifest reconciliation", Namespace: "flux-system"},
				{Name: "flux-helm", Label: "Flux Helm", Detail: "Helm release manager", Namespace: "flux-system"},
				{Name: "kubevela", Label: "KubeVela", Detail: "Application delivery", Namespace: "vela-system"},
				{Name: "external-dns", Label: "External DNS", Detail: "Cloudflare DNS automation", Namespace: "external-dns"},
				{Name: "cert-manager", Label: "Cert Manager", Detail: "TLS certificate lifecycle", Namespace: "cert-manager"},
			},
		},
		{
			Name: "Runtime",
			ID:   "runtime",
			Services: []PlatformService{
				{Name: "cilium", Label: "Cilium CNI", Detail: "eBPF data plane", Namespace: "kube-system"},
				{Name: "platform-website", Label: "platform-website", Detail: "Go / Fiber v2", Namespace: "platform-website"},
				{Name: "daprd", Label: "Dapr Sidecar", Detail: "daprd v1.15", Namespace: "platform-website"},
				{Name: "dapr-control-plane", Label: "Dapr Control Plane", Detail: "Placement + Sentry", Namespace: "dapr-system"},
			},
		},
		{
			Name: "Observability",
			ID:   "observability",
			Services: []PlatformService{
				{Name: "signoz-collector", Label: "SigNoz Collector", Detail: "OTEL Receiver", Namespace: "signoz"},
				{Name: "signoz-clickhouse", Label: "ClickHouse", Detail: "Columnar telemetry store", Namespace: "signoz"},
				{Name: "signoz-query", Label: "SigNoz Query Service", Detail: "API + Dashboards", Namespace: "signoz"},
			},
		},
	}
}

// MonitoredNamespaces returns the set of namespaces that have SigNoz scrape data.
func MonitoredNamespaces() map[string]bool {
	return map[string]bool{
		"flux-system":      true,
		"vela-system":      true,
		"kube-system":      true,
		"platform-website": true,
		"dapr-system":      true,
		"cert-manager":     true,
		"external-dns":     true,
		"velero":           true,
		"tikv-system":      true,
		"juicefs-csi":      true,
	}
}
