package obs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtureSeries builds a v3Series from labels and string values.
func fixtureSeries(labels map[string]string, values ...string) v3Series {
	pts := make([]v3Point, len(values))
	for i, v := range values {
		pts[i] = v3Point{Timestamp: int64(i) * 300000, Value: v}
	}
	return v3Series{Labels: labels, Values: pts}
}

func TestLatestByDeployment(t *testing.T) {
	labels := func(ns, deploy, node string) map[string]string {
		return map[string]string{
			"k8s_namespace_name":  ns,
			"k8s.deployment.name": deploy,
			"k8s_node_name":       node,
		}
	}

	t.Run("takes max across pods of same deployment", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(labels("flux-system", "source-controller", "node-a"), "0.1", "0.3"),
			fixtureSeries(labels("flux-system", "source-controller", "node-b"), "0.2", "0.5"),
		}
		result := LatestByDeployment(series)
		assert.InDelta(t, 0.5, result["flux-system/source-controller"], 0.001)
	})

	t.Run("skips entries with missing namespace", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{"k8s.deployment.name": "myapp"}, "0.5"),
		}
		result := LatestByDeployment(series)
		_, exists := result["/myapp"]
		assert.False(t, exists, "should skip when namespace is missing")
	})

	t.Run("skips entries with missing deployment name", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{"k8s_namespace_name": "flux-system"}, "0.5"),
		}
		result := LatestByDeployment(series)
		assert.Empty(t, result)
	})

	t.Run("skips empty series", func(t *testing.T) {
		series := []v3Series{
			{Labels: labels("flux-system", "source-controller", "node-a"), Values: []v3Point{}},
		}
		result := LatestByDeployment(series)
		assert.Empty(t, result)
	})

	t.Run("falls back to statefulset name", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{
				"k8s_namespace_name":   "signoz",
				"k8s.statefulset.name": "clickhouse",
				"k8s_node_name":        "node-a",
			}, "3.5"),
		}
		result := LatestByDeployment(series)
		assert.InDelta(t, 3.5, result["signoz/clickhouse"], 0.001)
	})

	t.Run("falls back to daemonset name", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{
				"k8s_namespace_name": "kube-system",
				"k8s.daemonset.name": "cilium-ds",
				"k8s_node_name":      "node-a",
			}, "0.8"),
		}
		result := LatestByDeployment(series)
		assert.InDelta(t, 0.8, result["kube-system/cilium-ds"], 0.001)
	})

	t.Run("skips unparseable values", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(labels("flux-system", "source-controller", "node-a"), "notanumber"),
		}
		result := LatestByDeployment(series)
		assert.Empty(t, result)
	})
}

func TestSparkByDeployment(t *testing.T) {
	labels := func(ns, deploy string) map[string]string {
		return map[string]string{
			"k8s_namespace_name":  ns,
			"k8s.deployment.name": deploy,
		}
	}

	t.Run("generates sparkline points", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(labels("flux-system", "source-controller"), "0.1", "0.3", "0.2", "0.4"),
		}
		result := SparkByDeployment(series)
		spark, ok := result["flux-system/source-controller"]
		require.True(t, ok)
		assert.Contains(t, spark, ",") // has x,y pairs
	})

	t.Run("first pod wins for same deployment", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(labels("flux-system", "source-controller"), "0.1", "0.2"),
			fixtureSeries(labels("flux-system", "source-controller"), "0.9", "0.8"),
		}
		result := SparkByDeployment(series)
		spark := result["flux-system/source-controller"]
		// First pod's values should be used (starts with 0.0 for x=0)
		assert.Contains(t, spark, "0.0,")
	})

	t.Run("clamps corrupted spikes > 100 to 0", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(labels("flux-system", "source-controller"), "0.1", "150", "0.2"),
		}
		result := SparkByDeployment(series)
		spark := result["flux-system/source-controller"]
		// Should not contain any point with 150
		assert.NotContains(t, spark, "150")
		assert.NotEmpty(t, spark)
	})

	t.Run("duplicates single point for sparkline", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(labels("flux-system", "source-controller"), "0.5"),
		}
		result := SparkByDeployment(series)
		spark := result["flux-system/source-controller"]
		assert.NotEmpty(t, spark, "single value should produce a flat line")
	})

	t.Run("skips missing namespace", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{"k8s.deployment.name": "myapp"}, "0.5"),
		}
		result := SparkByDeployment(series)
		assert.Empty(t, result)
	})
}

func TestLatestByNode(t *testing.T) {
	t.Run("extracts latest value per node", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{"k8s.node.name": "node-a"}, "0.5", "0.6"),
			fixtureSeries(map[string]string{"k8s.node.name": "node-b"}, "1.2", "1.3"),
		}
		result := LatestByNode(series)
		assert.InDelta(t, 0.6, result["node-a"], 0.001)
		assert.InDelta(t, 1.3, result["node-b"], 0.001)
	})

	t.Run("skips missing node name", func(t *testing.T) {
		series := []v3Series{
			fixtureSeries(map[string]string{}, "0.5"),
		}
		result := LatestByNode(series)
		assert.Empty(t, result)
	})
}

