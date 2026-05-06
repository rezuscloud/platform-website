package tests

import (
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

	t.Run("has LIVE indicator", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "animate-pulse")
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

	// Alpine.js

	t.Run("has x-data for live dashboard", func(t *testing.T) {
		section := doc.Find("#live")
		d, exists := section.Attr("x-data")
		assert.True(t, exists)
		assert.Contains(t, d, "liveDashboard")
	})

	t.Run("has x-init", func(t *testing.T) {
		section := doc.Find("#live")
		_, exists := section.Attr("x-init")
		assert.True(t, exists)
	})

	t.Run("has EventSource script", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "EventSource")
		assert.Contains(t, html, "/api/live/stream")
	})

	// Services

	t.Run("all 17 services render", func(t *testing.T) {
		services := doc.Find("[data-live-service]")
		assert.Equal(t, 17, services.Length())
	})

	// Status dots

	t.Run("healthy services have green dots with data-dot", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"platform-website\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "data-dot")
		assert.Contains(t, html, "bg-green-600")
	})

	t.Run("unmonitored services have hollow dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"arc-controller\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "data-dot")
		assert.Contains(t, html, "border border-rule")
	})

	t.Run("infrastructure nodes have accent dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"oci-cloud\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-accent-gold")
	})

	// Dynamic data attributes

	t.Run("services have data-metric element", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"platform-website\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "data-metric")
	})

	t.Run("services have data-memory element", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"platform-website\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "data-memory")
	})

	// Static banner

	t.Run("static banner shows when no metrics", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Showing platform topology")
	})

	// Responsive grid

	t.Run("uses responsive grid classes", func(t *testing.T) {
		grid := doc.Find("#live .grid")
		classes, _ := grid.Attr("class")
		assert.Contains(t, classes, "lg:grid-cols-5")
		assert.Contains(t, classes, "sm:grid-cols-2")
	})
}
