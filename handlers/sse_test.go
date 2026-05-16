package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/obs"
)

func TestSendSnapshotSuccess(t *testing.T) {
	origClient := liveClient
	defer func() { liveClient = origClient }()

	liveClient = &obs.MockClient{
		Data: obs.DefaultMockData(),
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	result := sendSnapshot(w)
	assert.True(t, result, "sendSnapshot should return true on success")

	output := buf.String()
	assert.Contains(t, output, "event: update")
	assert.Contains(t, output, `"hosts"`)
	assert.Contains(t, output, `"services"`)
	assert.Contains(t, output, `"hasMetrics":false`)
}

func TestSendSnapshotError(t *testing.T) {
	origClient := liveClient
	defer func() { liveClient = origClient }()

	liveClient = &obs.MockClient{
		Err: fmt.Errorf("signoz connection refused"),
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	result := sendSnapshot(w)
	assert.False(t, result, "sendSnapshot should return false on error")

	output := buf.String()
	assert.Contains(t, output, ": keepalive",
		"should send keepalive comment when Fetch fails")
	assert.NotContains(t, output, "event: update",
		"should not send event data when Fetch fails")
}

func TestSendSnapshotRecovery(t *testing.T) {
	// Verify that calling sendSnapshot after a failure still works
	origClient := liveClient
	defer func() { liveClient = origClient }()

	// First call fails
	liveClient = &obs.MockClient{Err: fmt.Errorf("down")}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	assert.False(t, sendSnapshot(w))
	buf.Reset()

	// Second call succeeds (simulates SigNoz recovery)
	liveClient = &obs.MockClient{Data: obs.DefaultMockData()}
	assert.True(t, sendSnapshot(w))
	assert.Contains(t, buf.String(), "event: update")
}

func TestLiveSSEStaleDetectionInTemplate(t *testing.T) {
	app := SetupApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	html := string(body)

	assert.Contains(t, html, "stale", "template should include stale state variable")
	assert.Contains(t, html, "try {", "template should include JSON.parse try/catch")
	assert.Contains(t, html, "Connection lost or data stale", "template should include stale banner")
}

// scanSSEEvent splits on double newlines (SSE event boundary).
func scanSSEEvent(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := strings.Index(string(data), "\n\n"); i >= 0 {
		return i + 2, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func TestScanSSEEvent(t *testing.T) {
	data := []byte("event: update\ndata: {\"test\":true}\n\nevent: other\n")
	advance, token, err := scanSSEEvent(data, false)
	assert.NoError(t, err)
	assert.Equal(t, "event: update\ndata: {\"test\":true}", string(token))
	assert.Equal(t, 35, advance)
	_ = bufio.NewScanner(strings.NewReader(""))
}
