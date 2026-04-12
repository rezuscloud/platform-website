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
	assert.Contains(t, html, "Your Personal")
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
		"hero":         "Your Personal",
		"features":     "Key Differentiators",
		"architecture": "Architecture",
		"getstarted":   "Get Started",
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
