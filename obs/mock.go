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
		Services: append(StaticHosts(), []Service{
			{Name: "forgejo", Namespace: "forgejo", Category: "dev", Status: "healthy", CPU: 0.03, RAM: 16, Uptime: "1h"},
			{Name: "source-controller", Namespace: "flux-system", Category: "deployment", Status: "healthy", CPU: 0.25, RAM: 105, Uptime: "4d"},
			{Name: "kustomize-controller", Namespace: "flux-system", Category: "deployment", Status: "healthy", CPU: 0.08, RAM: 79, Uptime: "4d"},
			{Name: "helm-controller", Namespace: "flux-system", Category: "deployment", Status: "healthy", CPU: 0.15, RAM: 96, Uptime: "4d"},
			{Name: "kubevela-vela-core", Namespace: "vela-system", Category: "deployment", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h"},
			{Name: "external-dns", Namespace: "external-dns", Category: "deployment", Status: "healthy", CPU: 0.10, RAM: 94, Uptime: "1d"},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "deployment", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Status: "healthy", CPU: 0.08, RAM: 45, Uptime: "4d"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "12h"},
			{Name: "signoz-otel-collector", Namespace: "signoz", Category: "observability", Status: "unmonitored"},
			{Name: "juicefs-tikv-pd", Namespace: "tikv-system", Category: "data", Status: "healthy", CPU: 3.82, RAM: 173, Uptime: "4d"},
		}...),
		HasMetrics: false,
		Timestamp:  now,
	}
}
