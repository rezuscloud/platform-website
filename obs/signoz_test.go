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

		resp := promResponse{Status: "success", Data: struct {
			ResultType string           `json:"resultType"`
			Result     []promResultItem `json:"result"`
		}{ResultType: "vector"}}

		switch r.URL.Path {
		case "/api/v1/query":
			query := r.URL.Query().Get("query")
			switch {
			case query == "up":
				resp.Data.Result = []promResultItem{
					{Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller", "k8s_node_name": "node-a", "k8s.pod.start_time": time.Now().Add(-4 * 24 * time.Hour).Format(time.RFC3339)}, Value: []interface{}{float64(0), "1"}},
					{Metric: map[string]string{"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website", "k8s_node_name": "node-b", "k8s.pod.start_time": time.Now().Add(-22 * time.Hour).Format(time.RFC3339)}, Value: []interface{}{float64(0), "1"}},
					{Metric: map[string]string{"k8s_namespace_name": "dapr-system", "k8s.deployment.name": "dapr-operator", "k8s_node_name": "node-b"}, Value: []interface{}{float64(0), "1"}},
				}
			case query == "rate(process_cpu_seconds_total[5m])*100":
				resp.Data.Result = []promResultItem{
					{Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller"}, Value: []interface{}{float64(0), "0.25"}},
					{Metric: map[string]string{"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website"}, Value: []interface{}{float64(0), "0.11"}},
					{Metric: map[string]string{"k8s_namespace_name": "dapr-system", "k8s.deployment.name": "dapr-operator"}, Value: []interface{}{float64(0), "0.13"}},
				}
			case query == "process_resident_memory_bytes":
				resp.Data.Result = []promResultItem{
					{Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller"}, Value: []interface{}{float64(0), "110100480"}},     // 105 MB
					{Metric: map[string]string{"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website"}, Value: []interface{}{float64(0), "126812160"}}, // 121 MB
					{Metric: map[string]string{"k8s_namespace_name": "dapr-system", "k8s.deployment.name": "dapr-operator"}, Value: []interface{}{float64(0), "55574528"}},          // 53 MB
				}
			}
		case "/api/v1/query_range":
			resp.Data.ResultType = "matrix"
			resp.Data.Result = []promResultItem{
				{
					Metric: map[string]string{"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller"},
					Values: []interface{}{
						[]interface{}{float64(0), "0.20"},
						[]interface{}{float64(300), "0.22"},
						[]interface{}{float64(600), "0.25"},
					},
				},
			}
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSigNozClient(server.URL, "test-api-key")

	t.Run("discovers services from up metric", func(t *testing.T) {
		data, err := client.Fetch(context.Background())
		require.NoError(t, err)
		assert.True(t, data.HasMetrics)
		assert.GreaterOrEqual(t, len(data.Services), 3, "should discover at least 3 services")
	})

	t.Run("services are discovered", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, svc := range data.Services {
			assert.NotEmpty(t, svc.Name)
			assert.NotEmpty(t, svc.Category)
		}
	})

	t.Run("services have uptime", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, svc := range data.Services {
			if svc.Name == "source-controller" {
				assert.Equal(t, "4d", svc.Uptime)
			}
		}
	})

	t.Run("services have category from namespace", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		for _, svc := range data.Services {
			if svc.Namespace == "flux-system" {
				assert.Equal(t, "deployment", svc.Category)
			}
			if svc.Namespace == "platform-website" {
				assert.Equal(t, "runtime", svc.Category)
			}
		}
	})

	t.Run("serializes to JSON", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		b, err := json.Marshal(data)
		require.NoError(t, err)
		assert.Contains(t, string(b), "source-controller")
		assert.Contains(t, string(b), "cpu")
		assert.Contains(t, string(b), "ram")
	})
}

func TestSigNozClientGracefulDegradation(t *testing.T) {
	t.Run("unreachable URL still returns static hosts", func(t *testing.T) {
		client := NewSigNozClient("http://127.0.0.1:1", "test-key")
		data, err := client.Fetch(context.Background())
		assert.NoError(t, err)
		// Static hosts are always present even without SigNoz
		assert.Equal(t, 2, len(data.Hosts))
	})
}

func TestSparklinePoints(t *testing.T) {
	t.Run("single value returns empty", func(t *testing.T) {
		result := sparklinePoints([]float64{5.0}, 48, 16)
		assert.Empty(t, result, "single point has no line")
	})

	t.Run("two values", func(t *testing.T) {
		result := sparklinePoints([]float64{1.0, 2.0}, 48, 16)
		assert.Contains(t, result, "0.0,")  // first point at x=0
		assert.Contains(t, result, "48.0,") // last point at x=width
	})

	t.Run("constant values", func(t *testing.T) {
		result := sparklinePoints([]float64{5.0, 5.0, 5.0}, 48, 16)
		assert.NotEmpty(t, result)
	})
}

func TestCategoryForNamespace(t *testing.T) {
	assert.Equal(t, "dev", CategoryForNamespace("forgejo"))
	assert.Equal(t, "deployment", CategoryForNamespace("flux-system"))
	assert.Equal(t, "data", CategoryForNamespace("tikv-system"))
	assert.Equal(t, "observability", CategoryForNamespace("signoz"))
}

func TestFormatUptime(t *testing.T) {
	assert.Equal(t, "2d", FormatUptime(48*time.Hour))
	assert.Equal(t, "3h", FormatUptime(3*time.Hour))
	assert.Equal(t, "45m", FormatUptime(45*time.Minute))
}
