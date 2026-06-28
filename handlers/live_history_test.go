package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/obs"
)

// withLiveClient swaps the package liveClient for the test and restores it.
func withLiveClient(t *testing.T, c obs.Client) {
	t.Helper()
	orig := liveClient
	liveClient = c
	t.Cleanup(func() { liveClient = orig })
}

func TestLiveServiceHistory_MissingParams(t *testing.T) {
	app := SetupApp()

	cases := []struct{ name, q string }{
		{"no params", ""},
		{"missing host", "?namespace=flux-system&name=source-controller"},
		{"missing name", "?namespace=flux-system&host=node-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/live/history"+tc.q, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestLiveServiceHistory_NotFound(t *testing.T) {
	withLiveClient(t, &obs.MockClient{Data: obs.LiveData{
		Services: []obs.Service{
			{Name: "source-controller", Namespace: "flux-system", Host: "node-1"},
		},
	}})
	app := SetupApp()

	req := httptest.NewRequest("GET",
		"/api/live/history?namespace=flux-system&name=absent&host=node-1", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestLiveServiceHistory_ReturnsHistoryOnDemand(t *testing.T) {
	withLiveClient(t, &obs.MockClient{Data: obs.LiveData{
		Services: []obs.Service{{
			Name: "source-controller", Namespace: "flux-system", Host: "node-1",
			CPU: 0.5, RAM: 120,
			CPUHist:  "0,0 1,1 2,0.5",
			RAMHist:  "0,0 1,1 2,1",
			NetHist:  "0,0 1,1 2,0",
			DiskHist: "0,0 1,0 2,1",
		}},
	}})
	app := SetupApp()

	req := httptest.NewRequest("GET",
		"/api/live/history?namespace=flux-system&name=source-controller&host=node-1", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// History must be served on demand (200), with the four polyline strings.
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, fiber.MIMEApplicationJSON, resp.Header.Get("Content-Type"))
	body := readBody(t, resp)
	assert.Contains(t, body, `"cpuHist":"0,0 1,1 2,0.5"`)
	assert.Contains(t, body, `"ramHist":"0,0 1,1 2,1"`)
	assert.Contains(t, body, `"namespace":"flux-system"`)
	assert.Contains(t, body, `"name":"source-controller"`)
}

func TestLiveServiceHistory_NeverCached(t *testing.T) {
	withLiveClient(t, &obs.MockClient{Data: obs.LiveData{
		Services: []obs.Service{
			{Name: "x", Namespace: "ns", Host: "h"},
		},
	}})
	app := SetupApp()
	req := httptest.NewRequest("GET", "/api/live/history?namespace=ns&name=x&host=h", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(b)
}
