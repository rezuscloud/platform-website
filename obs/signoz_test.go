package obs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigNozClientFetch(t *testing.T) {
	promResp := func(results []map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result":     results,
			},
		}
	}

	promResult := func(labels map[string]string, val string) map[string]interface{} {
		return map[string]interface{}{
			"metric": labels,
			"value":  []interface{}{0, val},
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("SIGNOZ-API-KEY"))
		w.Header().Set("Content-Type", "application/json")

		if strings.HasPrefix(r.URL.Path, "/api/v1/query") {
			query := r.URL.Query().Get("query")
			switch {
			case query == "up":
				json.NewEncoder(w).Encode(promResp([]map[string]interface{}{
					promResult(map[string]string{
						"k8s_namespace_name": "flux-system", "k8s.deployment.name": "source-controller",
						"k8s_node_name": "node-a", "k8s.pod.start_time": time.Now().Add(-4 * 24 * time.Hour).Format(time.RFC3339),
					}, "1"),
					promResult(map[string]string{
						"k8s_namespace_name": "platform-website", "k8s.deployment.name": "platform-website",
						"k8s_node_name": "node-b", "k8s.pod.start_time": time.Now().Add(-22 * time.Hour).Format(time.RFC3339),
					}, "1"),
					promResult(map[string]string{
						"k8s_namespace_name": "dapr-system", "k8s.deployment.name": "dapr-operator",
						"k8s_node_name": "node-b",
					}, "1"),
				}))
			case strings.Contains(query, "count(up)"):
				json.NewEncoder(w).Encode(promResp([]map[string]interface{}{
					promResult(map[string]string{"k8s_node_name": "node-a"}, "5"),
					promResult(map[string]string{"k8s_node_name": "node-b"}, "3"),
				}))
			default:
				json.NewEncoder(w).Encode(promResp(nil))
			}
			return
		}

		// ClickHouse queries: return JSONEachRow
		body := r.URL.Query().Get("query")
		if r.Method == "POST" {
			buf := make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			body = string(buf[:n])
		}
		_ = body
		// Return empty results for ClickHouse
		w.WriteHeader(200)
		w.Write([]byte(""))
	}))
	defer server.Close()

	client := NewSigNozClient(server.URL, "test-api-key")

	t.Run("discovers services from up metric", func(t *testing.T) {
		data, err := client.Fetch(context.Background())
		require.NoError(t, err)
		assert.True(t, data.HasMetrics)
		assert.GreaterOrEqual(t, len(data.Services), 3, "should discover at least 3 services")
	})

	t.Run("services have basic fields", func(t *testing.T) {
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
	})
}

func TestSigNozClientGracefulDegradation(t *testing.T) {
	t.Run("unreachable URL still returns static hosts", func(t *testing.T) {
		client := NewSigNozClient("http://127.0.0.1:1", "test-key")
		data, err := client.Fetch(context.Background())
		assert.NoError(t, err)
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
		assert.Contains(t, result, "0.0,")
		assert.Contains(t, result, "48.0,")
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

func jsonNumberOf(f float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", f), "0"), ".")
}
