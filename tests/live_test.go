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

	t.Run("has Live Infrastructure heading", func(t *testing.T) {
		heading := doc.Find("#live h2")
		assert.Equal(t, 1, heading.Length())
		assert.Contains(t, heading.Text(), "Live Infrastructure")
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

	// SVG diagram

	t.Run("has SVG diagram", func(t *testing.T) {
		svgs := doc.Find("#live svg")
		assert.Equal(t, 1, svgs.Length())
	})

	t.Run("SVG has aria-label", func(t *testing.T) {
		svg := doc.Find("#live svg")
		label, exists := svg.Attr("aria-label")
		assert.True(t, exists)
		assert.Contains(t, label, "Service dependency")
	})

	t.Run("SVG has 5 service nodes", func(t *testing.T) {
		services := doc.Find("[data-live-service]")
		assert.Equal(t, 5, services.Length())
	})

	t.Run("cilium-gateway renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"cilium-gateway\"]")
		assert.Equal(t, 1, svc.Length())
	})

	t.Run("platform-website renders with hero border", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		assert.Equal(t, 1, svc.Length())
		// Should have accent-colored border (hero treatment)
		html, err := svc.Html()
		require.NoError(t, err)
		assert.Contains(t, html, "stroke-accent-gold")
	})

	t.Run("daprd renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"daprd\"]")
		assert.Equal(t, 1, svc.Length())
	})

	t.Run("signoz-collector renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"signoz-collector\"]")
		assert.Equal(t, 1, svc.Length())
	})

	t.Run("dapr-control-plane renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"dapr-control-plane\"]")
		assert.Equal(t, 1, svc.Length())
	})

	// Lane labels

	t.Run("lane labels render", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "INGRESS")
		assert.Contains(t, html, "APPLICATION")
		assert.Contains(t, html, "SIDECAR")
		assert.Contains(t, html, "INFRA")
	})

	// Edge labels

	t.Run("edge labels render", func(t *testing.T) {
		html, err := doc.Find("[data-live-infra]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "HTTPS")
		assert.Contains(t, html, "localhost")
		assert.Contains(t, html, "OTLP")
		assert.Contains(t, html, "gRPC")
	})

	// Static banner (mock data = no SigNoz)

	t.Run("static banner shows when no metrics", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Showing static topology")
		assert.Contains(t, html, "require SigNoz")
	})

	// No health strip in mock mode

	t.Run("no health strip in static mode", func(t *testing.T) {
		health := doc.Find("[data-live-health]")
		assert.Equal(t, 0, health.Length())
	})

	// SVG has arrow markers

	t.Run("edges have arrow polygons", func(t *testing.T) {
		html, err := doc.Find("[data-live-infra]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "polygon")
	})
}
