package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rezuscloud/platform-website/obs"
)

func TestSetLiveClient(t *testing.T) {
	t.Run("replaces the default client", func(t *testing.T) {
		original := liveClient
		defer func() { liveClient = original }()

		mock := &obs.MockClient{}
		SetLiveClient(mock)
		assert.Equal(t, mock, liveClient)
	})

	t.Run("ignores nil", func(t *testing.T) {
		original := liveClient
		defer func() { liveClient = original }()

		SetLiveClient(nil)
		assert.Equal(t, original, liveClient)
	})
}

func TestDefaultMockData(t *testing.T) {
	data := obs.DefaultMockData()

	t.Run("has services", func(t *testing.T) {
		assert.NotEmpty(t, data.Services)
	})

	t.Run("services have names from SigNoz", func(t *testing.T) {
		names := make(map[string]bool)
		for _, svc := range data.Services {
			names[svc.Name] = true
		}
		assert.True(t, names["source-controller"])
		assert.True(t, names["platform-website"])
		assert.True(t, names["forgejo"])
	})

	t.Run("services have CPU and RAM", func(t *testing.T) {
		for _, svc := range data.Services {
			if svc.Status == "healthy" && svc.Name != "signoz-otel-collector" {
				assert.GreaterOrEqual(t, svc.CPU, float64(0), "Service %s", svc.Name)
				assert.GreaterOrEqual(t, svc.RAM, float64(0), "Service %s", svc.Name)
			}
		}
	})

	t.Run("services have categories", func(t *testing.T) {
		for _, svc := range data.Services {
			assert.NotEmpty(t, svc.Category, "Service %s", svc.Name)
		}
	})

	t.Run("has no live metrics in mock mode", func(t *testing.T) {
		assert.False(t, data.HasMetrics)
	})
}

func TestLiveClientInterface(t *testing.T) {
	t.Run("mock client implements Client interface", func(t *testing.T) {
		var _ obs.Client = &obs.MockClient{}
	})
}
