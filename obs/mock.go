package obs

import "context"

// MockClient returns static data for development and testing.
type MockClient struct {
	Data LiveData
	Err  error
}

func (m *MockClient) Fetch(_ context.Context) (LiveData, error) {
	return m.Data, m.Err
}

// DefaultMockData returns the platform topology with sample sparklines.
func DefaultMockData() LiveData {
	root := PlatformTopology()

	// Mark all healthy
	root.Walk(func(s *ServiceNode) {
		s.Status = "healthy"
	})

	// Walk to find platform-website and add metrics
	root.Walk(func(s *ServiceNode) {
		if s.Name == "platform-website" {
			s.Metrics = []MetricSeries{
				{
					Label:  "Goroutines",
					Value:  "72",
					Points: []float64{68, 70, 72, 71, 73, 74, 72, 70, 73, 71, 72, 74},
				},
				{
					Label:  "Heap",
					Value:  "17.7",
					Unit:   "MiB",
					Points: []float64{16.2, 16.8, 17.1, 17.4, 17.2, 17.7, 17.5, 17.3, 17.6, 17.8, 17.5, 17.7},
				},
			}
		}
		if s.Name == "daprd" {
			s.Metrics = []MetricSeries{
				{
					Label:  "Components",
					Value:  "3",
					Unit:   "loaded",
					Points: []float64{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
				},
			}
		}
	})

	return LiveData{Root: root}
}
