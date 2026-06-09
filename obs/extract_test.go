package obs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── MetricsSnapshot builder tests ──

func TestNewSnapshot(t *testing.T) {
	now := time.Now()

	labels := func(ns, deploy, node string) map[string]string {
		return map[string]string{
			"k8s_namespace_name":  ns,
			"k8s.deployment.name": deploy,
			"k8s_node_name":       node,
		}
	}
	nodeLabels := func(node string) map[string]string {
		return map[string]string{"k8s.node.name": node}
	}

	t.Run("extracts workload metrics", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "0.1"}, {Timestamp: 2, Value: "0.3"}}},
					{Labels: labels("platform-website", "platform-website", "node-b"),
						Values: []v3Point{{Timestamp: 1, Value: "0.11"}, {Timestamp: 2, Value: "0.12"}}},
				}},
				{QueryName: queryRAM, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "110100480"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		wm, ok := snap.Workloads["flux-system/source-controller"]
		require.True(t, ok)
		assert.InDelta(t, 0.3, wm.CPU, 0.001, "latest CPU from last point")
		assert.InDelta(t, 110100480, wm.RAM, 1, "RAM in bytes")
		assert.Equal(t, "node-a", wm.Host)
	})

	t.Run("extracts node metrics", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "0.1"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU, Series: []v3Series{
					{Labels: nodeLabels("node-a"), Values: []v3Point{{Timestamp: 1, Value: "0.5"}}},
					{Labels: nodeLabels("node-b"), Values: []v3Point{{Timestamp: 1, Value: "1.2"}}},
				}},
				{QueryName: queryNodeRAM, Series: []v3Series{
					{Labels: nodeLabels("node-a"), Values: []v3Point{{Timestamp: 1, Value: "4294967296"}}},
					{Labels: nodeLabels("node-b"), Values: []v3Point{{Timestamp: 1, Value: "8589934592"}}},
				}},
				{QueryName: queryNodeUp, Series: []v3Series{
					{Labels: nodeLabels("node-a"), Values: []v3Point{{Timestamp: 1, Value: "536077"}}},
				}},
			}},
		}
		snap := newSnapshot(resp, now)
		assert.Equal(t, 2, len(snap.Nodes))
		assert.InDelta(t, 0.5, snap.Nodes["node-a"].CPU, 0.001)
		assert.InDelta(t, 4294967296, snap.Nodes["node-a"].RAM, 1)
		assert.InDelta(t, 536077, snap.Nodes["node-a"].Uptime, 1)
	})

	t.Run("counts services per node", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("ns1", "svc1", "node-a"), Values: []v3Point{{Timestamp: 1, Value: "0.1"}}},
					{Labels: labels("ns2", "svc2", "node-a"), Values: []v3Point{{Timestamp: 1, Value: "0.2"}}},
					{Labels: labels("ns3", "svc3", "node-b"), Values: []v3Point{{Timestamp: 1, Value: "0.3"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		assert.Equal(t, 2, snap.NodeSvcCounts["node-a"])
		assert.Equal(t, 1, snap.NodeSvcCounts["node-b"])
	})

	t.Run("takes max across pods of same deployment", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "0.1"}, {Timestamp: 2, Value: "0.3"}}},
					{Labels: labels("flux-system", "source-controller", "node-b"),
						Values: []v3Point{{Timestamp: 1, Value: "0.2"}, {Timestamp: 2, Value: "0.5"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		wm := snap.Workloads["flux-system/source-controller"]
		assert.InDelta(t, 0.5, wm.CPU, 0.001)
	})

	t.Run("clamps CPU > 100 to zero", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "150"}}},
				}},
				{QueryName: queryRAM, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "110100480"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		wm := snap.Workloads["flux-system/source-controller"]
		assert.Equal(t, 0.0, wm.CPU, "corrupted CPU >100 should be clamped")
	})

	t.Run("generates sparkline history", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{
							{Timestamp: 1, Value: "0.1"},
							{Timestamp: 2, Value: "0.3"},
							{Timestamp: 3, Value: "0.2"},
							{Timestamp: 4, Value: "0.4"},
						}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		wm := snap.Workloads["flux-system/source-controller"]
		assert.Equal(t, 4, len(wm.CPUHist))
		assert.InDelta(t, 0.1, wm.CPUHist[0], 0.001)
		assert.InDelta(t, 0.4, wm.CPUHist[3], 0.001)
	})

	t.Run("skips entries with missing namespace", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: map[string]string{"k8s.deployment.name": "myapp"},
						Values: []v3Point{{Timestamp: 1, Value: "0.5"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		assert.Empty(t, snap.Workloads)
	})

	t.Run("falls back to statefulset name", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: map[string]string{
						"k8s_namespace_name":   "signoz",
						"k8s.statefulset.name": "clickhouse",
						"k8s_node_name":        "node-a",
					}, Values: []v3Point{{Timestamp: 1, Value: "3.5"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		wm, ok := snap.Workloads["signoz/clickhouse"]
		require.True(t, ok)
		assert.InDelta(t, 3.5, wm.CPU, 0.001)
	})

	t.Run("duplicates single point for sparkline", func(t *testing.T) {
		resp := &v3Response{
			Status: "success",
			Data: v3DataRoot{Result: []v3Result{
				{QueryName: queryCPU, Series: []v3Series{
					{Labels: labels("flux-system", "source-controller", "node-a"),
						Values: []v3Point{{Timestamp: 1, Value: "0.5"}}},
				}},
				{QueryName: queryDisk},
				{QueryName: queryNet},
				{QueryName: queryRAM},
				{QueryName: queryNodeCPU},
				{QueryName: queryNodeRAM},
				{QueryName: queryNodeUp},
			}},
		}
		snap := newSnapshot(resp, now)
		wm := snap.Workloads["flux-system/source-controller"]
		assert.Equal(t, 2, len(wm.CPUHist), "single value should be duplicated for sparkline")
		assert.InDelta(t, 0.5, wm.CPUHist[0], 0.001)
	})
}

func TestLatestByWorkload(t *testing.T) {
	labels := func(ns, deploy, node string) map[string]string {
		return map[string]string{
			"k8s_namespace_name":  ns,
			"k8s.deployment.name": deploy,
			"k8s_node_name":       node,
		}
	}

	t.Run("takes max across pods of same deployment", func(t *testing.T) {
		series := []v3Series{
			{Labels: labels("flux-system", "source-controller", "node-a"),
				Values: []v3Point{{Timestamp: 1, Value: "0.1"}, {Timestamp: 2, Value: "0.3"}}},
			{Labels: labels("flux-system", "source-controller", "node-b"),
				Values: []v3Point{{Timestamp: 1, Value: "0.2"}, {Timestamp: 2, Value: "0.5"}}},
		}
		result := latestByWorkload(series)
		assert.InDelta(t, 0.5, result["flux-system/source-controller"], 0.001)
	})

	t.Run("skips entries with missing deployment name", func(t *testing.T) {
		series := []v3Series{
			{Labels: map[string]string{"k8s_namespace_name": "flux-system"},
				Values: []v3Point{{Timestamp: 1, Value: "0.5"}}},
		}
		result := latestByWorkload(series)
		assert.Empty(t, result)
	})

	t.Run("skips empty series", func(t *testing.T) {
		series := []v3Series{
			{Labels: labels("flux-system", "source-controller", "node-a"), Values: []v3Point{}},
		}
		result := latestByWorkload(series)
		assert.Empty(t, result)
	})

	t.Run("skips unparseable values", func(t *testing.T) {
		series := []v3Series{
			{Labels: labels("flux-system", "source-controller", "node-a"),
				Values: []v3Point{{Timestamp: 1, Value: "notanumber"}}},
		}
		result := latestByWorkload(series)
		assert.Empty(t, result)
	})
}

func TestLatestByNode(t *testing.T) {
	t.Run("extracts latest value per node", func(t *testing.T) {
		series := []v3Series{
			{Labels: map[string]string{"k8s.node.name": "node-a"},
				Values: []v3Point{{Timestamp: 1, Value: "0.5"}, {Timestamp: 2, Value: "0.6"}}},
			{Labels: map[string]string{"k8s.node.name": "node-b"},
				Values: []v3Point{{Timestamp: 1, Value: "1.2"}, {Timestamp: 2, Value: "1.3"}}},
		}
		result := latestByNode(series)
		assert.InDelta(t, 0.6, result["node-a"], 0.001)
		assert.InDelta(t, 1.3, result["node-b"], 0.001)
	})

	t.Run("skips missing node name", func(t *testing.T) {
		series := []v3Series{
			{Labels: map[string]string{}, Values: []v3Point{{Timestamp: 1, Value: "0.5"}}},
		}
		result := latestByNode(series)
		assert.Empty(t, result)
	})
}

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
				"signoz/alertmanager": {
					Namespace: "signoz", Name: "alertmanager", Host: "node-a", CPU: 0.1,
				},
				"flux-system/source-controller": {
					Namespace: "flux-system", Name: "source-controller", Host: "node-a", CPU: 0.2,
				},
			},
		}
		services := BuildServices(snap, time.Now())
		require.Len(t, services, 2)
		// deployment (flux-system) comes before observability (signoz)
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
}

