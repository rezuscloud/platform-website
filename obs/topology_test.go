package obs

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
