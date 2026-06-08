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

// DefaultMockData returns a static snapshot showing the node strip layout.
func DefaultMockData() LiveData {
	now := time.Now().Unix()
	return LiveData{
		SelfNamespace: "platform-website",
		Hosts: []Host{
			{Name: "talosoci-control-plane-intent-moth", Label: "OCI Cloud", Detail: "ARM64 \u00b7 Ampere A1", CPU: 0.58, RAM: 5519.9, LoadAvg: 1.17, IOWait: 3.4, Uptime: "8d", SvcCount: 9},
			{Name: "talosedge-genmachiche-keen-lioness", Label: "Edge", Detail: "AMD64 \u00b7 Intel NUC", CPU: 4.77, RAM: 14589.8, LoadAvg: 7.29, IOWait: 2.2, Uptime: "8d", SvcCount: 21},
			{Name: "talos-os-w-good-weasel", Label: "Cloud", Detail: "AMD64 \u00b7 Proxmox VM", CPU: 0.22, RAM: 3200.4, LoadAvg: 0.41, IOWait: 1.1, Uptime: "2m", SvcCount: 4},
		},
		Services: []Service{
			{Name: "arc-controller-gha-rs-controller", Namespace: "arc-systems", Category: "dev", Host: "talosoci-control-plane-intent-moth", Status: "healthy", CPU: 0.1, RAM: 64},
			{Name: "forgejo-valkey-primary", Namespace: "forgejo", Category: "dev", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 0.0, RAM: 14},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "deployment", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "external-dns", Namespace: "external-dns", Category: "deployment", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 0.10, RAM: 94, Uptime: "1d"},
			{Name: "source-controller", Namespace: "flux-system", Category: "deployment", Host: "talosoci-control-plane-intent-moth", Status: "healthy", CPU: 0.44, RAM: 118, Uptime: "8d"},
			{Name: "kubevela-vela-core", Namespace: "vela-system", Category: "deployment", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "3d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", Uptime: "8d"},
			{Name: "chi-signoz-clickhouse-cluster-0-0", Namespace: "signoz", Category: "observability", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", Uptime: "6h"},
			{Name: "juicefs-csi-controller", Namespace: "juicefs-csi", Category: "data", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", Uptime: "1d"},
			{Name: "juicefs-tikv-pd", Namespace: "tikv-system", Category: "data", Host: "talosedge-genmachiche-keen-lioness", Status: "healthy", CPU: 3.82, RAM: 173, Uptime: "8d"},
			{Name: "actions-runner", Namespace: "arc-systems", Category: "dev", Host: "talos-os-w-good-weasel", Status: "healthy", CPU: 0.05, RAM: 42},
			{Name: "flux-kustomization-controller", Namespace: "flux-system", Category: "deployment", Host: "talos-os-w-good-weasel", Status: "healthy", CPU: 0.18, RAM: 88},
			{Name: "coredns", Namespace: "kube-system", Category: "runtime", Host: "talos-os-w-good-weasel", Status: "healthy", CPU: 0.02, RAM: 28},
		},
		HasMetrics: false,
		Timestamp:  now,
	}
}
