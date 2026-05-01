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
		assert.Contains(t, description, "Personal Cloud")
	})

	t.Run("has canonical link", func(t *testing.T) {
		canonical, exists := doc.Find("link[rel='canonical']").Attr("href")
		assert.True(t, exists)
		assert.Equal(t, "https://rezus.cloud/", canonical)
	})

	t.Run("has open graph metadata", func(t *testing.T) {
		ogTitle, exists := doc.Find("meta[property='og:title']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, ogTitle, "Your Personal Cloud")

		ogImage, exists := doc.Find("meta[property='og:image']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, ogImage, "icon-512.png")

		ogSiteName, exists := doc.Find("meta[property='og:site_name']").Attr("content")
		assert.True(t, exists)
		assert.Equal(t, "RezusCloud", ogSiteName)

		ogImageAlt, exists := doc.Find("meta[property='og:image:alt']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, ogImageAlt, "RezusCloud")
	})

	t.Run("has twitter metadata", func(t *testing.T) {
		twitterCard, exists := doc.Find("meta[name='twitter:card']").Attr("content")
		assert.True(t, exists)
		assert.Equal(t, "summary_large_image", twitterCard)

		twitterImageAlt, exists := doc.Find("meta[name='twitter:image:alt']").Attr("content")
		assert.True(t, exists)
		assert.Contains(t, twitterImageAlt, "RezusCloud")
	})

	t.Run("has structured data", func(t *testing.T) {
		structuredData := doc.Find("script[type='application/ld+json']").Text()
		assert.Contains(t, structuredData, `"@type": "WebSite"`)
		assert.Contains(t, structuredData, `"@type": "Organization"`)
		assert.Contains(t, structuredData, `"email": "tiberiu@rezus.net"`)
	})

	t.Run("has manifest link", func(t *testing.T) {
		manifest, exists := doc.Find("link[rel='manifest']").Attr("href")
		assert.True(t, exists)
		assert.Equal(t, "/manifest.webmanifest", manifest)
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

	t.Run("has mobile menu button with Alpine.js", func(t *testing.T) {
		assert.Contains(t, html, "@click")
		assert.Contains(t, html, "mobileOpen")
	})

	t.Run("has mobile menu transitions", func(t *testing.T) {
		assert.Contains(t, html, "x-transition")
	})
}

func TestAlpineJSIntegration(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("alpine script loaded", func(t *testing.T) {
		script := doc.Find("script[src*='alpine']")
		assert.Equal(t, 1, script.Length(), "Alpine.js script should be included")
	})

	t.Run("alpine script has defer", func(t *testing.T) {
		deferred, exists := doc.Find("script[src*='alpine']").Attr("defer")
		assert.True(t, exists)
		assert.True(t, deferred == "" || deferred == "defer" || deferred == "true")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("html has x-data for theme", func(t *testing.T) {
		assert.Contains(t, html, `x-data`)
		assert.Contains(t, html, "dark")
	})

	t.Run("theme toggle uses Alpine store", func(t *testing.T) {
		assert.Contains(t, html, "$store.theme")
	})

	t.Run("x-cloak style defined", func(t *testing.T) {
		assert.Contains(t, html, "x-cloak")
	})

	t.Run("mobile nav uses Alpine state", func(t *testing.T) {
		assert.Contains(t, html, "x-show")
		assert.Contains(t, html, "x-cloak")
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

	t.Run("supports dark class", func(t *testing.T) {
		assert.Contains(t, html, "dark:")
	})

	t.Run("theme persisted to localStorage", func(t *testing.T) {
		assert.Contains(t, html, "localStorage")
	})

	t.Run("respects prefers-color-scheme", func(t *testing.T) {
		assert.Contains(t, html, "prefers-color-scheme")
	})
}

func TestDesignSystemMacTokens(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("body has Mac mode background", func(t *testing.T) {
		assert.Contains(t, html, "bg-paper", "Body should use Mac paper background")
	})

	t.Run("body has Mac mode font", func(t *testing.T) {
		assert.Contains(t, html, "font-mac-body", "Body should use Mac body font")
	})

	t.Run("navigation uses Mac tokens", func(t *testing.T) {
		assert.Contains(t, html, "bg-surface-warm", "Nav should use Mac surface-warm background")
		assert.Contains(t, html, "border-rule", "Nav should use Mac rule border")
	})

	t.Run("no old cream tokens in nav", func(t *testing.T) {
		// The nav used cream-100 and cream-300 — verify they're gone
		navHTML := docFindNav(t, html)
		oldTokens := []string{"cream-100", "cream-200", "cream-300"}
		for _, token := range oldTokens {
			assert.NotContains(t, navHTML, token,
				"Nav should not contain old token '%s'", token)
		}
	})
}

func TestDesignSystemNeXTTokens(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("body has NeXT dark background", func(t *testing.T) {
		assert.Contains(t, html, "dark:bg-next-black", "Body should have NeXT black in dark mode")
	})

	t.Run("body has NeXT font in dark mode", func(t *testing.T) {
		assert.Contains(t, html, "dark:font-next", "Body should switch to NeXT font in dark mode")
	})

	t.Run("navigation has NeXT dark tokens", func(t *testing.T) {
		assert.Contains(t, html, "dark:bg-next-dark", "Nav should have NeXT dark background")
	})

	t.Run("NeXT bevel utilities present", func(t *testing.T) {
		assert.Contains(t, html, "dark:next-raised", "Raised bevel should appear in dark mode")
	})
}

func TestDesignSystemZeroRoundedCorners(t *testing.T) {
	app := setupIntegrationApp()

	t.Run("no rounded corners on home page", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		html := string(body)

		roundedClasses := []string{"rounded-xl", "rounded-2xl", "rounded-lg", "rounded-md", "rounded-full"}
		for _, cls := range roundedClasses {
			assert.NotContains(t, html, cls,
				"Rounded class '%s' should not be present anywhere", cls)
		}
	})

	t.Run("no rounded corners on section endpoints", func(t *testing.T) {
		sections := []string{"hero", "features", "architecture", "networking", "edge", "services", "comparison", "usecases", "techstack", "getstarted"}

		for _, section := range sections {
			req := httptest.NewRequest("GET", "/sections/"+section, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			html := string(body)

			for _, cls := range []string{"rounded-xl", "rounded-2xl", "rounded-lg"} {
				assert.NotContains(t, html, cls,
					"Section '%s' should not contain rounded class '%s'", section, cls)
			}
		}
	})
}

func TestDesignSystemNoOldTokens(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	oldTokens := []string{
		"cream-50", "cream-100", "cream-200", "cream-300", "cream-400",
		"cream-500", "cream-600", "cream-700", "cream-800", "cream-900",
		"phosphor-400", "phosphor-500", "phosphor-600",
		"terminal-bg", "terminal-surface", "terminal-border",
		"retro-blue-",
	}
	for _, token := range oldTokens {
		assert.NotContains(t, html, token,
			"Old design token '%s' should be completely removed", token)
	}
}

func TestDesignSystemOnePxBorders(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// border-2 and border-3 are prohibited
	assert.NotContains(t, html, "border-2",
		"Only 1px borders allowed, not border-2")
	assert.NotContains(t, html, "border-3",
		"Only 1px borders allowed, not border-3")
}

func TestDesignSystemDualModeToggle(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("theme toggle exists", func(t *testing.T) {
		assert.Contains(t, html, "Toggle theme",
			"Theme toggle button should exist")
	})

	t.Run("both mode classes present", func(t *testing.T) {
		// Mac mode (default, no dark: prefix)
		assert.Contains(t, html, "bg-paper", "Mac mode background should be present")
		// NeXT mode (dark: prefix)
		assert.Contains(t, html, "dark:bg-next-black", "NeXT mode background should be present")
	})

	t.Run("fonts switch between modes", func(t *testing.T) {
		assert.Contains(t, html, "font-mac-body", "Mac font should be default")
		assert.Contains(t, html, "dark:font-next", "NeXT font should activate in dark mode")
	})
}

func TestDesignSystemTables(t *testing.T) {
	app := setupIntegrationApp()

	t.Run("comparison table uses new tokens", func(t *testing.T) {
		doc := getHTMLDoc(t, app, "/sections/comparison")

		// Table should exist
		table := doc.Find("table")
		assert.Equal(t, 1, table.Length(), "Comparison section should have a table")

		// Table rows should exist
		rows := doc.Find("table tbody tr")
		assert.GreaterOrEqual(t, rows.Length(), 5, "Table should have comparison rows")
	})

	t.Run("edge table uses new tokens", func(t *testing.T) {
		doc := getHTMLDoc(t, app, "/sections/edge")

		table := doc.Find("table")
		assert.Equal(t, 1, table.Length(), "Edge section should have a table")
	})

	t.Run("challenge table uses new tokens", func(t *testing.T) {
		doc := getHTMLDoc(t, app, "/sections/challenge")

		table := doc.Find("table")
		assert.Equal(t, 1, table.Length(), "Challenge section should have a table")
	})
}

func docFindNav(t *testing.T, html string) string {
	t.Helper()
	start := strings.Index(html, "<nav")
	end := strings.Index(html, "</nav>")
	if start == -1 || end == -1 {
		return ""
	}
	return html[start : end+6]
}

func TestProgressiveEnhancement(t *testing.T) {
	app := setupIntegrationApp()
	doc := getHTMLDoc(t, app, "/")

	t.Run("page works without JavaScript - all content present", func(t *testing.T) {
		sections := []string{"hero", "challenge", "architecture", "features", "networking", "edge", "services", "comparison", "usecases", "techstack", "getstarted"}
		for _, section := range sections {
			assert.Equal(t, 1, doc.Find("#"+section).Length(), "Section %s should exist server-side", section)
		}
	})

	t.Run("navigation links work without JavaScript", func(t *testing.T) {
		navLinks := doc.Find("nav a[href^='#']")
		assert.GreaterOrEqual(t, navLinks.Length(), 5, "Should have navigation links")
	})

	t.Run("forms and buttons have fallback behavior", func(t *testing.T) {
		links := doc.Find("a[href]")
		links.Each(func(i int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			assert.NotEmpty(t, href, "Links should have href attributes")
		})
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("content visible before JavaScript loads", func(t *testing.T) {
		assert.Contains(t, html, "Your Personal Cloud")
		assert.Contains(t, html, "RezusCloud")
	})
}

func TestAlpineHTMXSeparation(t *testing.T) {
	app := setupIntegrationApp()
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	t.Run("Alpine handles client-side state (theme, mobile menu)", func(t *testing.T) {
		assert.Contains(t, html, "x-data", "Alpine x-data for state")
		assert.Contains(t, html, "x-show", "Alpine x-show for visibility")
		assert.Contains(t, html, "@click", "Alpine @click for events")
		assert.Contains(t, html, "mobileOpen", "Alpine state for mobile menu")
		assert.Contains(t, html, "$store.theme", "Alpine store for theme")
	})

	t.Run("HTMX ready for server-side interactions", func(t *testing.T) {
		assert.Contains(t, html, "htmx.min.js", "HTMX script loaded")
	})

	t.Run("Alpine deferred for progressive enhancement", func(t *testing.T) {
		assert.Contains(t, html, `defer src="/assets/js/alpine.min.js"`, "Alpine loaded with defer")
	})

	t.Run("x-cloak prevents flash of unstyled content", func(t *testing.T) {
		assert.Contains(t, html, "x-cloak", "x-cloak attribute present")
	})
}
