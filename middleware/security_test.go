package middleware

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeadersPresent(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.Equal(t, "max-age=63072000; includeSubDomains; preload", resp.Header.Get("Strict-Transport-Security"))
	assert.Equal(t, CSP, resp.Header.Get("Content-Security-Policy"))
}

func TestCSPAllowsAlpineJS(t *testing.T) {
	// Alpine.js v3 uses new Function() for expression evaluation.
	// Without 'unsafe-eval' in script-src, ALL Alpine expressions fail silently:
	// theme toggle, nav active state, SSE connection, dark mode, x-show, x-bind.
	//
	// This test ensures the CSP constant never accidentally removes 'unsafe-eval'.
	// Regression test for the CSP breakage that killed all interactivity.

	t.Run("unsafe-eval for Alpine.js expression evaluator", func(t *testing.T) {
		assert.Contains(t, CSP, "'unsafe-eval'",
			"CSP script-src must include 'unsafe-eval' for Alpine.js v3 (new Function)")
	})

	t.Run("unsafe-inline for inline scripts", func(t *testing.T) {
		assert.Contains(t, CSP, "'unsafe-inline'",
			"CSP script-src must include 'unsafe-inline' for inline scripts")
	})

	t.Run("connect-src self for SSE", func(t *testing.T) {
		assert.Contains(t, CSP, "connect-src 'self'",
			"CSP must allow connect-src 'self' for SSE EventSource")
	})

	t.Run("script-src has exactly self + two unsafes", func(t *testing.T) {
		scriptSrcRegex := regexp.MustCompile(`script-src\s+([^;]+)`)
		matches := scriptSrcRegex.FindStringSubmatch(CSP)
		require.Len(t, matches, 2, "CSP must have a script-src directive")
		directive := matches[1]
		assert.Contains(t, directive, "'self'")
		assert.Contains(t, directive, "'unsafe-inline'")
		assert.Contains(t, directive, "'unsafe-eval'")
	})
}
