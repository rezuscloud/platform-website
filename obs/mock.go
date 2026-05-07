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

// DefaultMockData returns a static snapshot showing the service map layout.
func DefaultMockData() LiveData {
	now := time.Now().Unix()
	return LiveData{
		Services: []Service{
			{Name: "talosoci-control-plane-legal-poodle", Category: "hosts", Status: "running", Detail: "ARM64 \u00b7 Ampere A1 \u00b7 9 svcs", CPU: 0.58, RAM: 5519.9, LoadAvg: 1.17, IOWait: 3.4, Uptime: "5d"},
			{Name: "talosedge-genmachiche-flowing-bluejay", Category: "hosts", Status: "running", Detail: "AMD64 \u00b7 Intel NUC \u00b7 21 svcs", CPU: 4.77, RAM: 14589.8, LoadAvg: 7.29, IOWait: 2.2, Uptime: "5d"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "3d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Status: "healthy", CPU: 0, RAM: 0, Uptime: "5d"},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "deployment", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "external-dns", Namespace: "external-dns", Category: "deployment", Status: "healthy", CPU: 0.10, RAM: 94, Uptime: "1d"},
			{Name: "source-controller", Namespace: "flux-system", Category: "deployment", Status: "healthy", CPU: 0.44, RAM: 118, Uptime: "5d"},
			{Name: "kubevela-vela-core", Namespace: "vela-system", Category: "deployment", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h"},
			{Name: "chi-signoz-clickhouse-cluster-0-0", Namespace: "signoz", Category: "observability", Status: "healthy", CPU: 0, RAM: 0, Uptime: "6h"},
			{Name: "juicefs-csi-controller", Namespace: "juicefs-csi", Category: "data", Status: "healthy", CPU: 0, RAM: 0, Uptime: "1d"},
			{Name: "juicefs-tikv-pd", Namespace: "tikv-system", Category: "data", Status: "healthy", CPU: 3.82, RAM: 173, Uptime: "5d"},
		},
		HasMetrics: false,
		Timestamp:  now,
	}
}
