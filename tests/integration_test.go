package tests

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/handlers"
)

func setupIntegrationApp() *fiber.App {
	app := fiber.New(fiber.Config{})
	app.Get("/", handlers.Home)
	app.Get("/sections/:name", handlers.Section)
	return app
}

func getHTMLDoc(t *testing.T, app *fiber.App, path string) *goquery.Document {
	req := httptest.NewRequest("GET", path, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	require.NoError(t, err)

	return doc
}

func TestHomePageHTMLStructure(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("has valid HTML5 doctype", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		assert.True(t, strings.Contains(bodyStr, "<!DOCTYPE html>") || strings.Contains(bodyStr, "<html"),
			"Response should contain HTML document structure")
	})

	t.Run("has title containing RezusCloud", func(t *testing.T) {
		title := doc.Find("title").Text()
		assert.Contains(t, title, "RezusCloud")
	})

	t.Run("has meta description", func(t *testing.T) {
		description, exists := doc.Find("meta[name='description']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, description, "Enterprise Kubernetes")
	})

	t.Run("has viewport meta tag", func(t *testing.T) {
		viewport, exists := doc.Find("meta[name='viewport']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, viewport, "width=device-width")
	})
}

func TestHomePageSections(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	sections := []string{
		"hero", "challenge", "architecture", "features",
		"networking", "edge", "services", "comparison",
		"usecases", "techstack", "getstarted",
	}

	for _, section := range sections {
		t.Run("section_"+section+"_exists", func(t *testing.T) {
			selection := doc.Find("#" + section)
			assert.Equal(t, 1, selection.Length(), "Expected exactly one element with id '%s'", section)
		})

		t.Run("section_"+section+"_has_content", func(t *testing.T) {
			selection := doc.Find("#" + section)
			text := selection.Text()
			assert.Greater(t, len(strings.TrimSpace(text)), 10,
				"Section '%s' should have meaningful content", section)
		})
	}
}

func TestNavigationHTML(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("navigation exists", func(t *testing.T) {
		nav := doc.Find("nav")
		assert.Equal(t, 1, nav.Length())
	})

	navLinks := []struct {
		href string
		text string
	}{
		{"#hero", "Home"},
		{"#architecture", "Architecture"},
		{"#features", "Features"},
		{"#comparison", "Compare"},
		{"#getstarted", "Get Started"},
	}

	for _, link := range navLinks {
		t.Run("nav_link_"+link.href, func(t *testing.T) {
			found := false
			doc.Find("nav a").Each(func(i int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				if href == link.href {
					found = true
					assert.Contains(t, s.Text(), link.text)
				}
			})
			assert.True(t, found, "Expected navigation link with href '%s'", link.href)
		})
	}
}

func TestFooterHTML(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("footer exists", func(t *testing.T) {
		footer := doc.Find("footer")
		assert.Equal(t, 1, footer.Length())
	})

	t.Run("footer has links", func(t *testing.T) {
		links := doc.Find("footer a")
		assert.Greater(t, links.Length(), 0, "Footer should contain links")
	})

	t.Run("footer links are internal anchors", func(t *testing.T) {
		doc.Find("footer a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			assert.True(t, exists)
			assert.True(t, strings.HasPrefix(href, "#"),
				"Footer link href should start with '#', got: %s", href)
		})
	})
}

func TestAccessibilityHTML(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("has main landmark", func(t *testing.T) {
		main := doc.Find("main")
		assert.Equal(t, 1, main.Length())
	})

	t.Run("theme toggle has aria-label", func(t *testing.T) {
		toggle := doc.Find("button[aria-label='Toggle theme']")
		assert.Equal(t, 1, toggle.Length())
	})

	t.Run("images have alt attributes", func(t *testing.T) {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			_, exists := s.Attr("alt")
			assert.True(t, exists, "Image should have alt attribute")
		})
	})

	t.Run("links have discernible text", func(t *testing.T) {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			ariaLabel, hasAriaLabel := s.Attr("aria-label")
			assert.True(t, len(text) > 0 || hasAriaLabel,
				"Link should have text or aria-label, got text: '%s', aria-label: '%s'",
				text, ariaLabel)
		})
	})
}

func TestHTMXScriptLoaded(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("htmx script tag exists", func(t *testing.T) {
		script := doc.Find("script[src*='htmx']")
		assert.Equal(t, 1, script.Length(), "HTMX script should be included")
	})

	t.Run("htmx script has correct path", func(t *testing.T) {
		src, exists := doc.Find("script[src*='htmx']").Attr("src")
		assert.True(t, exists)
		assert.Contains(t, src, "htmx.min.js")
	})

	t.Run("htmx script in head", func(t *testing.T) {
		script := doc.Find("head script[src*='htmx']")
		assert.Equal(t, 1, script.Length(), "HTMX script should be in <head>")
	})
}

