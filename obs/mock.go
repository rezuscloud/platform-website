package obs

import "context"

// MockClient returns the platform topology without live data.
type MockClient struct {
	Data LiveData
	Err  error
}

func (m *MockClient) Fetch(_ context.Context) (LiveData, error) {
	return m.Data, m.Err
}

// DefaultMockData returns the full E2E platform topology.
// Services in monitored namespaces show "healthy" to demonstrate the layout.
func DefaultMockData() LiveData {
	cats := PlatformCategories()
	monitored := MonitoredNamespaces()

	for i := range cats {
		for j := range cats[i].Services {
			svc := &cats[i].Services[j]
			if monitored[svc.Namespace] {
				svc.Status = "healthy"
			} else if svc.Namespace == "" {
				// Infrastructure nodes are always "running"
				svc.Status = "running"
			} else {
				svc.Status = "unmonitored"
			}
		}
	}

	// Add a sample metric to platform-website
	for i := range cats {
		if cats[i].ID == "runtime" {
			for j := range cats[i].Services {
				if cats[i].Services[j].Name == "platform-website" {
					cats[i].Services[j].Metric = "75 goroutines"
				}
				if cats[i].Services[j].Name == "daprd" {
					cats[i].Services[j].Metric = "2 components"
				}
			}
		}
	}

	return LiveData{
		Categories: cats,
		HasMetrics: false,
	}
}
