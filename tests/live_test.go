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
}
