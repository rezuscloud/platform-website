package handlers

import (
	"bufio"
	"bytes"
	"context"
	"strings"
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

func TestSSEWriteHelpers(t *testing.T) {
	t.Run("sseServiceCount writes tree service count", func(t *testing.T) {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		data := obs.DefaultMockData()
		sseServiceCount(w, data)
		w.Flush()

		assert.Equal(t, "event: services\ndata: 5\n\n", buf.String())
	})

	t.Run("sseHeartbeat writes heartbeat with unix timestamp", func(t *testing.T) {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		sseHeartbeat(w)
		w.Flush()

		output := buf.String()
		assert.True(t, strings.HasPrefix(output, "event: heartbeat\ndata: "))
		assert.True(t, strings.HasSuffix(output, "\n\n"))
	})

	t.Run("sseError writes error event", func(t *testing.T) {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		sseError(w, context.DeadlineExceeded)
		w.Flush()

		assert.Equal(t, "event: error\ndata: context deadline exceeded\n\n", buf.String())
	})
}

func TestDefaultMockData(t *testing.T) {
	data := obs.DefaultMockData()

	t.Run("root is cilium-gateway", func(t *testing.T) {
		assert.Equal(t, "cilium-gateway", data.Root.Name)
	})

	t.Run("all services are unknown", func(t *testing.T) {
		data.Root.Walk(func(s *obs.ServiceNode) {
			assert.Equal(t, "unknown", s.Status, "Service %s", s.Name)
		})
	})

	t.Run("has no metrics", func(t *testing.T) {
		assert.False(t, data.HasMetrics)
	})

	t.Run("has no health checks", func(t *testing.T) {
		assert.Empty(t, data.Health)
	})
}

func TestLiveSSEHandlerRegistration(t *testing.T) {
	t.Run("mock client implements Client interface", func(t *testing.T) {
		var _ obs.Client = &obs.MockClient{}
	})
}
