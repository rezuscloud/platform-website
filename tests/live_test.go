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

	// Category columns

	t.Run("has 5 category columns", func(t *testing.T) {
		cats := doc.Find("[data-live-category]")
		assert.Equal(t, 5, cats.Length())
	})

	t.Run("infrastructure column renders", func(t *testing.T) {
		col := doc.Find("[data-live-category=\"infra\"]")
		assert.Equal(t, 1, col.Length())
		assert.Contains(t, col.Text(), "Infrastructure")
	})

	t.Run("development column renders", func(t *testing.T) {
		col := doc.Find("[data-live-category=\"dev\"]")
		assert.Equal(t, 1, col.Length())
		assert.Contains(t, col.Text(), "Forgejo")
		assert.Contains(t, col.Text(), "ARC Controller")
	})

	t.Run("delivery column renders", func(t *testing.T) {
		col := doc.Find("[data-live-category=\"delivery\"]")
		assert.Equal(t, 1, col.Length())
		assert.Contains(t, col.Text(), "Flux Source")
		assert.Contains(t, col.Text(), "KubeVela")
		assert.Contains(t, col.Text(), "Cert Manager")
	})

	t.Run("runtime column renders", func(t *testing.T) {
		col := doc.Find("[data-live-category=\"runtime\"]")
		assert.Equal(t, 1, col.Length())
		assert.Contains(t, col.Text(), "Cilium CNI")
		assert.Contains(t, col.Text(), "platform-website")
		assert.Contains(t, col.Text(), "Dapr Sidecar")
	})

	t.Run("observability column renders", func(t *testing.T) {
		col := doc.Find("[data-live-category=\"observability\"]")
		assert.Equal(t, 1, col.Length())
		assert.Contains(t, col.Text(), "SigNoz Collector")
		assert.Contains(t, col.Text(), "ClickHouse")
	})

	// All services

	t.Run("all services render", func(t *testing.T) {
		services := doc.Find("[data-live-service]")
		assert.GreaterOrEqual(t, services.Length(), 15)
	})

	// Status dots

	t.Run("healthy services show green dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"platform-website\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-green-600")
	})

	t.Run("unmonitored services show hollow dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"forgejo\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "border border-rule")
	})

	t.Run("infrastructure nodes show accent dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"oci-cloud\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-accent-gold")
	})

	// Static banner

	t.Run("static banner shows when no metrics", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Showing platform topology")
	})

	// Responsive grid

	t.Run("uses responsive grid classes", func(t *testing.T) {
		grid := doc.Find("[data-live-infra]")
		classes, _ := grid.Attr("class")
		assert.Contains(t, classes, "lg:grid-cols-5")
		assert.Contains(t, classes, "sm:grid-cols-2")
	})
}
