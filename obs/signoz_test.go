package obs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigNozClientFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("SIGNOZ-API-KEY"))
		w.Header().Set("Content-Type", "application/json")

		resp := promResponse{Status: "success"}
		resp.Data.ResultType = "vector"
		resp.Data.Result = []promResultItem{
			{Metric: map[string]string{"k8s_namespace_name": "platform-website"}, Value: []interface{}{float64(1714896000), "1"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSigNozClient(server.URL, "test-api-key")

	t.Run("returns full category grid", func(t *testing.T) {
		data, err := client.Fetch(context.Background())
		require.NoError(t, err)
		assert.Len(t, data.Categories, 5)
		assert.True(t, data.HasMetrics)
	})

	t.Run("has 5 categories", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		ids := make([]string, len(data.Categories))
		for i, c := range data.Categories {
			ids[i] = c.ID
		}
		assert.Equal(t, []string{"infra", "dev", "delivery", "runtime", "observability"}, ids)
	})

	t.Run("monitored namespaces are healthy", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		// The mock server returns a result for platform-website namespace
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Namespace == "platform-website" {
					assert.Equal(t, "healthy", svc.Status, "Service %s", svc.Name)
				}
			}
		}
	})

	t.Run("unmonitored namespaces are unmonitored", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Namespace == "forgejo" || svc.Namespace == "signoz" || svc.Namespace == "arc-systems" {
					assert.Equal(t, "unmonitored", svc.Status, "Service %s", svc.Name)
				}
			}
		}
	})

	t.Run("infrastructure nodes are running", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, svc := range data.Categories[0].Services {
			assert.Equal(t, "running", svc.Status, "Node %s", svc.Name)
		}
	})
}

func TestSigNozClientGracefulDegradation(t *testing.T) {
	t.Run("unreachable URL still returns topology", func(t *testing.T) {
		client := NewSigNozClient("http://127.0.0.1:1", "test-key")
		data, err := client.Fetch(context.Background())
		assert.NoError(t, err)
		assert.Len(t, data.Categories, 5)
	})
}

func TestNewSigNozClientFromEnv(t *testing.T) {
	t.Run("returns nil when env vars are missing", func(t *testing.T) {
		result := NewSigNozClientFromEnv()
		assert.Nil(t, result)
	})
}

func TestPlatformCategories(t *testing.T) {
	cats := PlatformCategories()

	t.Run("has 5 categories", func(t *testing.T) {
		assert.Len(t, cats, 5)
	})

	t.Run("each category has services", func(t *testing.T) {
		for _, cat := range cats {
			assert.NotEmpty(t, cat.Services, "Category %s should have services", cat.Name)
		}
	})

	t.Run("total service count is reasonable", func(t *testing.T) {
		total := 0
		for _, cat := range cats {
			total += len(cat.Services)
		}
		assert.GreaterOrEqual(t, total, 15, "Should have 15+ services across all categories")
	})
}

func TestMonitoredNamespaces(t *testing.T) {
	monitored := MonitoredNamespaces()
	assert.True(t, monitored["flux-system"])
	assert.True(t, monitored["dapr-system"])
	assert.True(t, monitored["platform-website"])
	assert.False(t, monitored["forgejo"])
	assert.False(t, monitored["signoz"])
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "2d", formatDuration(48*time.Hour))
	assert.Equal(t, "3h", formatDuration(3*time.Hour))
	assert.Equal(t, "45m", formatDuration(45*time.Minute))
}
