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

	t.Run("has Watch It Run heading", func(t *testing.T) {
		heading := doc.Find("#live h2")
		assert.Equal(t, 1, heading.Length())
		assert.Contains(t, heading.Text(), "Watch It Run")
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

	// Terminal window

	t.Run("has terminal title bar", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "service-map")
	})

	t.Run("has VT323 terminal body", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "font-terminal")
	})

	// Service tree: all 5 services render

	t.Run("five services render", func(t *testing.T) {
		services := doc.Find("[data-live-service]")
		assert.Equal(t, 5, services.Length())
	})

	t.Run("cilium gateway renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"cilium-gateway\"]")
		assert.Equal(t, 1, svc.Length())
		assert.Contains(t, svc.Text(), "Cilium Gateway")
	})

	t.Run("platform-website renders as hero", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		assert.Equal(t, 1, svc.Length())
		assert.Contains(t, svc.Text(), "Go / Fiber v2")
	})

	t.Run("daprd renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"daprd\"]")
		assert.Equal(t, 1, svc.Length())
		assert.Contains(t, svc.Text(), "Dapr Sidecar")
	})

	t.Run("signoz-collector renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"signoz-collector\"]")
		assert.Equal(t, 1, svc.Length())
		assert.Contains(t, svc.Text(), "SigNoz Collector")
	})

	t.Run("dapr-control-plane renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"dapr-control-plane\"]")
		assert.Equal(t, 1, svc.Length())
		assert.Contains(t, svc.Text(), "Dapr Control Plane")
	})

	// Tree structure: edges with labels

	t.Run("edge labels show connections", func(t *testing.T) {
		html, err := doc.Find("[data-live-infra]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "HTTPS")
		assert.Contains(t, html, "localhost")
		assert.Contains(t, html, "OTLP")
		assert.Contains(t, html, "gRPC")
	})

	// Tree indicators

	t.Run("tree has branch indicators", func(t *testing.T) {
		html, err := doc.Find("[data-live-infra]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "├─")
	})

	// Status dots

	t.Run("healthy services show green dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"platform-website\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "text-green-600")
	})

	// Sparklines

	t.Run("platform-website has metrics", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		sparklines := svc.Find("[data-live-metric]")
		assert.Equal(t, 2, sparklines.Length())
	})

	t.Run("sparklines are SVG polylines", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		svgs := svc.Find("svg")
		assert.Equal(t, 2, svgs.Length())
	})

	t.Run("sparklines use dual accent colors", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		html, err := svc.Html()
		require.NoError(t, err)
		assert.Contains(t, html, "stroke-accent-gold")
		assert.Contains(t, html, "dark:stroke-next-teal")
	})

	t.Run("daprd has components metric", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"daprd\"]")
		sparklines := svc.Find("[data-live-metric]")
		assert.Equal(t, 1, sparklines.Length())
	})

	// Metrics show readable values

	t.Run("metrics show values with units", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		html, err := svc.Html()
		require.NoError(t, err)
		assert.Contains(t, html, "MiB")
	})
}
