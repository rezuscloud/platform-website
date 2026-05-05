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
				Name:   "talosoci-control-plane-legal-poodle",
				Tier:   "oci-cloud",
				Status: "Ready",
				CPU:    "12%",
				Mem:    "4.2 GiB",
				Pods: []Pod{
					{Name: "kube-apiserver-talosoci-...", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "cilium-bnj7n", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "coredns-7859998f6-bjrj2", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "source-controller-5f5f984f54", Namespace: "flux-system", Status: "Running", Restarts: 0},
					{Name: "kustomize-controller-86c4c6f9f7", Namespace: "flux-system", Status: "Running", Restarts: 0},
					{Name: "forgejo-76bb8587c6-r4bbc", Namespace: "forgejo", Status: "Running", Restarts: 0},
				},
			},
			{
				Name:   "talosedge-genmachiche-flowing-bluejay",
				Tier:   "edge",
				Status: "Ready",
				CPU:    "34%",
				Mem:    "5.8 GiB",
				Pods: []Pod{
					{Name: "platform-website-69f7bffd5f-trltm", Namespace: "platform-website", Status: "Running", Restarts: 0},
					{Name: "signoz-0", Namespace: "signoz", Status: "Running", Restarts: 0},
					{Name: "signoz-otel-collector-86659589f", Namespace: "signoz", Status: "Running", Restarts: 0},
					{Name: "chi-signoz-clickhouse-0-0-0", Namespace: "signoz", Status: "Running", Restarts: 0},
					{Name: "opencloud-7b45456dc9-bk48r", Namespace: "opencloud", Status: "Running", Restarts: 0},
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
