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
		assert.Contains(t, d, "liveDashboard")
	})

	t.Run("has EventSource script", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "EventSource")
		assert.Contains(t, html, "/api/live/stream")
	})

	t.Run("services render with data-live-svc", func(t *testing.T) {
		services := doc.Find("[data-live-svc]")
		assert.GreaterOrEqual(t, services.Length(), 5, "should have at least 5 services")
	})

	t.Run("services have status dots with data-dot", func(t *testing.T) {
		dots := doc.Find("[data-dot]")
		assert.Equal(t, doc.Find("[data-live-svc]").Length(), dots.Length())
	})

	t.Run("services have CPU elements", func(t *testing.T) {
		cpu := doc.Find("[data-cpu]")
		assert.GreaterOrEqual(t, cpu.Length(), 1)
	})

	t.Run("services have RAM elements", func(t *testing.T) {
		ram := doc.Find("[data-ram]")
		assert.GreaterOrEqual(t, ram.Length(), 1)
	})

	t.Run("services have uptime elements", func(t *testing.T) {
		uptime := doc.Find("[data-uptime]")
		assert.GreaterOrEqual(t, uptime.Length(), 1)
	})

	t.Run("services have CPU sparklines", func(t *testing.T) {
		charts := doc.Find("[data-cpu-hist]")
		assert.GreaterOrEqual(t, charts.Length(), 1)
	})

	t.Run("services have RAM sparklines", func(t *testing.T) {
		charts := doc.Find("[data-ram-hist]")
		assert.GreaterOrEqual(t, charts.Length(), 1)
	})

	t.Run("unmonitored services show hollow dots", func(t *testing.T) {
		html, err := doc.Find("[data-live-svc*=\"signoz\"]").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "border border-rule")
	})

	t.Run("static banner shows", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "Showing platform topology")
	})

	t.Run("service names come from mock data", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "source-controller")
		assert.Contains(t, html, "platform-website")
		assert.Contains(t, html, "forgejo")
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
