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
		selection := doc.Find("#live")
		assert.Equal(t, 1, selection.Length(), "Expected exactly one #live section")
	})

	t.Run("live section has Watch It Run heading", func(t *testing.T) {
		heading := doc.Find("#live h2")
		assert.Equal(t, 1, heading.Length())
		assert.Contains(t, heading.Text(), "Watch It Run")
	})

	t.Run("live section has LIVE indicator", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Live")
		assert.Contains(t, html, "animate-pulse")
	})

	t.Run("live section has green dot", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-green-600")
		assert.Contains(t, html, "dark:bg-green-500")
	})

	t.Run("live section after architecture", func(t *testing.T) {
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
		assert.True(t, foundLive, "#live should appear after #architecture in DOM order")
	})

	t.Run("nodes render as containers", func(t *testing.T) {
		nodes := doc.Find("[data-live-node]")
		assert.Equal(t, 2, nodes.Length(), "Expected 2 nodes from default data")
	})

	t.Run("node shows name", func(t *testing.T) {
		talosOCI := doc.Find("[data-live-node=\"talosoci-control-plane-legal-poodle\"]")
		assert.Equal(t, 1, talosOCI.Length())
		assert.Contains(t, talosOCI.Text(), "talosoci-control-plane-legal-poodle")
	})

	t.Run("node shows tier label", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talosoci-control-plane-legal-poodle\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "oci-cloud")
	})

	t.Run("edge node shows edge tier", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talosedge-genmachiche-flowing-bluejay\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "edge")
	})

	t.Run("node shows resource usage", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talosoci-control-plane-legal-poodle\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "12%")
		assert.Contains(t, html, "4.2 GiB")
	})

	t.Run("node shows status indicator", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talosoci-control-plane-legal-poodle\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-green-600")
	})

	t.Run("pods render inside nodes", func(t *testing.T) {
		edgeNode := doc.Find("[data-live-node=\"talosedge-genmachiche-flowing-bluejay\"]")
		pods := edgeNode.Find("[data-live-pod]")
		assert.Equal(t, 5, pods.Length(), "Expected 5 pods on edge node")
	})

	t.Run("pod shows name and status", func(t *testing.T) {
		pod := doc.Find("[data-live-pod=\"platform-website-69f7bffd5f-trltm\"]")
		assert.Equal(t, 1, pod.Length())
		assert.Contains(t, pod.Text(), "Running")
	})

	// Sparkline tests

	t.Run("metrics container exists", func(t *testing.T) {
		metrics := doc.Find("[data-live-metrics]")
		assert.Equal(t, 1, metrics.Length())
	})

	t.Run("three metric sparklines render", func(t *testing.T) {
		metrics := doc.Find("[data-live-metric]")
		assert.Equal(t, 3, metrics.Length())
	})

	t.Run("metrics show label and value", func(t *testing.T) {
		req := doc.Find("[data-live-metric=\"Requests\"]")
		assert.Equal(t, 1, req.Length())
		assert.Contains(t, req.Text(), "1,247")
		assert.Contains(t, req.Text(), "req/min")
	})

	t.Run("metrics have SVG sparkline", func(t *testing.T) {
		metric := doc.Find("[data-live-metric=\"Requests\"]")
		svg := metric.Find("svg")
		assert.Equal(t, 1, svg.Length())

		polyline := svg.Find("polyline")
		assert.Equal(t, 1, polyline.Length())
		points, exists := polyline.Attr("points")
		assert.True(t, exists)
		assert.NotEmpty(t, points)

		polygon := svg.Find("polygon")
		assert.Equal(t, 1, polygon.Length())
		areaPoints, exists := polygon.Attr("points")
		assert.True(t, exists)
		assert.NotEmpty(t, areaPoints)
	})

	t.Run("sparkline uses accent colors", func(t *testing.T) {
		metric := doc.Find("[data-live-metric=\"Requests\"]")
		html, err := metric.Html()
		require.NoError(t, err)
		assert.Contains(t, html, "stroke-accent-gold")
		assert.Contains(t, html, "dark:stroke-next-teal")
		assert.Contains(t, html, "fill-accent-gold")
		assert.Contains(t, html, "dark:fill-next-teal")
	})

	t.Run("layout uses grid with 2/3 and 1/3 split", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "lg:col-span-2")
	})
}
