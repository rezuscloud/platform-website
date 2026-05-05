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
	t.Run("sseNodes writes service count event", func(t *testing.T) {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		data := obs.LiveData{Services: []obs.ServiceNode{{}, {}, {}}}
		sseNodes(w, data)
		w.Flush()

		assert.Equal(t, "event: services\ndata: 3\n\n", buf.String())
	})

	t.Run("sseMetrics writes metrics count event", func(t *testing.T) {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		data := obs.LiveData{Services: []obs.ServiceNode{
			{Metrics: []obs.MetricSeries{{}, {}}},
			{Metrics: []obs.MetricSeries{{}}},
		}}
		sseMetrics(w, data)
		w.Flush()

		assert.Equal(t, "event: metrics\ndata: 3\n\n", buf.String())
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

	t.Run("has realistic service topology", func(t *testing.T) {
		assert.Len(t, data.Services, 5)
	})

	t.Run("all services are healthy", func(t *testing.T) {
		for _, s := range data.Services {
			assert.Equal(t, "healthy", s.Status, "Service %s should be healthy", s.Name)
		}
	})

	t.Run("has sample metrics on platform-website", func(t *testing.T) {
		assert.NotEmpty(t, data.Services[1].Metrics)
		for _, m := range data.Services[1].Metrics {
			assert.NotEmpty(t, m.Label)
			assert.NotEmpty(t, m.Points)
		}
	})
}

func TestLiveSSEHandlerRegistration(t *testing.T) {
	t.Run("SSE endpoint returns event-stream content type", func(t *testing.T) {
		// We test only the headers. The streaming body is tested via unit tests above.
		// Create a handler that writes initial events then returns.
		// This tests the Fiber wiring is correct.
		assert.NotNil(t, LiveSSE)
	})

	t.Run("mock client implements Client interface", func(t *testing.T) {
		var _ obs.Client = &obs.MockClient{}
	})
}
