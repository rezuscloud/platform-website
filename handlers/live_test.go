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

	t.Run("reports no live metrics", func(t *testing.T) {
		assert.False(t, data.HasMetrics, "fallback must not claim live metrics")
	})

	t.Run("fabricates no hosts", func(t *testing.T) {
		// The fallback must not invent a cluster. The site's /privacy notice
		// promises the live grid shows real infrastructure, so on failure the
		// fallback surfaces "topology unavailable" rather than decoy nodes.
		assert.Empty(t, data.Hosts, "fallback must not fabricate nodes")
	})

	t.Run("fabricates no services", func(t *testing.T) {
		assert.Empty(t, data.Services, "fallback must not fabricate services")
	})

	t.Run("carries a timestamp", func(t *testing.T) {
		assert.NotZero(t, data.Timestamp)
	})
}

func TestLiveClientInterface(t *testing.T) {
	t.Run("mock client implements Client interface", func(t *testing.T) {
		var _ obs.Client = &obs.MockClient{}
	})
}