func TestBuildHosts(t *testing.T) {
	t.Run("labels control-plane nodes as Cloud", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"talosoci-control-plane-abc": {CPU: 0.5, RAM: 4294967296},
			},
		}
		hosts := BuildHosts(snap, nil)
		require.Len(t, hosts, 1)
		assert.Equal(t, "Cloud", hosts[0].Label)
		assert.Equal(t, "Control plane", hosts[0].Detail)
	})

	t.Run("labels worker nodes as Edge", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"talosedge-xyz": {CPU: 1.2, RAM: 8589934592},
			},
		}
		hosts := BuildHosts(snap, nil)
		require.Len(t, hosts, 1)
		assert.Equal(t, "Edge", hosts[0].Label)
		assert.Equal(t, "Worker node", hosts[0].Detail)
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
}

func TestBuildHosts_WithNodeInfo(t *testing.T) {
	info := func(name string) (NodeInfo, bool) {
		return NodeInfo{
			IsControlPlane: name == "talos-oci-c-ultimate-parakeet",
			Provider:       "OCI Cloud",
			Arch:           "ARM64",
		}, true
	}

	t.Run("uses k8s node info for labeling", func(t *testing.T) {
		snap := MetricsSnapshot{
			Nodes: map[string]NodeMetrics{
				"talos-oci-c-ultimate-parakeet": {CPU: 0.5},
				"talos-os-w-logical-mule":       {CPU: 1.2},
			},
		}
		hosts := BuildHosts(snap, info)
		require.Len(t, hosts, 2)

		// Control plane sorts first
		assert.Equal(t, "talos-oci-c-ultimate-parakeet", hosts[0].Name)
		assert.Equal(t, "OCI Cloud", hosts[0].Label)
		assert.Contains(t, hosts[0].Detail, "Control plane")
		assert.Contains(t, hosts[0].Detail, "ARM64")

		assert.Equal(t, "talos-os-w-logical-mule", hosts[1].Name)
		assert.Equal(t, "OCI Cloud", hosts[1].Label)
		assert.Contains(t, hosts[1].Detail, "Worker node")
	})
}
