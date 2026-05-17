package obs

import (
	"context"
)

// Service represents one live service discovered from SigNoz metrics.
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
	Name     string  `json:"name"`
	Label    string  `json:"label"`
	Detail   string  `json:"detail"`
	CPU      float64 `json:"cpu"`
	RAM      float64 `json:"ram"`
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

// Client fetches live service data from SigNoz metrics.
type Client interface {
	Fetch(ctx context.Context) (LiveData, error)
}
