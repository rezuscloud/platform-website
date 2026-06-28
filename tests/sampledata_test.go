package tests

import "github.com/rezuscloud/platform-website/obs"

// sampleLiveData is a realistic, current-cluster snapshot used ONLY by the
// integration/e2e tests to verify the live section's rendering structure. It is
// not the production fallback (that is obs.DefaultMockData, which is
// deliberately empty so the site never fabricates a cluster on failure).
//
// Node names match the real cluster at time of writing so test fixtures do not
// drift into obvious fiction, but this is test data, not a source of truth.
func sampleLiveData() obs.LiveData {
	return obs.LiveData{
		SelfNamespace: "platform-website",
		HasMetrics:    false,
		Hosts: []obs.Host{
			{Name: "talos-oci-c-beloved-kite", Label: "Control plane", Detail: "OCI Cloud \u00b7 ARM64", CPU: 0.6, CPUCores: 4, RAM: 4300, RAMTotal: 23950, Uptime: "8d", SvcCount: 9},
			{Name: "talos-edge-w-frank-antelope", Label: "Worker", Detail: "Edge \u00b7 AMD64", CPU: 4.8, CPUCores: 16, RAM: 13814, RAMTotal: 48038, Uptime: "8d", SvcCount: 21},
			{Name: "talos-edge-w-next-bug", Label: "Worker", Detail: "Edge \u00b7 AMD64", CPU: 0.2, CPUCores: 64, RAM: 3051, RAMTotal: 257800, Uptime: "2m", SvcCount: 4},
		},
		Services: []obs.Service{
			{Name: "source-controller", Namespace: "flux-system", Category: "deployment", Host: "talos-oci-c-beloved-kite", Status: "healthy", CPU: 0.44, RAM: 118, Uptime: "8d"},
			{Name: "kube-prometheus-stack-operator", Namespace: "monitoring", Category: "deployment", Host: "talos-edge-w-frank-antelope", Status: "healthy", CPU: 0.29, RAM: 111, Uptime: "12h"},
			{Name: "platform-website", Namespace: "platform-website", Category: "runtime", Host: "talos-edge-w-frank-antelope", Status: "healthy", CPU: 0.11, RAM: 121, Uptime: "22h"},
			{Name: "dapr-operator", Namespace: "dapr-system", Category: "runtime", Host: "talos-edge-w-frank-antelope", Status: "healthy", CPU: 0.13, RAM: 53, Uptime: "3d"},
			{Name: "cilium-operator", Namespace: "kube-system", Category: "runtime", Host: "talos-edge-w-frank-antelope", Status: "healthy", Uptime: "8d"},
			{Name: "prometheus", Namespace: "monitoring", Category: "observability", Host: "talos-edge-w-frank-antelope", Status: "healthy", Uptime: "6h"},
			{Name: "cert-manager", Namespace: "cert-manager", Category: "deployment", Host: "talos-edge-w-frank-antelope", Status: "healthy", CPU: 0.06, RAM: 73, Uptime: "1d"},
			{Name: "juicefs-tikv-pd", Namespace: "tikv-system", Category: "data", Host: "talos-edge-w-frank-antelope", Status: "healthy", CPU: 3.82, RAM: 173, Uptime: "8d"},
			{Name: "forgejo-runner", Namespace: "forgejo", Category: "dev", Host: "talos-edge-w-next-bug", Status: "healthy", CPU: 0.05, RAM: 42},
			{Name: "actions-runner", Namespace: "arc-systems", Category: "dev", Host: "talos-edge-w-next-bug", Status: "healthy", CPU: 0.18, RAM: 88, Uptime: "1d"},
		},
	}
}