func TestHTMXAttributes(t *testing.T) {
	app := setupIntegrationApp()

	htmxAttrs := []string{
		"hx-get", "hx-post", "hx-put", "hx-delete", "hx-patch",
		"hx-trigger", "hx-swap", "hx-target", "hx-swap-oob",
		"hx-boost", "hx-include", "hx-params", "hx-headers",
		"hx-indicator", "hx-push-url", "hx-confirm", "hx-disabled-elt",
		"hx-ext", "hx-history", "hx-history-elt", "hx-on",
		"hx-preserve", "hx-prompt", "hx-replace-url", "hx-request",
		"hx-select", "hx-select-oob", "hx-sync", "hx-validate",
		"hx-vals", "hx-ws", "hx-sse",
	}

	t.Run("find any htmx attributes on home page", func(t *testing.T) {
		doc := getHTMLDoc(t, app, "/")

		foundAttrs := make(map[string][]string)
		for _, attr := range htmxAttrs {
			doc.Find("[" + attr + "]").Each(func(i int, s *goquery.Selection) {
				val, _ := s.Attr(attr)
				tag := goquery.NodeName(s)
				id, _ := s.Attr("id")
				selector := tag
				if id != "" {
					selector += "#" + id
				}
				foundAttrs[attr] = append(foundAttrs[attr], selector+"="+val)
			})
		}

		if len(foundAttrs) > 0 {
			t.Logf("Found HTMX attributes: %v", foundAttrs)
		}
	})

	t.Run("section endpoints return valid HTML", func(t *testing.T) {
		sections := []string{"hero", "features", "architecture", "getstarted"}

		for _, section := range sections {
			doc := getHTMLDoc(t, app, "/sections/"+section)

			sectionEl := doc.Find("#" + section)
			assert.Equal(t, 1, sectionEl.Length(),
				"Section endpoint should return element with id '%s'", section)
		}
	})

	t.Run("sections contain expected structure", func(t *testing.T) {
		doc := getHTMLDoc(t, app, "/sections/hero")

		headings := doc.Find("h1, h2, h3")
		assert.Greater(t, headings.Length(), 0,
			"Hero section should contain headings")
	})

	t.Run("section endpoints could be used with hx-get", func(t *testing.T) {
		sections := []string{"hero", "features", "architecture", "networking", "edge", "services", "comparison", "usecases", "techstack", "getstarted"}

		for _, section := range sections {
			doc := getHTMLDoc(t, app, "/sections/"+section)

			assert.Equal(t, 1, doc.Find("#"+section).Length(),
				"Section %s endpoint should return single element for hx-swap", section)

			content := doc.Find("#" + section).Text()
			assert.Greater(t, len(strings.TrimSpace(content)), 10,
				"Section %s should have content suitable for HTMX swap", section)
		}
	})
}

func TestHTMXDataAttributes(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	dataHtmxAttrs := []string{
		"data-hx-get", "data-hx-post", "data-hx-put", "data-hx-delete",
		"data-hx-trigger", "data-hx-swap", "data-hx-target",
	}

	t.Run("find data-hx attributes", func(t *testing.T) {
		for _, attr := range dataHtmxAttrs {
			count := doc.Find("[" + attr + "]").Length()
			if count > 0 {
				t.Logf("Found %d elements with %s", count, attr)
			}
		}
	})
}

func TestHTMXExtensionSupport(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("htmx-indicator class available", func(t *testing.T) {
		indicator := doc.Find(".htmx-indicator")
		t.Logf("Found %d htmx-indicator elements", indicator.Length())
	})

	t.Run("htmx-request class check", func(t *testing.T) {
		request := doc.Find(".htmx-request")
		t.Logf("Found %d htmx-request elements", request.Length())
	})
}

func TestResponsiveClasses(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("has responsive navigation classes", func(t *testing.T) {
		assert.Contains(t, html, "md:flex")
		assert.Contains(t, html, "hidden")
	})

	t.Run("has mobile menu button", func(t *testing.T) {
		assert.Contains(t, html, "mobile-menu-btn")
	})
}

func TestDarkModeSupport(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("has theme toggle script", func(t *testing.T) {
		assert.Contains(t, html, "theme")
	})

	t.Run("supports dark class", func(t *testing.T) {
		assert.Contains(t, html, "dark:")
	})
}
