package handlers

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupApp() *fiber.App {
	app := fiber.New(fiber.Config{})
	app.Get("/", Home)
	app.Get("/sections/:name", Section)
	return app
}

func TestHomeHandler(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestHomeHandlerContainsExpectedContent(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.Contains(t, html, "RezusCloud")
	assert.Contains(t, html, ">YOUR</span>")
	assert.Contains(t, html, ">PERSONAL</span>")
	assert.Contains(t, html, ">CLOUD</span>")
	assert.Contains(t, html, "personal computer changed everything")
}

func TestSectionHandler(t *testing.T) {
	sections := []string{
		"hero", "challenge", "architecture", "features",
		"networking", "edge", "services", "comparison",
		"usecases", "techstack", "getstarted",
	}

	app := setupApp()

	for _, section := range sections {
		t.Run(section, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/sections/"+section, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
		})
	}
}

func TestSectionHandlerContainsContent(t *testing.T) {
	app := setupApp()

	expectedContent := map[string]string{
		"hero":         "personal computer changed everything",
		"features":     "What You Get",
		"architecture": "How It Works",
		"getstarted":   "Start Your Cloud",
	}

	for section, expected := range expectedContent {
		t.Run(section+"_content", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/sections/"+section, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), expected)
		})
	}
}

func TestSectionHandlerNotFound(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/sections/nonexistent", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestSectionHandlerReturnsSectionID(t *testing.T) {
	app := setupApp()

	sections := []string{"hero", "features", "architecture"}

	for _, section := range sections {
		t.Run(section+"_has_id", func(t *testing.T) {
			req := httptest.NewRequest("GET", "/sections/"+section, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), `id="`+section+`"`)
		})
	}
}

func TestHomePageContainsAllSections(t *testing.T) {
	app := setupApp()

	sections := []string{
		"hero", "challenge", "architecture", "features",
		"networking", "edge", "services", "comparison",
		"usecases", "techstack", "getstarted",
	}

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	for _, section := range sections {
		assert.True(t, strings.Contains(html, `id="`+section+`"`),
			"Expected section %s to be present in home page", section)
	}
}

func TestHomePageNoOldDesignTokens(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// Verify old color tokens are completely removed
	oldTokens := []string{
		"cream-50", "cream-100", "cream-200", "cream-300",
		"cream-400", "cream-500", "cream-600", "cream-700", "cream-800", "cream-900",
		"phosphor-", "retro-blue-",
		"terminal-bg", "terminal-surface", "terminal-border",
	}
	for _, token := range oldTokens {
		assert.NotContains(t, html, token,
			"Old design token '%s' should be removed from HTML", token)
	}
}

func TestHomePageNewDesignTokens(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// Verify Mac mode tokens are present
	macTokens := []string{"bg-paper", "bg-surface", "text-ink", "text-ink-muted", "border-rule", "text-accent-gold"}
	for _, token := range macTokens {
		assert.Contains(t, html, token,
			"Expected Mac mode token '%s' to be present", token)
	}

	// Verify NeXT mode tokens are present
	nextTokens := []string{"dark:bg-next-black", "dark:bg-next-dark", "dark:text-next-white", "dark:text-next-subtle"}
	for _, token := range nextTokens {
		assert.Contains(t, html, token,
			"Expected NeXT mode token '%s' to be present", token)
	}
}

func TestHomePageNoRoundedCorners(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// Verify no rounded corner classes
	roundedClasses := []string{"rounded-xl", "rounded-2xl", "rounded-lg", "rounded-md", "rounded-full"}
	for _, cls := range roundedClasses {
		assert.NotContains(t, html, cls,
			"Rounded corner class '%s' should not be present", cls)
	}
}

func TestHomePageNoBorder2(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// border-2 is prohibited; only border (1px) is allowed
	assert.NotContains(t, html, "border-2",
		"border-2 should not be present; only 1px borders allowed")
}

func TestHomePageFontFamilies(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// Verify font hierarchy: Silkscreen headings + system-ui body (same in both modes)
	assert.Contains(t, html, "font-mac", "Display font (Silkscreen) should be present")
	assert.Contains(t, html, "font-mac-body", "Body font (system-ui) should be present")

	// Verify old font class is gone
	assert.NotContains(t, html, "font-retro", "Old font-retro class should be removed")
	assert.NotContains(t, html, "IBM Plex Mono", "IBM Plex Mono should be removed")
}

func TestHomePageNoCRTEffects(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// CRT effects should be removed
	crtClasses := []string{"crt-scanlines", "crt-glow", "retro-bevel"}
	for _, cls := range crtClasses {
		assert.NotContains(t, html, cls,
			"CRT/retro effect '%s' should be removed", cls)
	}
}

func TestHomePageNeXTBevels(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)

	// NeXT bevel utilities should be present in dark mode
	assert.Contains(t, html, "dark:next-raised",
		"NeXT raised bevel should be used in dark mode")
	assert.Contains(t, html, "dark:next-sunken",
		"NeXT sunken bevel should be used in dark mode")
}

func TestHomePageContainsNavigation(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.Contains(t, html, "<nav")
	assert.Contains(t, html, "</nav>")
}

func TestHomePageContainsFooter(t *testing.T) {
	app := setupApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.Contains(t, html, "<footer")
	assert.Contains(t, html, "</footer>")
}
