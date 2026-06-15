package obs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildServices(t *testing.T) {
	t.Run("filters services with all-zero metrics", func(t *testing.T) {
		snap := MetricsSnapshot{
			Workloads: map[string]WorkloadMetrics{
				"flux-system/source-controller": {
					Namespace: "flux-system", Name: "source-controller", Host: "node-a",
					CPU: 0.1, RAM: 110100480,
				},
				"kube-system/some-dormant-svc": {
					Namespace: "kube-system", Name: "some-dormant-svc", Host: "node-a",
					CPU: 0, RAM: 0, Net: 0, Disk: 0,
				},
			},
		}
		services := BuildServices(snap, time.Now())
		names := make(map[string]bool)
		for _, s := range services {
			names[s.Name] = true
		}
		assert.True(t, names["source-controller"], "service with metrics should be included")
		assert.False(t, names["some-dormant-svc"], "service with all-zero metrics should be filtered")
	})

	t.Run("sorts by category order then name", func(t *testing.T) {
		snap := MetricsSnapshot{
			Workloads: map[string]WorkloadMetrics{
				"monitoring/alertmanager": {
					Namespace: "monitoring", Name: "alertmanager", Host: "node-a", CPU: 0.1,
				},
				"flux-system/source-controller": {
					Namespace: "flux-system", Name: "source-controller", Host: "node-a", CPU: 0.2,
				},
			},
		}
		services := BuildServices(snap, time.Now())
		require.Len(t, services, 2)
		// deployment (flux-system) comes before observability (monitoring)
		assert.Equal(t, "source-controller", services[0].Name)
		assert.Equal(t, "alertmanager", services[1].Name)
	})

	t.Run("converts RAM bytes to MB", func(t *testing.T) {
		snap := MetricsSnapshot{
			Workloads: map[string]WorkloadMetrics{
				"flux-system/source-controller": {
					Namespace: "flux-system", Name: "source-controller",
					RAM: 110100480, // ~105 MB
				},
			},
		}
		services := BuildServices(snap, time.Now())
		require.Len(t, services, 1)
		assert.InDelta(t, 105.0, services[0].RAM, 1.0)
	})

	t.Run("emits sparkline history points", func(t *testing.T) {
		snap := MetricsSnapshot{
			Workloads: map[string]WorkloadMetrics{
				"flux-system/source-controller": {
					Namespace: "flux-system", Name: "source-controller", CPU: 0.3,
					CPUHist: []float64{0.1, 0.3, 0.2, 0.4},
				},
			},
		}
		services := BuildServices(snap, time.Now())
		require.Len(t, services, 1)
		assert.NotEmpty(t, services[0].CPUHist)
	})
}

func TestBuildHosts(t *testing.T) {
	t.Run("hostname fallback labels control-plane nodes", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"talos-oci-c-control-plane-abc": {CPU: 0.5, RAM: 4294967296},
			},
		}
		hosts := BuildHosts(snap, nil)
		require.Len(t, hosts, 1)
		assert.Equal(t, "Control plane", hosts[0].Label)
	})

	t.Run("hostname fallback labels other nodes as Worker", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"talosedge-xyz": {CPU: 1.2, RAM: 8589934592},
			},
		}
		hosts := BuildHosts(snap, nil)
		require.Len(t, hosts, 1)
		assert.Equal(t, "Worker", hosts[0].Label)
	})

	t.Run("computes uptime from node metrics", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"node-a": {Uptime: 536077}, // ~6.2 days
			},
		}
		hosts := BuildHosts(snap, nil)
		require.Len(t, hosts, 1)
		assert.Equal(t, "6d", hosts[0].Uptime)
	})

	t.Run("includes service count per node", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"node-a": {CPU: 0.5},
				"node-b": {CPU: 1.0},
			},
			NodeSvcCounts: map[string]int{"node-a": 3, "node-b": 1},
		}
		hosts := BuildHosts(snap, nil)
		hostMap := make(map[string]Host)
		for _, h := range hosts {
			hostMap[h.Name] = h
		}
		assert.Equal(t, 3, hostMap["node-a"].SvcCount)
		assert.Equal(t, 1, hostMap["node-b"].SvcCount)
	})

	t.Run("control plane sorts before worker", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"worker-z":  {CPU: 1},
				"control-1": {CPU: 1},
			},
		}
		hosts := BuildHosts(snap, nil)
		require.Len(t, hosts, 2)
		assert.Equal(t, "control-1", hosts[0].Name)
	})
}

