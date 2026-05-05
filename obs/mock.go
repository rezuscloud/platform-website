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
	data := PlatformTopology()

	// Mark all healthy in mock
	for i := range data.Services {
		data.Services[i].Status = "healthy"
	}

	// Add sample metrics to platform-website
	data.Services[1].Metrics = []MetricSeries{
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

	// Add sample metrics to daprd
	data.Services[2].Metrics = []MetricSeries{
		{
			Label:  "Components",
			Value:  "3",
			Unit:   "loaded",
			Points: []float64{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
		},
	}

	return data
}
