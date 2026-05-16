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

// DefaultMockData returns a static snapshot showing the matrix layout.
func DefaultMockData() LiveData {
	now := time.Now().Unix()
	return LiveData{
		SelfNamespace: "platform-website",
		Hosts: []Host{
			{Name: "talosoci-control-plane-legal-poodle", Label: "OCI Cloud", Detail: "ARM64 \u00b7 Ampere A1", CPU: 0.58, RAM: 5519.9, LoadAvg: 1.17, IOWait: 3.4, Uptime: "5d", SvcCount: 9},
			{Name: "talosedge-genmachiche-flowing-bluejay", Label: "Edge Node", Detail: "AMD64 \u00b7 Intel NUC", CPU: 4.77, RAM: 14589.8, LoadAvg: 7.29, IOWait: 2.2, Uptime: "5d", SvcCount: 21},
		},
		Services: []Service{
			{Name: "arc-controller-gha-rs-controller", Namespace: "arc-systems", Category: "dev", Host: "talosoci-control-plane-legal-poodle", Status: "healthy", CPU: 0.1, RAM: 64},
			{Name: "forgejo-valkey-primary", Namespace: "forgejo", Category: "dev", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 0.0, RAM: 14},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "deployment", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "external-dns", Namespace: "external-dns", Category: "deployment", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 0.10, RAM: 94, Uptime: "1d"},
			{Name: "source-controller", Namespace: "flux-system", Category: "deployment", Host: "talosoci-control-plane-legal-poodle", Status: "healthy", CPU: 0.44, RAM: 118, Uptime: "5d"},
			{Name: "kubevela-vela-core", Namespace: "vela-system", Category: "deployment", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "3d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", Uptime: "5d"},
			{Name: "chi-signoz-clickhouse-cluster-0-0", Namespace: "signoz", Category: "observability", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", Uptime: "6h"},
			{Name: "juicefs-csi-controller", Namespace: "juicefs-csi", Category: "data", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", Uptime: "1d"},
			{Name: "juicefs-tikv-pd", Namespace: "tikv-system", Category: "data", Host: "talosedge-genmachiche-flowing-bluejay", Status: "healthy", CPU: 3.82, RAM: 173, Uptime: "5d"},
		},
		HasMetrics: false,
		Timestamp:  now,
	}
}
