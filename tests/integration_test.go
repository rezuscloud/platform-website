package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/internal/platform"
	"github.com/rezuscloud/platform-website/internal/server"
)

func setupIntegrationApp() *fiber.App {
	return server.NewGatewayApp(platform.NewLocalRuntime())
}

func getHTMLDoc(t *testing.T, app *fiber.App, path string) *goquery.Document {
	req := httptest.NewRequest("GET", path, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	require.NoError(t, err)

	return doc
}

func getHTMLString(t *testing.T, app *fiber.App, path string) string {
	req := httptest.NewRequest("GET", path, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}

func TestHomePageHTMLStructure(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")
	html := getHTMLString(t, app, "/")

	t.Run("has valid HTML5 doctype", func(t *testing.T) {
		assert.True(t, strings.Contains(html, "<!DOCTYPE html>") || strings.Contains(html, "<html"))
	})

	t.Run("has title containing RezusCloud", func(t *testing.T) {
		title := doc.Find("title").Text()
		assert.Contains(t, title, "RezusCloud")
		assert.Contains(t, title, "Live Dapr shell")
	})

	t.Run("has meta description for shell composition", func(t *testing.T) {
		description, exists := doc.Find("meta[name='description']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, description, "terminal, Macintosh, and Linux")
		assert.Contains(t, description, "shared state")
	})

	t.Run("has canonical link", func(t *testing.T) {
		canonical, exists := doc.Find("link[rel='canonical']").Attr("href")
		assert.True(t, exists)
		assert.Equal(t, "https://rezus.cloud/", canonical)
	})

	t.Run("has structured data", func(t *testing.T) {
		structuredData := doc.Find("script[type='application/ld+json']").Text()
		assert.Contains(t, structuredData, `"@type": "WebSite"`)
		assert.Contains(t, structuredData, `"@type": "Organization"`)
		assert.Contains(t, structuredData, `"email": "tiberiu@rezus.net"`)
	})

	t.Run("has shell summary and app surfaces", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("#terminal-panel").Length())
		assert.Equal(t, 1, doc.Find("#mac-panel").Length())
		assert.Equal(t, 1, doc.Find("#linux-panel").Length())
		assert.Equal(t, 1, doc.Find("#linux-panel #mac-panel #terminal-panel").Length())
		assert.Equal(t, 1, doc.Find("[data-scene-root]").Length())
		assert.Equal(t, 1, doc.Find("#xterm-mount").Length())
	})

	t.Run("has htmx and scene scripts", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find(`script[src='/assets/js/htmx.min.js']`).Length())
		assert.Equal(t, 1, doc.Find(`script[src='/assets/js/scene.js']`).Length())
	})
}

func TestHomePageShellContent(t *testing.T) {
	app := setupIntegrationApp()
	html := getHTMLString(t, app, "/")

	t.Run("contains route links and nested surfaces", func(t *testing.T) {
		assert.Contains(t, html, "/apps/terminal")
		assert.Contains(t, html, "/apps/mac")
		assert.Contains(t, html, "/apps/linux")
	})

	t.Run("contains terminal boot data and nested panels", func(t *testing.T) {
		assert.Contains(t, html, "data-term-api")
		assert.Contains(t, html, "Mini vMac")
		assert.Contains(t, html, "idle")
		assert.Contains(t, html, "Artifact drawer empty")
	})
}

func TestStandaloneAppPages(t *testing.T) {
	app := setupIntegrationApp()

	tests := []struct {
		path           string
		expected       string
		hasLinux       bool
		hasMac         bool
		hasTerminal    bool
		nestedMac      bool
		nestedTerminal bool
	}{
		{path: "/apps/terminal", expected: "Command surface", hasTerminal: true},
		{path: "/apps/mac", expected: "Inspection surface", hasMac: true, hasTerminal: true, nestedTerminal: true},
		{path: "/apps/linux", expected: "Execution surface", hasLinux: true, hasMac: true, hasTerminal: true, nestedMac: true, nestedTerminal: true},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			doc := getHTMLDoc(t, app, tc.path)
			html := getHTMLString(t, app, tc.path)
			assert.Contains(t, html, tc.expected)
			assert.Contains(t, html, "RezusCloud")
			assert.Equal(t, boolToInt(tc.hasLinux), doc.Find("#linux-panel").Length())
			assert.Equal(t, boolToInt(tc.hasMac), doc.Find("#mac-panel").Length())
			assert.Equal(t, boolToInt(tc.hasTerminal), doc.Find("#terminal-panel").Length())
			if tc.nestedMac {
				assert.Equal(t, 1, doc.Find("#linux-panel #mac-panel").Length())
			}
			if tc.nestedTerminal {
				assert.Equal(t, 1, doc.Find("#mac-panel #terminal-panel").Length())
			}
		})
	}
}

