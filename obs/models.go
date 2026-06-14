package obs

import (
	"context"
	"time"
)

// Service represents one live service discovered from Prometheus metrics.
type Service struct {
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Category  string  `json:"category"`
	Host      string  `json:"host"`
	Status    string  `json:"status"`
	Detail    string  `json:"detail,omitempty"`
	CPU       float64 `json:"cpu"`
	RAM       float64 `json:"ram"`
	NetKB     float64 `json:"netKB"`
	DiskMB    float64 `json:"diskMB"`
	LoadAvg   float64 `json:"loadAvg,omitempty"`
	IOWait    float64 `json:"ioWait,omitempty"`
	Uptime    string  `json:"uptime,omitempty"`
	CPUHist   string  `json:"cpuHist,omitempty"`
	RAMHist   string  `json:"ramHist,omitempty"`
	NetHist   string  `json:"netHist,omitempty"`
	DiskHist  string  `json:"diskHist,omitempty"`
}

// Host represents a physical or virtual machine node.
type Host struct {
	Name     string  `json:"name"`     // real node name
	Label    string  `json:"label"`    // role: "Control plane" / "Worker"
	Detail   string  `json:"detail"`   // "OCI Cloud · ARM64"
	CPU      float64 `json:"cpu"`      // cores used
	CPUCores float64 `json:"cpuCores"` // total cores (capacity)
	RAM      float64 `json:"ram"`      // MB used
	RAMTotal float64 `json:"ramTotal"` // MB total (capacity)
	LoadAvg  float64 `json:"loadAvg"`
	IOWait   float64 `json:"ioWait"`
	Uptime   string  `json:"uptime"`
	SvcCount int     `json:"svcCount"`
}

// LiveData holds everything the live section needs.
type LiveData struct {
	Hosts         []Host    `json:"hosts"`
	Services      []Service `json:"services"`
	HasMetrics    bool      `json:"hasMetrics"`
	Timestamp     int64     `json:"timestamp"`
	SelfNamespace string    `json:"selfNamespace"`
}

// MetricsSnapshot is the platform-native shape of a Prometheus + K8s API query.
// PrometheusClient builds this once per refresh; BuildServices and BuildHosts
// consume it. All Prometheus wire-format types stay private to prometheus.go.
type MetricsSnapshot struct {
	Workloads     map[string]WorkloadMetrics // key = "namespace/deployment"
	Nodes         map[string]NodeMetrics     // key = node name
	NodeSvcCounts map[string]int             // key = node name
}

// WorkloadMetrics holds the latest values and sparkline history for one workload.
type WorkloadMetrics struct {
	Namespace string
	Name      string
	Host      string
	Uptime    string // RFC3339 start time

	CPU  float64
	RAM  float64 // bytes
	Net  float64 // bytes/sec
	Disk float64 // bytes

	CPUHist  []float64
	RAMHist  []float64
	NetHist  []float64
	DiskHist []float64
}

// NodeMetrics holds the latest values for one cluster node.
type NodeMetrics struct {
	CPU      float64 // cores used
	RAM      float64 // bytes used
	CPUCores float64 // capacity (cores)
	RAMTotal float64 // capacity (bytes)
	Uptime   float64 // seconds
}

// NodeInfo describes what we know about a cluster node from the K8s API.
type NodeInfo struct {
	IsControlPlane bool
	Provider       string  // e.g. "OCI Cloud", "Edge", empty if unknown
	Arch           string  // e.g. "ARM64", "AMD64", empty if unknown
	CPUCores       float64 // capacity (cores)
	RAMBytes       float64 // capacity (bytes)
	Created        time.Time
}

// NodeInfoFunc looks up node metadata from the Kubernetes API.
// Returns (NodeInfo, true) if found, or (NodeInfo{}, false) if unknown.
type NodeInfoFunc func(nodeName string) (NodeInfo, bool)

// Client fetches live service data from Prometheus metrics.
type Client interface {
	Fetch(ctx context.Context) (LiveData, error)
}
