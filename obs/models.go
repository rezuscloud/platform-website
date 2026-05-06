package obs

import (
	"context"
	"fmt"
	"time"
)

// PlatformCategory groups related services in the platform dashboard.
type PlatformCategory struct {
	Name     string            `json:"name"`
	ID       string            `json:"id"`
	Services []PlatformService `json:"services"`
}

// PlatformService is one service in the platform dashboard.
type PlatformService struct {
	Name       string `json:"name"`                 // unique ID
	Label      string `json:"label"`                // display name
	Detail     string `json:"detail"`               // e.g. "Self-hosted Git"
	Status     string `json:"status"`               // "healthy", "unknown", "unmonitored", "running"
	Namespace  string `json:"namespace,omitempty"`  // K8s namespace
	Deployment string `json:"deployment,omitempty"` // K8s deployment for SigNoz matching
	Metric     string `json:"metric,omitempty"`     // live metric, e.g. "75 goroutines"
	Memory     string `json:"memory,omitempty"`     // e.g. "121 MB"
	Uptime     string `json:"uptime,omitempty"`     // e.g. "2d"
	UpdatedAt  int64  `json:"updatedAt,omitempty"`  // unix timestamp
}

// StatsStrip shows runtime metadata above the dashboard.
type StatsStrip struct {
	Uptime    string `json:"uptime,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
	NodeCount int    `json:"nodeCount"`
}

// LiveData holds everything the live section template needs.
type LiveData struct {
	Categories []PlatformCategory `json:"categories"`
	Stats      StatsStrip         `json:"stats"`
	HasMetrics bool               `json:"hasMetrics"`
	Timestamp  int64              `json:"timestamp"`
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
				{Name: "oci-cloud", Label: "OCI Cloud Node", Detail: "ARM64 · Ampere A1"},
				{Name: "edge-node", Label: "Edge Node", Detail: "AMD64 · Intel NUC"},
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
				{Name: "flux-source", Label: "Flux Source", Detail: "Git repository sync", Namespace: "flux-system", Deployment: "source-controller"},
				{Name: "flux-kustomize", Label: "Flux Kustomize", Detail: "Manifest reconciliation", Namespace: "flux-system", Deployment: "kustomize-controller"},
				{Name: "flux-helm", Label: "Flux Helm", Detail: "Helm release manager", Namespace: "flux-system", Deployment: "helm-controller"},
				{Name: "kubevela", Label: "KubeVela", Detail: "Application delivery", Namespace: "vela-system", Deployment: "kubevela-vela-core"},
				{Name: "external-dns", Label: "External DNS", Detail: "Cloudflare DNS automation", Namespace: "external-dns", Deployment: "external-dns"},
				{Name: "cert-manager", Label: "Cert Manager", Detail: "TLS certificate lifecycle", Namespace: "cert-manager", Deployment: "cert-manager"},
			},
		},
		{
			Name: "Runtime",
			ID:   "runtime",
			Services: []PlatformService{
				{Name: "cilium", Label: "Cilium CNI", Detail: "eBPF data plane", Namespace: "kube-system", Deployment: "cilium-operator"},
				{Name: "platform-website", Label: "platform-website", Detail: "Go / Fiber v2", Namespace: "platform-website", Deployment: "platform-website"},
				{Name: "daprd", Label: "Dapr Sidecar", Detail: "daprd v1.15", Namespace: "platform-website", Deployment: "platform-website"},
				{Name: "dapr-control-plane", Label: "Dapr Control Plane", Detail: "Operator + Placement + Sentry", Namespace: "dapr-system", Deployment: "dapr-operator"},
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
	}
}

// deploymentKey creates a match key from namespace + deployment.
func deploymentKey(ns, deploy string) string {
	return ns + "/" + deploy
}

// formatDuration returns a human-readable duration string.
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	if h >= 24 {
		return fmt.Sprintf("%dd", h/24)
	}
	if h >= 1 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
