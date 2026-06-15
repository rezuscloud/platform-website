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
			{Name: "talos-oci-c-rapid-gator", Label: "Control plane", Detail: "OCI Cloud \u00b7 ARM64", CPU: 0.58, CPUCores: 4, RAM: 4.3, RAMTotal: 23950, Uptime: "8d", SvcCount: 9},
			{Name: "talos-edge-w-known-foal", Label: "Worker", Detail: "Edge \u00b7 AMD64", CPU: 4.77, CPUCores: 16, RAM: 13814, RAMTotal: 48038, Uptime: "8d", SvcCount: 21},
			{Name: "talos-os-w-lasting-phoenix", Label: "Worker", Detail: "Cloud \u00b7 AMD64", CPU: 0.22, CPUCores: 16, RAM: 3051, RAMTotal: 32068, Uptime: "2m", SvcCount: 4},
		},
		Services: []Service{
			{Name: "arc-controller-gha-rs-controller", Namespace: "arc-systems", Category: "dev", Host: "talos-oci-c-rapid-gator", Status: "healthy", CPU: 0.1, RAM: 64},
			{Name: "forgejo-valkey-primary", Namespace: "forgejo", Category: "dev", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 0.0, RAM: 14},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "deployment", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "external-dns", Namespace: "external-dns", Category: "deployment", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 0.10, RAM: 94, Uptime: "1d"},
			{Name: "source-controller", Namespace: "flux-system", Category: "deployment", Host: "talos-oci-c-rapid-gator", Status: "healthy", CPU: 0.44, RAM: 118, Uptime: "8d"},
			{Name: "kube-prometheus-stack-operator", Namespace: "monitoring", Category: "deployment", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "3d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Host: "talos-edge-w-known-foal", Status: "healthy", Uptime: "8d"},
			{Name: "prometheus", Namespace: "monitoring", Category: "observability", Host: "talos-edge-w-known-foal", Status: "healthy", Uptime: "6h"},
			{Name: "juicefs-csi-controller", Namespace: "juicefs-csi", Category: "data", Host: "talos-edge-w-known-foal", Status: "healthy", Uptime: "1d"},
			{Name: "juicefs-tikv-pd", Namespace: "tikv-system", Category: "data", Host: "talos-edge-w-known-foal", Status: "healthy", CPU: 3.82, RAM: 173, Uptime: "8d"},
			{Name: "actions-runner", Namespace: "arc-systems", Category: "dev", Host: "talos-os-w-lasting-phoenix", Status: "healthy", CPU: 0.05, RAM: 42},
			{Name: "flux-kustomization-controller", Namespace: "flux-system", Category: "deployment", Host: "talos-os-w-lasting-phoenix", Status: "healthy", CPU: 0.18, RAM: 88, Uptime: "1d"},
			{Name: "coredns", Namespace: "kube-system", Category: "runtime", Host: "talos-os-w-lasting-phoenix", Status: "healthy", CPU: 0.02, RAM: 28},
		},
		HasMetrics: false,
		Timestamp:  now,
	}
}
