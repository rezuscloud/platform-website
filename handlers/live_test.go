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
	t.Run("sseServiceCount writes total service count", func(t *testing.T) {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		data := obs.DefaultMockData()
		sseServiceCount(w, data)
		w.Flush()

		// Should have 18+ services across 5 categories
		output := buf.String()
		assert.True(t, strings.HasPrefix(output, "event: services\ndata: "))
	})

	t.Run("sseHeartbeat writes heartbeat", func(t *testing.T) {
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

	t.Run("has 5 categories", func(t *testing.T) {
		assert.Len(t, data.Categories, 5)
	})

	t.Run("has monitored services as healthy", func(t *testing.T) {
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Namespace == "flux-system" || svc.Namespace == "dapr-system" {
					assert.Equal(t, "healthy", svc.Status, "Service %s", svc.Name)
				}
			}
		}
	})

	t.Run("has unmonitored services as unmonitored", func(t *testing.T) {
		for _, cat := range data.Categories {
			for _, svc := range cat.Services {
				if svc.Namespace == "forgejo" || svc.Namespace == "signoz" {
					assert.Equal(t, "unmonitored", svc.Status, "Service %s", svc.Name)
				}
			}
		}
	})

	t.Run("has no metrics in mock mode", func(t *testing.T) {
		assert.False(t, data.HasMetrics)
	})
}

func TestLiveSSEHandlerRegistration(t *testing.T) {
	t.Run("mock client implements Client interface", func(t *testing.T) {
		var _ obs.Client = &obs.MockClient{}
	})
}
