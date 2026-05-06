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

// DefaultMockData returns a static snapshot showing the dashboard layout.
func DefaultMockData() LiveData {
	now := time.Now().Unix()
	return LiveData{
		Services: []Service{
			{Name: "forgejo", Namespace: "forgejo", Category: "dev", Status: "healthy", CPU: 0.03, RAM: 16, Uptime: "1h", CPUHist: "0,16 4,15.5 8,16 12,15.8 16,16 20,15.6 24,16 28,15.9 32,16 36,15.7 40,16 44,15.8", RAMHist: "0,8 4,8 8,7.5 12,8 16,8.2 20,8 24,7.8 28,8 32,8 36,8.1 40,8 44,8"},
			{Name: "source-controller", Namespace: "flux-system", Category: "delivery", Status: "healthy", CPU: 0.25, RAM: 105, Uptime: "4d", CPUHist: "0,12 4,11 8,13 12,10 16,12 20,11 24,13 28,10 32,12 36,11 40,12 44,10", RAMHist: "0,6 4,6.5 8,7 12,6 16,6.5 20,7 24,6 28,6.5 32,7 36,6 40,6.5 44,7"},
			{Name: "kustomize-controller", Namespace: "flux-system", Category: "delivery", Status: "healthy", CPU: 0.08, RAM: 79, Uptime: "4d", CPUHist: "0,14 4,14 8,14 12,14 16,14 20,14 24,14 28,14 32,14 36,14 40,14 44,14", RAMHist: "0,10 4,10 8,10 12,10 16,10 20,10 24,10 28,10 32,10 36,10 40,10 44,10"},
			{Name: "helm-controller", Namespace: "flux-system", Category: "delivery", Status: "healthy", CPU: 0.15, RAM: 96, Uptime: "4d", CPUHist: "0,13 4,12 8,13 12,11 16,12 20,13 24,12 28,11 32,12 36,13 40,12 44,12", RAMHist: "0,7 4,7.5 8,8 12,7 16,7.5 20,8 24,7 28,7.5 32,8 36,7 40,7.5 44,8"},
			{Name: "kubevela-vela-core", Namespace: "vela-system", Category: "delivery", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h", CPUHist: "0,10 4,11 8,9 12,10 16,11 20,10 24,9 28,10 32,11 36,10 40,9 44,10", RAMHist: "0,5 4,5.5 8,5 12,5.5 16,5 20,5.5 24,5 28,5.5 32,5 36,5.5 40,5 44,5.5"},
			{Name: "external-dns", Namespace: "external-dns", Category: "delivery", Status: "healthy", CPU: 0.10, RAM: 94, Uptime: "1d"},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "delivery", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Status: "healthy", CPU: 0.08, RAM: 45, Uptime: "4d"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h", CPUHist: "0,15 4,14.5 8,15 12,14.8 16,15 20,14.6 24,15 28,14.9 32,15 36,14.7 40,15 44,14.8", RAMHist: "0,4 4,4.2 8,4 12,4.1 16,4.3 20,4 24,4.2 28,4 32,4.1 36,4.3 40,4 44,4.2"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "12h"},
			{Name: "signoz-otel-collector", Namespace: "signoz", Category: "observability", Status: "unmonitored"},
		},
		HasMetrics: false,
		Timestamp:  now,
	}
}
