package handlers_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/internal/platform"
	"github.com/rezuscloud/platform-website/internal/server"
)

func TestHomeHandler(t *testing.T) {
	app := server.NewGatewayApp(platform.NewLocalRuntime())

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestHomeHandlerContainsExpectedContent(t *testing.T) {
	app := server.NewGatewayApp(platform.NewLocalRuntime())

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.Contains(t, html, "/apps/terminal")
	assert.Contains(t, html, "data-term-api")
	assert.Contains(t, html, "linux-panel")
	assert.Contains(t, html, "mac-panel")
}

func TestAppSurfaceRoutes(t *testing.T) {
	app := server.NewGatewayApp(platform.NewLocalRuntime())

	routes := []string{
		"/apps/terminal",
		"/apps/terminal/embed",
		"/apps/mac",
		"/apps/mac/embed",
		"/apps/linux",
		"/apps/linux/embed",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
		})
	}
}

func TestTerminalActionUpdatesSharedState(t *testing.T) {
	app := server.NewGatewayApp(platform.NewLocalRuntime())

	bootstrapReq := httptest.NewRequest("GET", "/", nil)
	bootstrapResp, err := app.Test(bootstrapReq, -1)
	require.NoError(t, err)
	defer bootstrapResp.Body.Close()

	cookies := bootstrapResp.Cookies()
	require.NotEmpty(t, cookies)

	actionReq := httptest.NewRequest("POST", "/apps/terminal/actions/run", strings.NewReader("preset=rezus+sync+demo"))
	actionReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		actionReq.AddCookie(cookie)
	}

	actionResp, err := app.Test(actionReq, -1)
	require.NoError(t, err)
	defer actionResp.Body.Close()

	body, err := io.ReadAll(actionResp.Body)
	require.NoError(t, err)
	html := string(body)

	assert.Contains(t, html, "linux-app")
	assert.Contains(t, html, "artifact.published")
	assert.Contains(t, actionResp.Header.Get("HX-Trigger"), "session-updated")

	macReq := httptest.NewRequest("GET", "/apps/mac/embed", nil)
	for _, cookie := range cookies {
		macReq.AddCookie(cookie)
	}
	macResp, err := app.Test(macReq, -1)
	require.NoError(t, err)
	defer macResp.Body.Close()

	macBody, err := io.ReadAll(macResp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(macBody), "Deployment dossier")

	shellReq := httptest.NewRequest("GET", "/shell/summary", nil)
	for _, cookie := range cookies {
		shellReq.AddCookie(cookie)
	}
	shellResp, err := app.Test(shellReq, -1)
	require.NoError(t, err)
	defer shellResp.Body.Close()

	shellBody, err := io.ReadAll(shellResp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(shellBody), "One command moved through three services")
}

func TestLegacySectionsRemoved(t *testing.T) {
	app := server.NewGatewayApp(platform.NewLocalRuntime())

	req := httptest.NewRequest("GET", "/sections/hero", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestHomePageContainsAllPrimarySurfaces(t *testing.T) {
	app := server.NewGatewayApp(platform.NewLocalRuntime())

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.True(t, strings.Contains(html, `id="terminal-panel"`))
	assert.True(t, strings.Contains(html, `id="mac-panel"`))
	assert.True(t, strings.Contains(html, `id="linux-panel"`))
	assert.True(t, strings.Contains(html, `data-scene-root`))
	assert.True(t, strings.Contains(html, `xterm-mount`))
	assert.True(t, strings.Contains(html, `snap-track`))
}
