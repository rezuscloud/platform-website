package tests

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/handlers"
)

func setupIntegrationApp() *fiber.App {
	app := fiber.New(fiber.Config{})
	app.Get("/", handlers.Home)
	app.Get("/sections/:name", handlers.Section)
	return app
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
		assert.Contains(t, title, "Terminal")
	})

	t.Run("has meta description for retro scene", func(t *testing.T) {
		description, exists := doc.Find("meta[name='description']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, description, "phosphor terminal")
		assert.Contains(t, description, "Macintosh System 1")
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

	t.Run("has homepage scene root and track", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("#scene[data-scene-root]").Length())
		assert.Equal(t, 1, doc.Find("[data-scene-track]").Length())
		assert.Equal(t, 1, doc.Find("[data-scene-camera]").Length())
		assert.Equal(t, 1, doc.Find("[data-scene-world]").Length())
	})

	t.Run("has nested scene targets", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find(`[data-scene-target='terminal']`).Length())
		assert.Equal(t, 1, doc.Find(`[data-scene-target='mac']`).Length())
		assert.Equal(t, 1, doc.Find(`[data-scene-target='linux']`).Length())
	})

	t.Run("has scene script", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find(`script[src='/assets/js/scene.js']`).Length())
	})

	t.Run("does not render nav or footer chrome on home", func(t *testing.T) {
		assert.Equal(t, 0, doc.Find("nav").Length())
		assert.Equal(t, 0, doc.Find("footer").Length())
	})
}

func TestHomePageSceneContent(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")
	html := getHTMLString(t, app, "/")

	t.Run("scene section exists", func(t *testing.T) {
		selection := doc.Find("#scene")
		assert.Equal(t, 1, selection.Length())
		assert.Greater(t, len(strings.TrimSpace(selection.Text())), 20)
	})

	t.Run("contains terminal manifesto and boot output", func(t *testing.T) {
		assert.Contains(t, html, "REZUS OS ROM 1.0.0")
		assert.Contains(t, html, "64K SYSTEM RAM PASSED")
		assert.Contains(t, html, "your machines")
		assert.Contains(t, html, "your network")
		assert.Contains(t, html, "your rules")
	})

	t.Run("contains macintosh and linux world details", func(t *testing.T) {
		assert.Contains(t, html, "Mini vMac")
		assert.Contains(t, html, "MacTerminal")
		assert.Contains(t, html, "xterm")
		assert.Contains(t, html, "xclock")
	})
}

func TestHomePageNoWebsiteChrome(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")
	html := getHTMLString(t, app, "/")

	t.Run("main landmark exists", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("main").Length())
	})

	t.Run("no theme toggle or Alpine state on home", func(t *testing.T) {
		assert.Equal(t, 0, doc.Find(`button[aria-label='Toggle theme']`).Length())
		assert.NotContains(t, html, "$store.theme")
		assert.NotContains(t, html, "alpine.min.js")
	})

	t.Run("no htmx script on home", func(t *testing.T) {
		assert.Equal(t, 0, doc.Find("script[src*='htmx']").Length())
	})
}

func TestAccessibilityHTML(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("has main landmark", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("main").Length())
	})

	t.Run("has hidden semantic copy for assistive tech", func(t *testing.T) {
		srCopy := doc.Find(".scene-sr-copy")
		assert.Equal(t, 1, srCopy.Length())
		assert.Contains(t, srCopy.Text(), "RezusCloud")
	})

	t.Run("links have discernible text", func(t *testing.T) {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			ariaLabel, hasAriaLabel := s.Attr("aria-label")
			assert.True(t, len(text) > 0 || hasAriaLabel,
				"Link should have text or aria-label, got text: '%s', aria-label: '%s'",
				text, ariaLabel)
		})
	})
}

func TestSectionEndpointsRemainValidForHTMX(t *testing.T) {
	app := setupIntegrationApp()

	t.Run("section endpoints return valid HTML", func(t *testing.T) {
		sections := []string{"hero", "features", "architecture", "getstarted"}

		for _, section := range sections {
			doc := getHTMLDoc(t, app, "/sections/"+section)
			assert.Equal(t, 1, doc.Find("#"+section).Length(),
				"Section endpoint should return element with id '%s'", section)
		}
	})

	t.Run("sections contain expected structure", func(t *testing.T) {
		doc := getHTMLDoc(t, app, "/sections/hero")
		assert.Greater(t, doc.Find("h1, h2, h3").Length(), 0)
	})

	t.Run("section endpoints could be used with hx-get", func(t *testing.T) {
		sections := []string{"hero", "features", "architecture", "networking", "edge", "services", "comparison", "usecases", "techstack", "getstarted"}

		for _, section := range sections {
			doc := getHTMLDoc(t, app, "/sections/"+section)
			assert.Equal(t, 1, doc.Find("#"+section).Length(),
				"Section %s endpoint should return single element for hx-swap", section)

			content := doc.Find("#" + section).Text()
			assert.Greater(t, len(strings.TrimSpace(content)), 10,
				"Section %s should have content suitable for HTMX swap", section)
		}
	})
}

func TestProgressiveEnhancement(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")
	html := getHTMLString(t, app, "/")

	t.Run("page works without JavaScript", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("#scene").Length())
		assert.Equal(t, 1, doc.Find(`[data-scene-target='terminal']`).Length())
		assert.Equal(t, 1, doc.Find(`[data-scene-target='mac']`).Length())
		assert.Equal(t, 1, doc.Find(`[data-scene-target='linux']`).Length())
	})

	t.Run("scene scroll track exists server-side", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("[data-scene-track]").Length())
	})

	t.Run("content visible before JavaScript loads", func(t *testing.T) {
		assert.Contains(t, html, "REZUS OS ROM 1.0.0")
		assert.Contains(t, html, "Mini vMac")
		assert.Contains(t, html, "xterm")
	})

	t.Run("no javascript fallback class present", func(t *testing.T) {
		assert.Contains(t, html, "scene-html no-js")
	})
}