func TestBuildHosts_WithTopology(t *testing.T) {
	info := func(name string) (NodeInfo, bool) {
		switch name {
		case "talos-oci-c-rapid-gator":
			return NodeInfo{
				IsControlPlane: true,
				Provider:       "OCI Cloud",
				Arch:           "ARM64",
				CPUCores:       4,
				RAMBytes:       25116016640,
			}, true
		case "talos-os-w-lasting-phoenix":
			return NodeInfo{
				IsControlPlane: false,
				Provider:       "Cloud",
				Arch:           "AMD64",
				CPUCores:       16,
				RAMBytes:       33629507584,
			}, true
		}
		return NodeInfo{}, false
	}

	t.Run("uses k8s topology for role, provider, arch, capacity", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"talos-oci-c-rapid-gator":    {CPU: 0.5, RAM: 4.3e9},
				"talos-os-w-lasting-phoenix": {CPU: 1.2, RAM: 3.0e9},
			},
		}
		hosts := BuildHosts(snap, info)
		require.Len(t, hosts, 2)

		// Control plane sorts first.
		assert.Equal(t, "talos-oci-c-rapid-gator", hosts[0].Name)
		assert.Equal(t, "Control plane", hosts[0].Label)
		assert.Contains(t, hosts[0].Detail, "OCI Cloud")
		assert.Contains(t, hosts[0].Detail, "ARM64")
		assert.InDelta(t, 4, hosts[0].CPUCores, 0.001)

		assert.Equal(t, "talos-os-w-lasting-phoenix", hosts[1].Name)
		assert.Equal(t, "Worker", hosts[1].Label)
		assert.Contains(t, hosts[1].Detail, "Cloud")
		assert.InDelta(t, 16, hosts[1].CPUCores, 0.001)
	})
}

// ── K8s topology parsing helpers ──

func TestParseCores(t *testing.T) {
	assert.InDelta(t, 4, parseCores("4"), 0.001)
	assert.InDelta(t, 1.6, parseCores("1600m"), 0.001)
	assert.InDelta(t, 0, parseCores(""), 0.001)
	assert.InDelta(t, 0, parseCores("nope"), 0.001)
}

func TestParseBytes(t *testing.T) {
	assert.InDelta(t, 25116016640, parseBytes("25116016640"), 1)
	assert.InDelta(t, 25165824000, parseBytes("24576000Ki"), 1)
	assert.InDelta(t, 0, parseBytes(""), 0)
}

func TestProviderFromID(t *testing.T) {
	assert.Equal(t, "OCI Cloud", providerFromID("ocid1.instance.oc1.phx.abcd", ""))
	assert.Equal(t, "Edge", providerFromID("metal://edge-node", ""))
	assert.Equal(t, "OCI Cloud", providerFromID("", "talos-oci-c-foo"))
	assert.Equal(t, "Edge", providerFromID("", "talosedge-bar"))
	assert.Equal(t, "Cloud", providerFromID("", "talos-os-w-baz"))
	assert.Equal(t, "", providerFromID("", "random-name"))
}

func TestArchLabel(t *testing.T) {
	assert.Equal(t, "ARM64", archLabel("arm64"))
	assert.Equal(t, "ARM64", archLabel("aarch64"))
	assert.Equal(t, "AMD64", archLabel("amd64"))
	assert.Equal(t, "AMD64", archLabel("x86_64"))
}

func TestResolveWorkload(t *testing.T) {
	rs := map[string]string{"default/myapp-abc": "myapp"}

	t.Run("replicaset resolves to deployment via rs map", func(t *testing.T) {
		got := resolveWorkload([]ownerRef{{Kind: "ReplicaSet", Name: "myapp-abc"}}, "default", rs)
		assert.Equal(t, "myapp", got)
	})
	t.Run("statefulset owner is the workload name", func(t *testing.T) {
		got := resolveWorkload([]ownerRef{{Kind: "StatefulSet", Name: "clickhouse"}}, "signoz", nil)
		assert.Equal(t, "clickhouse", got)
	})
	t.Run("bare pod returns empty", func(t *testing.T) {
		assert.Equal(t, "", resolveWorkload(nil, "default", nil))
	})
}
