package obs

import (
	"context"
	"time"
)

// MockClient returns the platform topology without live data.
type MockClient struct {
	Data LiveData
	Err  error
}

func (m *MockClient) Fetch(_ context.Context) (LiveData, error) {
	return m.Data, m.Err
}

// DefaultMockData returns the full E2E platform topology with mock live data.
// Shows what the dashboard looks like with metrics flowing.
func DefaultMockData() LiveData {
	cats := PlatformCategories()
	now := time.Now().Unix()

	for i := range cats {
		for j := range cats[i].Services {
			svc := &cats[i].Services[j]
			svc.UpdatedAt = now

			if svc.Namespace == "" {
				svc.Status = "running"
			} else if svc.Deployment == "" {
				svc.Status = "unmonitored"
			} else {
				svc.Status = "healthy"
			}
		}
	}

	// Sample metrics for monitored services
	mockMetrics := map[string]struct{ metric, memory string }{
		"flux-source":        {"66 goroutines", "106 MB"},
		"flux-kustomize":     {"80 goroutines", "79 MB"},
		"flux-helm":          {"83 goroutines", "101 MB"},
		"kubevela":           {"154 goroutines", "114 MB"},
		"external-dns":       {"89 goroutines", "92 MB"},
		"cert-manager":       {"400 goroutines", "71 MB"},
		"cilium":             {"460 goroutines", "45 MB"},
		"platform-website":   {"75 goroutines", "121 MB"},
		"daprd":              {"2 components", ""},
		"dapr-control-plane": {"339 goroutines", "54 MB"},
	}

	for i := range cats {
		for j := range cats[i].Services {
			if m, ok := mockMetrics[cats[i].Services[j].Name]; ok {
				cats[i].Services[j].Metric = m.metric
				cats[i].Services[j].Memory = m.memory
			}
		}
	}

	return LiveData{
		Categories: cats,
		HasMetrics: false,
		Timestamp:  now,
	}
}
