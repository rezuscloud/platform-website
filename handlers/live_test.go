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

	t.Run("has 5 categories", func(t *testing.T) {
		assert.Len(t, data.Categories, 5)
	})

	t.Run("has 17 services", func(t *testing.T) {
		total := 0
		for _, cat := range data.Categories {
			total += len(cat.Services)
		}
		assert.Equal(t, 17, total)
	})

	t.Run("monitored services are healthy with metrics", func(t *testing.T) {
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Deployment != "" {
					assert.Equal(t, "healthy", svc.Status, "Service %s", svc.Name)
					assert.NotEmpty(t, svc.Metric, "Service %s should have metric", svc.Name)
				}
			}
		}
	})

	t.Run("unmonitored services are unmonitored", func(t *testing.T) {
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Deployment == "" && svc.Namespace != "" {
					assert.Equal(t, "unmonitored", svc.Status, "Service %s", svc.Name)
				}
			}
		}
	})

	t.Run("infrastructure nodes are running", func(t *testing.T) {
		for _, svc := range data.Categories[0].Services {
			assert.Equal(t, "running", svc.Status)
		}
	})

	t.Run("has no live metrics in mock mode", func(t *testing.T) {
		assert.False(t, data.HasMetrics)
	})

	t.Run("has timestamp", func(t *testing.T) {
		assert.Greater(t, data.Timestamp, int64(0))
	})

	t.Run("each service has updatedAt", func(t *testing.T) {
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				assert.Greater(t, svc.UpdatedAt, int64(0), "Service %s", svc.Name)
			}
		}
	})
}

func TestLiveClientInterface(t *testing.T) {
	t.Run("mock client implements Client interface", func(t *testing.T) {
		var _ obs.Client = &obs.MockClient{}
	})
}
