package tests

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiveSectionHTML(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("live section exists", func(t *testing.T) {
		assert.Equal(t, 1, doc.Find("#live").Length())
	})

	t.Run("has Live Platform heading", func(t *testing.T) {
		heading := doc.Find("#live h2")
		assert.Equal(t, 1, heading.Length())
		assert.Contains(t, heading.Text(), "Live Platform")
	})

	t.Run("has Alpine.js x-data", func(t *testing.T) {
		section := doc.Find("#live")
		d, exists := section.Attr("x-data")
		assert.True(t, exists)
		assert.Contains(t, d, "liveDashboard")
	})

	t.Run("has EventSource", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "EventSource")
		assert.Contains(t, html, "/api/live/stream")
	})

	t.Run("has 6 category columns", func(t *testing.T) {
		// Find the outer grid (the one with lg:grid-cols-6)
		found := false
		doc.Find("#live .grid").Each(func(i int, s *goquery.Selection) {
			classes, _ := s.Attr("class")
			if strings.Contains(classes, "lg:grid-cols-6") {
				found = true
			}
		})
		assert.True(t, found, "should find a grid with lg:grid-cols-6")
	})

	t.Run("hosts column exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Hosts")
		assert.Contains(t, html, "talosoci-control-plane-legal-poodle")
		assert.Contains(t, html, "talosedge-genmachiche-flowing-bluejay")
	})

	t.Run("development column exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Development")
	})

	t.Run("deployment column exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Deployment")
		assert.Contains(t, html, "source-controller")
	})

	t.Run("runtime column exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Runtime")
		assert.Contains(t, html, "platform-website")
	})

	t.Run("data column exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Data")
	})

	t.Run("observability column exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Observability")
	})

	t.Run("services have data-live-svc", func(t *testing.T) {
		services := doc.Find("[data-live-svc]")
		assert.GreaterOrEqual(t, services.Length(), 8)
	})

	t.Run("services have status dots", func(t *testing.T) {
		dots := doc.Find("[data-dot]")
		assert.Equal(t, doc.Find("[data-live-svc]").Length(), dots.Length())
	})

	t.Run("services have CPU and RAM elements", func(t *testing.T) {
		cpu := doc.Find("[data-cpu]")
		ram := doc.Find("[data-ram]")
		assert.GreaterOrEqual(t, cpu.Length(), 1)
		assert.GreaterOrEqual(t, ram.Length(), 1)
	})

	t.Run("click-to-expand has toggleService", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "toggleService")
		assert.Contains(t, html, "selected")
	})

	t.Run("histogram panel exists", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "x-data-hist-cpu")
		assert.Contains(t, html, "x-data-hist-ram")
	})

	t.Run("static banner shows", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Showing platform topology")
	})

	t.Run("after architecture in DOM order", func(t *testing.T) {
		var foundArch, foundLive bool
		doc.Find("section[id]").Each(func(i int, s *goquery.Selection) {
			id, _ := s.Attr("id")
			if id == "architecture" {
				foundArch = true
			}
			if id == "live" && foundArch {
				foundLive = true
			}
		})
		assert.True(t, foundLive)
	})
}
