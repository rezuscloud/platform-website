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

// DefaultMockData returns a realistic cluster topology matching the production environment.
func DefaultMockData() LiveData {
	return LiveData{
		Nodes: []Node{
			{
				Name:   "talos-oci-cp-0",
				Tier:   "oci-cloud",
				Status: "Ready",
				CPU:    "12%",
				Mem:    "4.2 GiB",
				Pods: []Pod{
					{Name: "cilium-operator-6d8f9c7b4-x2k1p", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "coredns-7db8d4f5b9-mn3vq", Namespace: "kube-system", Status: "Running", Restarts: 0},
				},
			},
			{
				Name:   "talos-oci-cp-1",
				Tier:   "oci-cloud",
				Status: "Ready",
				CPU:    "8%",
				Mem:    "3.1 GiB",
				Pods: []Pod{
					{Name: "etcd-talos-oci-cp-1", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "kube-apiserver-talos-oci-cp-1", Namespace: "kube-system", Status: "Running", Restarts: 0},
				},
			},
			{
				Name:   "talosedge-genmachiche-flowing-bluejay",
				Tier:   "edge",
				Status: "Ready",
				CPU:    "34%",
				Mem:    "5.8 GiB",
				Pods: []Pod{
					{Name: "platform-website-6f8d6fd5fc-4rgj9", Namespace: "platform-website-pr57", Status: "Running", Restarts: 0},
					{Name: "signoz-otel-collector-7c9d8f6b5-jk2mn", Namespace: "signoz", Status: "Running", Restarts: 0},
					{Name: "signoz-query-service-0", Namespace: "signoz", Status: "Running", Restarts: 0},
				},
			},
		},
		Metrics: []MetricSeries{
			{
				Label:  "Requests",
				Value:  "1,247",
				Unit:   "req/min",
				Points: []float64{12, 18, 15, 22, 19, 24, 21, 26, 23, 28, 25, 30},
			},
			{
				Label:  "Latency P99",
				Value:  "42",
				Unit:   "ms",
				Points: []float64{45, 38, 52, 41, 37, 48, 42, 39, 44, 36, 42, 38},
			},
			{
				Label:  "Memory",
				Value:  "128",
				Unit:   "MiB",
				Points: []float64{120, 122, 125, 124, 128, 126, 130, 127, 125, 128, 126, 129},
			},
		},
	}
}
