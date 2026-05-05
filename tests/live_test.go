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

	// Service map: tier containers

	t.Run("tier containers render", func(t *testing.T) {
		tiers := doc.Find("[data-tier]")
		assert.Equal(t, 2, tiers.Length(), "Expected 2 tier containers")
	})

	t.Run("OCI Cloud tier exists", func(t *testing.T) {
		html, err := doc.Find("[data-tier=\"oci-cloud\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "OCI Cloud")
	})

	t.Run("Edge tier exists", func(t *testing.T) {
		html, err := doc.Find("[data-tier=\"edge\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Edge")
	})

	t.Run("tier headers use inverted background", func(t *testing.T) {
		html, err := doc.Find("[data-tier=\"oci-cloud\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-ink")
	})

	// Service map: nodes

	t.Run("nodes render inside tiers", func(t *testing.T) {
		nodes := doc.Find("[data-live-node]")
		assert.Equal(t, 2, nodes.Length(), "Expected 2 nodes")
	})

	t.Run("OCI node name renders", func(t *testing.T) {
		node := doc.Find("[data-live-node=\"talosoci-control-plane-legal-poodle\"]")
		assert.Equal(t, 1, node.Length())
	})

	t.Run("edge node name renders", func(t *testing.T) {
		node := doc.Find("[data-live-node=\"talosedge-genmachiche-flowing-bluejay\"]")
		assert.Equal(t, 1, node.Length())
	})

	// Service map: pods as blocks

	t.Run("pods render as blocks", func(t *testing.T) {
		pods := doc.Find("[data-live-pod]")
		assert.Equal(t, 11, pods.Length(), "Expected 11 pods total")
	})

	t.Run("pods show namespace", func(t *testing.T) {
		html, err := doc.Find("[data-live-pod]").First().Html()
		require.NoError(t, err)
		assert.Contains(t, html, "kube-system")
	})

	// Sparklines

	t.Run("metrics container exists", func(t *testing.T) {
		metrics := doc.Find("[data-live-metrics]")
		assert.Equal(t, 1, metrics.Length())
	})

	t.Run("three metric sparklines render", func(t *testing.T) {
		metrics := doc.Find("[data-live-metric]")
		assert.Equal(t, 3, metrics.Length())
	})

	t.Run("sparkline uses accent colors", func(t *testing.T) {
		metric := doc.Find("[data-live-metric=\"Requests\"]")
		html, err := metric.Html()
		require.NoError(t, err)
		assert.Contains(t, html, "stroke-accent-gold")
		assert.Contains(t, html, "dark:stroke-next-teal")
	})

	t.Run("layout uses grid with 2/3 split", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "lg:col-span-2")
	})
}