func TestBuildServices(t *testing.T) {
	cpuLabels := func(ns, deploy, node string) map[string]string {
		return map[string]string{
			"k8s_namespace_name":  ns,
			"k8s.deployment.name": deploy,
			"k8s_node_name":       node,
		}
	}

	t.Run("filters services with all-zero metrics", func(t *testing.T) {
		cpuSeries := []v3Series{
			fixtureSeries(cpuLabels("flux-system", "source-controller", "node-a"), "0.1"),
			fixtureSeries(cpuLabels("kube-system", "some-dormant-svc", "node-a"), "0"),
		}
		allResults := map[string][]v3Series{
			"cpu": cpuSeries,
			"ram": {
				fixtureSeries(cpuLabels("flux-system", "source-controller", "node-a"), "110100480"),
			},
		}
		services := BuildServices(cpuSeries, allResults, time.Now())
		names := make(map[string]bool)
		for _, s := range services {
			names[s.Name] = true
		}
		assert.True(t, names["source-controller"], "service with metrics should be included")
		assert.False(t, names["some-dormant-svc"], "service with all-zero metrics should be filtered")
	})

	t.Run("sorts by category order then name", func(t *testing.T) {
		cpuSeries := []v3Series{
			fixtureSeries(cpuLabels("signoz", "alertmanager", "node-a"), "0.1"),
			fixtureSeries(cpuLabels("flux-system", "source-controller", "node-a"), "0.2"),
		}
		allResults := map[string][]v3Series{"cpu": cpuSeries}
		services := BuildServices(cpuSeries, allResults, time.Now())
		require.Len(t, services, 2)
		// deployment (flux-system) comes before observability (signoz)
		assert.Equal(t, "source-controller", services[0].Name)
		assert.Equal(t, "alertmanager", services[1].Name)
	})

	t.Run("clamps CPU values > 100 cores to zero", func(t *testing.T) {
		cpuSeries := []v3Series{
			fixtureSeries(cpuLabels("flux-system", "source-controller", "node-a"), "150"),
		}
		allResults := map[string][]v3Series{
			"cpu": cpuSeries,
			"ram": {
				fixtureSeries(cpuLabels("flux-system", "source-controller", "node-a"), "110100480"),
			},
		}
		services := BuildServices(cpuSeries, allResults, time.Now())
		require.Len(t, services, 1)
		assert.Equal(t, 0.0, services[0].CPU, "corrupted CPU >100 should be clamped")
	})
}

func TestBuildHosts(t *testing.T) {
	t.Run("labels control-plane nodes as Cloud", func(t *testing.T) {
		results := map[string][]v3Series{
			"nodeCpu": {fixtureSeries(map[string]string{"k8s.node.name": "talosoci-control-plane-abc"}, "0.5")},
			"nodeRam": {fixtureSeries(map[string]string{"k8s.node.name": "talosoci-control-plane-abc"}, "4294967296")},
		}
		hosts := BuildHosts(results)
		require.Len(t, hosts, 1)
		assert.Equal(t, "Cloud", hosts[0].Label)
		assert.Equal(t, "Control plane", hosts[0].Detail)
	})

	t.Run("labels worker nodes as Edge", func(t *testing.T) {
		results := map[string][]v3Series{
			"nodeCpu": {fixtureSeries(map[string]string{"k8s.node.name": "talosedge-xyz"}, "1.2")},
			"nodeRam": {fixtureSeries(map[string]string{"k8s.node.name": "talosedge-xyz"}, "8589934592")},
		}
		hosts := BuildHosts(results)
		require.Len(t, hosts, 1)
		assert.Equal(t, "Edge", hosts[0].Label)
		assert.Equal(t, "Worker node", hosts[0].Detail)
	})

	t.Run("counts services per node from CPU series", func(t *testing.T) {
		results := map[string][]v3Series{
			"cpu": {
				fixtureSeries(map[string]string{"k8s_node_name": "node-a", "k8s_namespace_name": "ns1", "k8s.deployment.name": "svc1"}, "0.1"),
				fixtureSeries(map[string]string{"k8s_node_name": "node-a", "k8s_namespace_name": "ns2", "k8s.deployment.name": "svc2"}, "0.2"),
				fixtureSeries(map[string]string{"k8s_node_name": "node-b", "k8s_namespace_name": "ns3", "k8s.deployment.name": "svc3"}, "0.3"),
			},
		}
		hosts := BuildHosts(results)
		hostMap := map[string]Host{}
		for _, h := range hosts {
			hostMap[h.Name] = h
		}
		assert.Equal(t, 2, hostMap["node-a"].SvcCount)
		assert.Equal(t, 1, hostMap["node-b"].SvcCount)
	})

	t.Run("computes uptime from nodeUp series", func(t *testing.T) {
		results := map[string][]v3Series{
			"nodeUp": {fixtureSeries(map[string]string{"k8s.node.name": "node-a"}, "536077")},
		}
		hosts := BuildHosts(results)
		require.Len(t, hosts, 1)
		assert.Equal(t, "6d", hosts[0].Uptime)
	})
}
