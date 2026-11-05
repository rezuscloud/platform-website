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
		assert.Contains(t, d, "liveMatrix")
	})

	t.Run("has EventSource", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "EventSource")
		assert.Contains(t, html, "/api/live/stream")
	})

	// Blueprint-tier structure tests
	t.Run("has node cards with host labels", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "OCI Cloud")
		assert.Contains(t, html, "Edge")
		assert.Contains(t, html, "data-node-card")
	})

	t.Run("has tier bars for the blueprint presentation", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "TIER 1 · SUBSTRATE")
		assert.Contains(t, html, "TIER 2 · GUARANTEES")
		assert.Contains(t, html, "TIER 3 · PLANES")
		assert.Contains(t, html, "HOSTED WORKLOADS")
	})

	t.Run("categories named by blueprint capability class", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		// substrate capability classes (contract rows)
		assert.Contains(t, html, "A · Compute &amp; orchestration")
		assert.Contains(t, html, "B · Network")
		assert.Contains(t, html, "D · Block storage")
		// guarantees rows
		assert.Contains(t, html, "G · Identity &amp; access")
		assert.Contains(t, html, "J · Observability &amp; measurement")
		// hosted workload purposes
		assert.Contains(t, html, "Code &amp; CI")
		assert.Contains(t, html, "Web")
		// legacy flat categories must be gone
		assert.NotContains(t, html, ">Development<")
		assert.NotContains(t, html, ">Runtime<")
	})

	t.Run("service rows have clickable cells", func(t *testing.T) {
		cells := doc.Find("[data-svc-key]")
		assert.GreaterOrEqual(t, cells.Length(), 8, "should have at least 8 clickable service cells")
	})

	t.Run("cells have status dots", func(t *testing.T) {
		dots := doc.Find("[data-svc-key] [data-dot]")
		cells := doc.Find("[data-svc-key]")
		assert.Equal(t, cells.Length(), dots.Length(), "each cell should have a status dot")
	})

	t.Run("click-to-expand has selectService", func(t *testing.T) {
		html, err := doc.Find("#live").Html()
		require.NoError(t, err)
		assert.Contains(t, html, "selectService")
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
