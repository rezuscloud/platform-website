package obs

import (
	"context"
	"fmt"
	"time"
)

// Service represents one live service discovered from SigNoz metrics.
type Service struct {
	Name      string  `json:"name"`               // deployment or statefulset name
	Namespace string  `json:"namespace"`          // k8s namespace
	Category  string  `json:"category"`           // dev, deployment, runtime, observability, data
	Host      string  `json:"host"`               // k8s node name this service runs on
	Status    string  `json:"status"`             // healthy, running
	Detail    string  `json:"detail,omitempty"`   // e.g. "ARM64 · Ampere A1" for hosts
	CPU       float64 `json:"cpu"`                // CPU % (rate over 5m)
	RAM       float64 `json:"ram"`                // RAM in MB
	NetKB     float64 `json:"netKB"`              // Network throughput in KB/s
	DiskMB    float64 `json:"diskMB"`             // Disk (filesystem) usage in MB
	LoadAvg   float64 `json:"loadAvg,omitempty"`  // system load average 5m (hosts only)
	IOWait    float64 `json:"ioWait,omitempty"`   // CPU % spent in iowait (hosts only)
	Uptime    string  `json:"uptime,omitempty"`   // e.g. "4d", "12h", "45m"
	CPUHist   string  `json:"cpuHist,omitempty"`  // SVG polyline points for CPU sparkline
	RAMHist   string  `json:"ramHist,omitempty"`  // SVG polyline points for RAM sparkline
	NetHist   string  `json:"netHist,omitempty"`  // SVG polyline points for network sparkline
	DiskHist  string  `json:"diskHist,omitempty"` // SVG polyline points for disk sparkline
}

// Host represents a physical or virtual machine node.
type Host struct {
	Name     string  `json:"name"`     // k8s node name
	Label    string  `json:"label"`    // display name (e.g. "OCI Cloud")
	Detail   string  `json:"detail"`   // e.g. "ARM64 · Ampere A1"
	CPU      float64 `json:"cpu"`      // node CPU %
	RAM      float64 `json:"ram"`      // node RAM in MB
	LoadAvg  float64 `json:"loadAvg"`  // system load average 5m
	IOWait   float64 `json:"ioWait"`   // CPU % spent in iowait
	Uptime   string  `json:"uptime"`   // e.g. "5d"
	SvcCount int     `json:"svcCount"` // number of monitored services on this node
}

// LiveData holds everything the live section needs.
type LiveData struct {
	Hosts         []Host    `json:"hosts"`
	Services      []Service `json:"services"`
	HasMetrics    bool      `json:"hasMetrics"`
	Timestamp     int64     `json:"timestamp"`
	SelfNamespace string    `json:"selfNamespace"`
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

// CategoryOrder defines the row group order.
var CategoryOrder = []string{"dev", "deployment", "runtime", "observability", "data"}

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
