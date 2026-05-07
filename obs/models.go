package obs

import (
	"context"
	"fmt"
	"time"
)

// Service represents one live service discovered from SigNoz metrics.
type Service struct {
	Name      string  `json:"name"`              // deployment or statefulset name
	Namespace string  `json:"namespace"`         // k8s namespace
	Category  string  `json:"category"`          // hosts, dev, deployment, runtime, observability, data
	Status    string  `json:"status"`            // healthy, unknown, running, unmonitored
	Detail    string  `json:"detail,omitempty"`  // e.g. "ARM64 · Ampere A1" for hosts
	CPU       float64 `json:"cpu"`               // CPU % (rate over 5m)
	RAM       float64 `json:"ram"`               // RAM in MB
	LoadAvg   float64 `json:"loadAvg,omitempty"` // system load average 5m
	IOWait    float64 `json:"ioWait,omitempty"`  // CPU % spent in iowait
	Uptime    string  `json:"uptime,omitempty"`  // e.g. "4d", "12h", "45m"
	CPUHist   string  `json:"cpuHist,omitempty"` // SVG polyline points for CPU sparkline
	RAMHist   string  `json:"ramHist,omitempty"` // SVG polyline points for RAM sparkline
}

// LiveData holds everything the live section needs.
type LiveData struct {
	Services   []Service `json:"services"`
	HasMetrics bool      `json:"hasMetrics"`
	Timestamp  int64     `json:"timestamp"`
}

// Client fetches live service data from SigNoz metrics.
type Client interface {
	Fetch(ctx context.Context) (LiveData, error)
}

// CategoryForNamespace maps a namespace to a platform category.
func CategoryForNamespace(ns string) string {
	switch ns {
	case "forgejo", "arc-systems":
		return "dev"
	case "flux-system", "vela-system", "external-dns", "cert-manager":
		return "deployment"
	case "kube-system", "platform-website", "dapr-system":
		return "runtime"
	case "signoz", "monitoring":
		return "observability"
	case "tikv-system", "juicefs-csi", "velero":
		return "data"
	default:
		return "runtime"
	}
}

// CategoryOrder defines the left-to-right column order.
var CategoryOrder = []string{"hosts", "dev", "deployment", "runtime", "observability", "data"}

// CategoryLabel returns the display name for a category ID.
func CategoryLabel(id string) string {
	switch id {
	case "hosts":
		return "Hosts"
	case "dev":
		return "Development"
	case "deployment":
		return "Deployment"
	case "runtime":
		return "Runtime"
	case "observability":
		return "Observability"
	case "data":
		return "Data"
	default:
		return id
	}
}

// StaticHosts returns infrastructure host entries.
func StaticHosts() []Service {
	return []Service{
		{Name: "oci-cloud", Namespace: "", Category: "hosts", Status: "running", Detail: "ARM64 · Ampere A1"},
		{Name: "edge-node", Namespace: "", Category: "hosts", Status: "running", Detail: "AMD64 · Intel NUC"},
	}
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
