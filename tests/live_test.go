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

	// Service map topology

	t.Run("five services render", func(t *testing.T) {
		services := doc.Find("[data-live-service]")
		assert.Equal(t, 5, services.Length())
	})

	t.Run("cilium gateway renders", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"cilium-gateway\"]")
		assert.Equal(t, 1, svc.Length())
		assert.Contains(t, svc.Text(), "Cilium Gateway")
	})

	t.Run("platform-website renders", func(t *testing.T) {
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

	// Service edges

	t.Run("edge labels render", func(t *testing.T) {
		html, err := doc.Find("[data-live-infra]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "HTTPS")
		assert.Contains(t, html, "OTLP")
		assert.Contains(t, html, "gRPC")
	})

	t.Run("services have status dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-service=\"platform-website\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "bg-green-600")
	})

	// Sparklines

	t.Run("metrics render on platform-website", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		sparklines := svc.Find("[data-live-metric]")
		assert.Equal(t, 2, sparklines.Length(), "Expected Goroutines + Heap")
	})

	t.Run("sparkline has SVG", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		svgs := svc.Find("svg")
		assert.Equal(t, 2, svgs.Length())
	})

	t.Run("sparkline uses accent colors", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"platform-website\"]")
		html, err := svc.Html()
		require.NoError(t, err)
		assert.Contains(t, html, "stroke-accent-gold")
		assert.Contains(t, html, "dark:stroke-next-teal")
	})

	t.Run("metrics show on daprd", func(t *testing.T) {
		svc := doc.Find("[data-live-service=\"daprd\"]")
		sparklines := svc.Find("[data-live-metric]")
		assert.Equal(t, 1, sparklines.Length(), "Expected Components metric")
	})
}
