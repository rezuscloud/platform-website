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
	// Create a mock SigNoz server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-api-key", r.Header.Get("SIGNOZ-API-KEY"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/query":
			// Return mock "up" response with two pods on two nodes
			resp := promResponse{
				Status: "success",
			}
			resp.Data.ResultType = "vector"
			resp.Data.Result = []promResultItem{
				{
					Metric: map[string]string{
						"k8s_node_name":      "talosoci-control-plane-legal-poodle",
						"k8s_pod_name":       "forgejo-76bb8587c6-r4bbc",
						"k8s_namespace_name": "forgejo",
					},
					Value: []interface{}{float64(1714896000), "1"},
				},
				{
					Metric: map[string]string{
						"k8s_node_name":      "talosoci-control-plane-legal-poodle",
						"k8s_pod_name":       "source-controller-5f5f984f54",
						"k8s_namespace_name": "flux-system",
					},
					Value: []interface{}{float64(1714896000), "1"},
				},
				{
					Metric: map[string]string{
						"k8s_node_name":      "talosedge-genmachiche-flowing-bluejay",
						"k8s_pod_name":       "platform-website-69f7bffd5f-trltm",
						"k8s_namespace_name": "platform-website",
					},
					Value: []interface{}{float64(1714896000), "1"},
				},
			}
			json.NewEncoder(w).Encode(resp)

		case "/api/v1/query_range":
			// Return mock sparkline data
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

	t.Run("Fetch returns nodes from up metric", func(t *testing.T) {
		data, err := client.Fetch(context.Background())
		require.NoError(t, err)

		assert.Len(t, data.Nodes, 2)
		assert.Equal(t, "talosoci-control-plane-legal-poodle", data.Nodes[0].Name)
		assert.Equal(t, "talosedge-genmachiche-flowing-bluejay", data.Nodes[1].Name)
	})

	t.Run("nodes have correct tier", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		assert.Equal(t, "oci-cloud", data.Nodes[0].Tier)
		assert.Equal(t, "edge", data.Nodes[1].Tier)
	})

	t.Run("pods are grouped by node", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		assert.Len(t, data.Nodes[0].Pods, 2) // forgejo, source-controller
		assert.Len(t, data.Nodes[1].Pods, 1) // platform-website
	})

	t.Run("pods have namespace and status", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		pod := data.Nodes[0].Pods[0]
		assert.Equal(t, "forgejo", pod.Namespace)
		assert.Equal(t, "Running", pod.Status)
	})

	t.Run("metrics sparklines are populated", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())

		assert.NotEmpty(t, data.Metrics)
		assert.Equal(t, "Goroutines", data.Metrics[0].Label)
		assert.NotEmpty(t, data.Metrics[0].Points)
	})
}

func TestTierFromNodeName(t *testing.T) {
	assert.Equal(t, "oci-cloud", tierFromNodeName("talosoci-control-plane-legal-poodle"))
	assert.Equal(t, "edge", tierFromNodeName("talosedge-genmachiche-flowing-bluejay"))
	assert.Equal(t, "edge", tierFromNodeName("some-other-node"))
}

func TestSigNozClientError(t *testing.T) {
	t.Run("invalid URL returns error", func(t *testing.T) {
		client := NewSigNozClient("http://localhost:1", "test-key")
		_, err := client.Fetch(context.Background())
		assert.Error(t, err)
	})

	t.Run("non-success status returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": "bad query"})
		}))
		defer server.Close()

		client := NewSigNozClient(server.URL, "test-key")
		_, err := client.Fetch(context.Background())
		assert.Error(t, err)
	})
}

func TestNewSigNozClientFromEnv(t *testing.T) {
	t.Run("returns nil when env vars are missing", func(t *testing.T) {
		result := NewSigNozClientFromEnv()
		assert.Nil(t, result)
	})
}
