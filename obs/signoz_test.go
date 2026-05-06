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
	// Mock server that returns deployment-level metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("SIGNOZ-API-KEY"))
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query().Get("query")
		resp := promResponse{Status: "success", Data: struct {
			ResultType string           `json:"resultType"`
			Result     []promResultItem `json:"result"`
		}{ResultType: "vector"}}

		switch {
		case query == "up":
			resp.Data.Result = []promResultItem{
				{Metric: map[string]string{"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website", "k8s.pod.start_time": time.Now().Add(-2 * time.Hour).Format(time.RFC3339)}, Value: []interface{}{float64(1714896000), "1"}},
				{Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller"}, Value: []interface{}{float64(1714896000), "1"}},
				{Metric: map[string]string{"k8s_namespace_name": "dapr-system", "k8s.deployment.name": "dapr-operator"}, Value: []interface{}{float64(1714896000), "1"}},
			}
		case query == "go_goroutines":
			resp.Data.Result = []promResultItem{
				{Metric: map[string]string{"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website"}, Value: []interface{}{float64(1714896000), "75"}},
				{Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller"}, Value: []interface{}{float64(1714896000), "66"}},
			}
		case query == "process_resident_memory_bytes":
			resp.Data.Result = []promResultItem{
				{Metric: map[string]string{"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website"}, Value: []interface{}{float64(1714896000), "127926272"}}, // 122 MB
				{Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller"}, Value: []interface{}{float64(1714896000), "111149056"}},     // 106 MB
			}
		case query == `dapr_runtime_component_loaded{k8s_namespace_name="platform-website"}`:
			resp.Data.Result = []promResultItem{
				{Metric: map[string]string{"k8s_namespace_name": "platform-website"}, Value: []interface{}{float64(1714896000), "2"}},
			}
		case query == `go_info{k8s_namespace_name="platform-website"}`:
			resp.Data.Result = []promResultItem{
				{Metric: map[string]string{"k8s_namespace_name": "platform-website", "version": "go1.23.6"}, Value: []interface{}{float64(1714896000), "1"}},
			}
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

	t.Run("each category has services", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			assert.NotEmpty(t, cat.Services, "Category %s should have services", cat.Name)
		}
	})

	t.Run("monitored services get goroutines", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Name == "platform-website" {
					assert.Equal(t, "75 goroutines", svc.Metric)
				}
				if svc.Name == "flux-source" {
					assert.Equal(t, "66 goroutines", svc.Metric)
				}
			}
		}
	})

	t.Run("monitored services get memory", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Name == "platform-website" {
					assert.Equal(t, "122 MB", svc.Memory)
				}
				if svc.Name == "flux-source" {
					assert.Equal(t, "106 MB", svc.Memory)
				}
			}
		}
	})

	t.Run("dapr sidecar shows component count", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Name == "daprd" {
					assert.Equal(t, "2 components", svc.Metric)
					assert.Equal(t, "healthy", svc.Status)
				}
			}
		}
	})

	t.Run("unmonitored services are unmonitored", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Name == "forgejo" || svc.Name == "signoz-collector" {
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

	t.Run("stats strip populated", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		assert.Contains(t, data.Stats.GoVersion, "go1.23.6")
		assert.Equal(t, 2, data.Stats.NodeCount)
	})

	t.Run("timestamp is set", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		assert.Greater(t, data.Timestamp, int64(0))
	})

	t.Run("each service has updatedAt", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				assert.Greater(t, svc.UpdatedAt, int64(0), "Service %s", svc.Name)
			}
		}
	})

	t.Run("serializes to JSON", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		b, err := json.Marshal(data)
		require.NoError(t, err)
		assert.Contains(t, string(b), "platform-website")
		assert.Contains(t, string(b), "goroutines")
		assert.Contains(t, string(b), "MB")
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

	t.Run("total service count is 17", func(t *testing.T) {
		total := 0
		for _, cat := range cats {
			total += len(cat.Services)
		}
		assert.Equal(t, 17, total)
	})

	t.Run("services have deployment where monitored", func(t *testing.T) {
		for _, cat := range cats {
			for _, svc := range cat.Services {
				if svc.Namespace != "" && MonitoredNamespaces()[svc.Namespace] {
					assert.NotEmpty(t, svc.Deployment, "Service %s should have deployment", svc.Name)
				}
			}
		}
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

func TestDeploymentKey(t *testing.T) {
	assert.Equal(t, "flux-system/source-controller", deploymentKey("flux-system", "source-controller"))
}
