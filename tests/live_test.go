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
		assert.Equal(t, 3, nodes.Length(), "Expected 3 nodes from default data")
	})

	t.Run("node shows name", func(t *testing.T) {
		talosCP0 := doc.Find("[data-live-node=\"talos-oci-cp-0\"]")
		assert.Equal(t, 1, talosCP0.Length())
		assert.Contains(t, talosCP0.Text(), "talos-oci-cp-0")
	})

	t.Run("node shows tier label", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talos-oci-cp-0\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "oci-cloud")
	})

	t.Run("edge node shows edge tier", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talosedge-genmachiche-flowing-bluejay\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "edge")
	})

	t.Run("node shows resource usage", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talos-oci-cp-0\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "12%")
		assert.Contains(t, html, "4.2 GiB")
	})

	t.Run("node shows status indicator", func(t *testing.T) {
		html, err := doc.Find("[data-live-node=\"talos-oci-cp-0\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-green-600")
	})

	t.Run("pods render inside nodes", func(t *testing.T) {
		edgeNode := doc.Find("[data-live-node=\"talosedge-genmachiche-flowing-bluejay\"]")
		pods := edgeNode.Find("[data-live-pod]")
		assert.Equal(t, 3, pods.Length(), "Expected 3 pods on edge node")
	})

	t.Run("pod shows name and status", func(t *testing.T) {
		pod := doc.Find("[data-live-pod=\"platform-website-6f8d6fd5fc-4rgj9\"]")
		assert.Equal(t, 1, pod.Length())
		assert.Contains(t, pod.Text(), "Running")
	})
}
