package obs

import "context"

// MockClient returns static topology without metrics.
type MockClient struct {
	Data LiveData
	Err  error
}

func (m *MockClient) Fetch(_ context.Context) (LiveData, error) {
	return m.Data, m.Err
}

// DefaultMockData returns the platform topology without live metrics.
func DefaultMockData() LiveData {
	root := PlatformTopology()

	root.Walk(func(s *ServiceNode) {
		s.Status = "unknown"
	})

	return LiveData{
		Root:       root,
		HasMetrics: false,
	}
}
