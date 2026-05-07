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

	t.Run("has Alpine.js x-data", func(t *testing.T) {
		section := doc.Find("#live")
		d, exists := section.Attr("x-data")
		assert.True(t, exists)
		assert.Contains(t, d, "liveMap")
	})

	t.Run("has EventSource", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "EventSource")
		assert.Contains(t, html, "/api/live/stream")
	})

	// Service map layer tests
	t.Run("has 7 layers", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Traffic")
		assert.Contains(t, html, "Application")
		assert.Contains(t, html, "Runtime")
		assert.Contains(t, html, "Delivery")
		assert.Contains(t, html, "Observability")
		assert.Contains(t, html, "Storage")
		assert.Contains(t, html, "Infrastructure")
	})

	t.Run("traffic layer has Visitor and Gateway", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Visitor")
		assert.Contains(t, html, "Gateway")
		assert.Contains(t, html, "→")
	})

	t.Run("infrastructure layer has real hosts", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "OCI Cloud")
		assert.Contains(t, html, "Edge Node")
	})

	t.Run("live nodes have data-map-key", func(t *testing.T) {
		nodes := doc.Find("[data-map-key]")
		assert.GreaterOrEqual(t, nodes.Length(), 8, "should have at least 8 live nodes with data-map-key")
	})

	t.Run("live nodes have status dots", func(t *testing.T) {
		dots := doc.Find("[data-map-key] [data-dot]")
		liveNodes := doc.Find("[data-map-key]")
		assert.Equal(t, liveNodes.Length(), dots.Length(), "each live node should have a status dot")
	})

	t.Run("live nodes have metric elements", func(t *testing.T) {
		metrics := doc.Find("[data-map-key] [data-metric]")
		assert.GreaterOrEqual(t, metrics.Length(), 1, "should have at least 1 node with metrics")
	})

	t.Run("click-to-expand has selectNode", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "selectNode")
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