func TestSharedStateFlowAcrossSurfaces(t *testing.T) {
	app := setupIntegrationApp()

	bootstrapReq := httptest.NewRequest("GET", "/", nil)
	bootstrapResp, err := app.Test(bootstrapReq, -1)
	require.NoError(t, err)
	defer bootstrapResp.Body.Close()

	cookies := bootstrapResp.Cookies()
	require.NotEmpty(t, cookies)

	actionReq := httptest.NewRequest("POST", "/apps/terminal/actions/run", strings.NewReader("preset=rezus+fanout+edge"))
	actionReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	actionReq.Header.Set("HX-Request", "true")
	for _, cookie := range cookies {
		actionReq.AddCookie(cookie)
	}

	actionResp, err := app.Test(actionReq, -1)
	require.NoError(t, err)
	defer actionResp.Body.Close()

	body, err := io.ReadAll(actionResp.Body)
	require.NoError(t, err)
	html := string(body)
	assert.Contains(t, html, "edge.shifted")
	assert.Contains(t, html, "JetStream")
	assert.Contains(t, actionResp.Header.Get("HX-Trigger"), "session-updated")

	macReq := httptest.NewRequest("GET", "/apps/mac/embed", nil)
	for _, cookie := range cookies {
		macReq.AddCookie(cookie)
	}
	macHTML := getHTMLFromRequest(t, app, macReq)
	assert.Contains(t, macHTML, "Edge failover memo")

	linuxReq := httptest.NewRequest("GET", "/apps/linux/embed", nil)
	for _, cookie := range cookies {
		linuxReq.AddCookie(cookie)
	}
	linuxHTML := getHTMLFromRequest(t, app, linuxReq)
	assert.Contains(t, linuxHTML, "edge-active")

	shellReq := httptest.NewRequest("GET", "/shell/summary", nil)
	for _, cookie := range cookies {
		shellReq.AddCookie(cookie)
	}
	shellHTML := getHTMLFromRequest(t, app, shellReq)
	assert.Contains(t, shellHTML, "Linux shifted edge state and broadcast the change")
	assert.Contains(t, shellHTML, "fanout")
	assert.Contains(t, shellHTML, "Redis")
}

func TestShellDaprSubscriptionUsesConfiguredPubsub(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/dapr/subscribe", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	html := string(body)
	assert.Contains(t, html, "pubsub")
	assert.Contains(t, html, "homepage.events")
}

func TestShellObserverEndpointAcceptsSessionEvent(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("POST", "/events/shell", strings.NewReader(`{"sessionId":"session-1","type":"artifact.published","source":"platform-website-linux","message":"published"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestLegacySectionsRemoved(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/sections/hero", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestProgressiveEnhancement(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")
	html := getHTMLString(t, app, "/")

	t.Run("page works without client-side execution", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("#terminal-panel").Length())
		assert.Equal(t, 1, doc.Find("#mac-panel").Length())
		assert.Equal(t, 1, doc.Find("#linux-panel").Length())
		assert.Equal(t, 1, doc.Find("#linux-panel #mac-panel #terminal-panel").Length())
		assert.Equal(t, 1, doc.Find("#xterm-mount").Length())
	})

	t.Run("top level routes expose htmx endpoints server-side", func(t *testing.T) {
		macHTML := getHTMLString(t, app, "/apps/mac")
		linuxHTML := getHTMLString(t, app, "/apps/linux")
		terminalHTML := getHTMLString(t, app, "/apps/terminal")
		assert.Contains(t, html, `hx-get="/apps/linux/embed"`)
		assert.Contains(t, macHTML, `hx-get="/apps/mac/embed"`)
		assert.Contains(t, linuxHTML, `hx-get="/apps/linux/embed"`)
		assert.Contains(t, terminalHTML, `hx-get="/apps/terminal/embed"`)
	})

	t.Run("no javascript bootstrap class present", func(t *testing.T) {
		assert.Contains(t, html, "brand-html no-js")
	})
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

func getHTMLFromRequest(t *testing.T, app *fiber.App, req *http.Request) string {
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}
