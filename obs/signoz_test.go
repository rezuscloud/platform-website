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

	t.Run("returns full tree with metrics flag", func(t *testing.T) {
		data, err := client.Fetch(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "cilium-gateway", data.Root.Name)
		assert.Equal(t, 5, data.Root.ServiceCount())
		assert.True(t, data.HasMetrics)
	})

	t.Run("populates status from up metric", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		data.Root.Walk(func(s *ServiceNode) {
			assert.Contains(t, []string{"healthy", "unknown"}, s.Status, "Service %s", s.Name)
		})
	})

	t.Run("returns health checks", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		assert.Len(t, data.Health, 5)
	})

	t.Run("returns stats strip", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		assert.Equal(t, 2, data.Stats.NodeCount)
	})

	t.Run("edges connect correctly", func(t *testing.T) {
		data, _ := client.Fetch(context.Background())
		assert.Equal(t, "HTTPS", data.Root.Out[0].Label)
		assert.Equal(t, "platform-website", data.Root.Out[0].Target.Name)
	})
}

func TestSigNozClientGracefulDegradation(t *testing.T) {
	t.Run("unreachable URL still returns topology", func(t *testing.T) {
		client := NewSigNozClient("http://127.0.0.1:1", "test-key")
		data, err := client.Fetch(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "cilium-gateway", data.Root.Name)
		data.Root.Walk(func(s *ServiceNode) {
			assert.Equal(t, "unknown", s.Status, "Service %s", s.Name)
		})
	})
}

func TestNewSigNozClientFromEnv(t *testing.T) {
	t.Run("returns nil when env vars are missing", func(t *testing.T) {
		result := NewSigNozClientFromEnv()
		assert.Nil(t, result)
	})
}

func TestPlatformTopology(t *testing.T) {
	root := PlatformTopology()

	t.Run("root is cilium-gateway", func(t *testing.T) {
		assert.Equal(t, "cilium-gateway", root.Name)
		assert.Equal(t, "ingress", root.Kind)
	})

	t.Run("has 5 total services", func(t *testing.T) {
		assert.Equal(t, 5, root.ServiceCount())
	})

	t.Run("has correct tree structure", func(t *testing.T) {
		assert.Len(t, root.Out, 1)
		pw := root.Out[0].Target
		assert.Equal(t, "platform-website", pw.Name)
		assert.Len(t, pw.Out, 1)
		daprd := pw.Out[0].Target
		assert.Equal(t, "daprd", daprd.Name)
		assert.Len(t, daprd.Out, 2)
		assert.Equal(t, "signoz-collector", daprd.Out[0].Target.Name)
		assert.Equal(t, "dapr-control-plane", daprd.Out[1].Target.Name)
	})

	t.Run("walk visits all 5 services", func(t *testing.T) {
		var names []string
		root.Walk(func(s *ServiceNode) { names = append(names, s.Name) })
		assert.Equal(t, []string{"cilium-gateway", "platform-website", "daprd", "signoz-collector", "dapr-control-plane"}, names)
	})

	t.Run("find returns correct node", func(t *testing.T) {
		pw := root.Find("platform-website")
		require.NotNil(t, pw)
		assert.Equal(t, "app", pw.Kind)
		assert.Nil(t, root.Find("nonexistent"))
	})

	t.Run("edge labels are correct", func(t *testing.T) {
		assert.Equal(t, "HTTPS", root.Out[0].Label)
		assert.Equal(t, "localhost", root.Out[0].Target.Out[0].Label)
		assert.Equal(t, "OTLP", root.Out[0].Target.Out[0].Target.Out[0].Label)
		assert.Equal(t, "gRPC", root.Out[0].Target.Out[0].Target.Out[1].Label)
	})
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "2d", formatDuration(48*time.Hour))
	assert.Equal(t, "3h", formatDuration(3*time.Hour))
	assert.Equal(t, "45m", formatDuration(45*time.Minute))
}
