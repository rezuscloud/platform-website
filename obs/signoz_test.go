package obs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigNozClientFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("SIGNOZ-API-KEY"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/query":
			resp := promResponse{Status: "success"}
			resp.Data.ResultType = "vector"
			resp.Data.Result = []promResultItem{
				{
					Metric: map[string]string{
						"k8s_node_name":      "talosedge-genmachiche-flowing-bluejay",
						"k8s_pod_name":       "platform-website-abc123",
						"k8s_namespace_name": "platform-website",
					},
					Value: []interface{}{float64(1714896000), "1"},
				},
			}
			json.NewEncoder(w).Encode(resp)

		case "/api/v1/query_range":
			resp := promRangeResponse{Status: "success"}
			resp.Data.ResultType = "matrix"
			resp.Data.Result = []promResultItem{
				{
					Metric: map[string]string{"k8s_namespace_name": "platform-website"},
					Values: [][]interface{}{
						{float64(1714896000), "72"},
						{float64(1714896060), "74"},
						{float64(1714896120), "71"},
						{float64(1714896180), "73"},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := NewSigNozClient(server.URL, "test-api-key")

	t.Run("Fetch returns platform topology", func(t *testing.T) {
		data, err := client.Fetch(context.Background())
		require.NoError(t, err)

		assert.Len(t, data.Services, 5)
		assert.Len(t, data.Edges, 4)
	})

	t.Run("services have names from topology", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		assert.Equal(t, "cilium-gateway", data.Services[0].Name)
		assert.Equal(t, "platform-website", data.Services[1].Name)
		assert.Equal(t, "daprd", data.Services[2].Name)
	})

	t.Run("services get status from up metric", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		// platform-website should be healthy (the mock server returns a result)
		assert.Equal(t, "healthy", data.Services[1].Status)
	})

	t.Run("metrics sparklines populated for platform-website", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		assert.NotEmpty(t, data.Services[1].Metrics)
		assert.Equal(t, "Goroutines", data.Services[1].Metrics[0].Label)
		assert.NotEmpty(t, data.Services[1].Metrics[0].Points)
	})

	t.Run("edges connect services correctly", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		assert.Equal(t, "cilium-gateway", data.Edges[0].From)
		assert.Equal(t, "platform-website", data.Edges[0].To)
		assert.Equal(t, "HTTPS", data.Edges[0].Label)
	})
}

func TestSigNozClientGracefulDegradation(t *testing.T) {
	t.Run("unreachable URL still returns topology", func(t *testing.T) {
		client := NewSigNozClient("http://127.0.0.1:1", "test-key")
		data, err := client.Fetch(context.Background())
		// Returns topology with services marked as unknown, no hard error
		assert.NoError(t, err)
		assert.Len(t, data.Services, 5)
		for _, s := range data.Services {
			assert.Equal(t, "unknown", s.Status)
		}
	})
}

func TestNewSigNozClientFromEnv(t *testing.T) {
	t.Run("returns nil when env vars are missing", func(t *testing.T) {
		result := NewSigNozClientFromEnv()
		assert.Nil(t, result)
	})
}

func TestPlatformTopology(t *testing.T) {
	data := PlatformTopology()

	t.Run("has 5 services", func(t *testing.T) {
		assert.Len(t, data.Services, 5)
	})

	t.Run("has 4 edges", func(t *testing.T) {
		assert.Len(t, data.Edges, 4)
	})

	t.Run("has kinds", func(t *testing.T) {
		kinds := make(map[string]string)
		for _, s := range data.Services {
			kinds[s.Name] = s.Kind
		}
		assert.Equal(t, "ingress", kinds["cilium-gateway"])
		assert.Equal(t, "app", kinds["platform-website"])
		assert.Equal(t, "sidecar", kinds["daprd"])
		assert.Equal(t, "infra", kinds["signoz-collector"])
		assert.Equal(t, "infra", kinds["dapr-control-plane"])
	})
}
