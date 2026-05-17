package obs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryForNamespace(t *testing.T) {
	tests := []struct {
		ns       string
		expected string
	}{
		// Development
		{"forgejo", "dev"},
		{"arc-systems", "dev"},
		// Deployment
		{"flux-system", "deployment"},
		{"vela-system", "deployment"},
		{"external-dns", "deployment"},
		{"cert-manager", "deployment"},
		// Runtime
		{"kube-system", "runtime"},
		{"platform-website", "runtime"},
		{"dapr-system", "runtime"},
		// Observability
		{"signoz", "observability"},
		{"monitoring", "observability"},
		// Data
		{"tikv-system", "data"},
		{"juicefs-csi", "data"},
		{"velero", "data"},
		// Default (unknown namespace)
		{"unknown-namespace", "runtime"},
		{"custom-app", "runtime"},
	}
	for _, tt := range tests {
		t.Run(tt.ns, func(t *testing.T) {
			assert.Equal(t, tt.expected, CategoryForNamespace(tt.ns))
		})
	}
}

func TestCategoryLabel(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"hosts", "Hosts"},
		{"dev", "Development"},
		{"deployment", "Deployment"},
		{"runtime", "Runtime"},
		{"observability", "Observability"},
		{"data", "Data"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			assert.Equal(t, tt.expected, CategoryLabel(tt.id))
		})
	}
}

func TestCategoryOrder(t *testing.T) {
	assert.Equal(t, []string{"dev", "deployment", "runtime", "observability", "data"}, CategoryOrder)
	assert.Len(t, CategoryOrder, 5)
}

func TestDiscoverNodeNames(t *testing.T) {
	t.Run("merges from multiple sources", func(t *testing.T) {
		floatMap := map[string]float64{"node-a": 0.5, "node-b": 1.2}
		intMap := map[string]int{"node-c": 3}
		names := DiscoverNodeNames(floatMap, intMap)
		assert.ElementsMatch(t, []string{"node-a", "node-b", "node-c"}, names)
	})

	t.Run("deduplicates", func(t *testing.T) {
		floatMap := map[string]float64{"node-a": 0.5}
		intMap := map[string]int{"node-a": 5}
		names := DiscoverNodeNames(floatMap, intMap)
		assert.Equal(t, []string{"node-a"}, names)
	})

	t.Run("control-plane nodes sort first", func(t *testing.T) {
		m := map[string]float64{
			"talosedge-worker-xyz":             1.2,
			"talosoci-control-plane-abc":       0.5,
			"talosedge-another-node":           0.8,
			"talosoci-control-plane-secondary": 0.3,
		}
		names := DiscoverNodeNames(m)
		require.Len(t, names, 4)
		assert.Contains(t, names[0], "control-plane")
		assert.Contains(t, names[1], "control-plane")
		assert.Contains(t, names[2], "edge")
		assert.Contains(t, names[3], "edge")
	})

	t.Run("empty sources", func(t *testing.T) {
		names := DiscoverNodeNames()
		assert.Empty(t, names)
	})
}

func TestSortServices(t *testing.T) {
	t.Run("sorts by category order", func(t *testing.T) {
		services := []Service{
			{Name: "clickhouse", Category: "observability"},
			{Name: "source-controller", Category: "deployment"},
			{Name: "forgejo", Category: "dev"},
		}
		SortServices(services)
		assert.Equal(t, "forgejo", services[0].Name)
		assert.Equal(t, "source-controller", services[1].Name)
		assert.Equal(t, "clickhouse", services[2].Name)
	})

	t.Run("sorts by name within same category", func(t *testing.T) {
		services := []Service{
			{Name: "velero", Category: "data"},
			{Name: "juicefs", Category: "data"},
		}
		SortServices(services)
		assert.Equal(t, "juicefs", services[0].Name)
		assert.Equal(t, "velero", services[1].Name)
	})
}

func TestLabelStr(t *testing.T) {
	t.Run("returns first matching key", func(t *testing.T) {
		labels := map[string]string{
			"k8s_namespace_name": "flux-system",
			"k8s.namespace.name": "wrong",
		}
		assert.Equal(t, "flux-system", LabelStr(labels, "k8s_namespace_name", "k8s.namespace.name"))
	})

	t.Run("falls back to second key", func(t *testing.T) {
		labels := map[string]string{
			"k8s.namespace.name": "flux-system",
		}
		assert.Equal(t, "flux-system", LabelStr(labels, "k8s_namespace_name", "k8s.namespace.name"))
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		labels := map[string]string{"other": "value"}
		assert.Equal(t, "", LabelStr(labels, "k8s_namespace_name", "k8s.namespace.name"))
	})

	t.Run("returns empty for empty label value", func(t *testing.T) {
		labels := map[string]string{"k8s_namespace_name": ""}
		assert.Equal(t, "", LabelStr(labels, "k8s_namespace_name"))
	})

	t.Run("returns empty for nil labels", func(t *testing.T) {
		assert.Equal(t, "", LabelStr(nil, "k8s_namespace_name"))
	})
}
