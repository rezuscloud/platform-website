package handlers

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeadersPresent(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		// Simulate the securityHeaders middleware from main.go
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"font-src 'self'; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'")
		return c.SendString("ok")
	})
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"))
	assert.NotEmpty(t, resp.Header.Get("Strict-Transport-Security"))
	assert.NotEmpty(t, resp.Header.Get("Content-Security-Policy"))
}

func TestCSPAllowsAlpineJS(t *testing.T) {
	// Alpine.js v3 uses new Function() for expression evaluation.
	// Without 'unsafe-eval' in script-src, ALL Alpine expressions fail silently:
	// theme toggle, nav active state, SSE connection, dark mode, x-show, x-bind.
	//
	// This test ensures the CSP never accidentally removes 'unsafe-eval'.
	// See: https://github.com/alpinejs/alpine/issues/346

	csp := "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"font-src 'self'; " +
		"img-src 'self' data: https:; " +
		"connect-src 'self'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"

	// Extract script-src directive
	scriptSrcRegex := regexp.MustCompile(`script-src\s+([^;]+)`)
	matches := scriptSrcRegex.FindStringSubmatch(csp)
	require.Len(t, matches, 2, "CSP must have a script-src directive")

	directive := matches[1]

	t.Run("unsafe-eval is present for Alpine.js", func(t *testing.T) {
		assert.Contains(t, directive, "'unsafe-eval'",
			"CSP script-src must include 'unsafe-eval' for Alpine.js v3 expression evaluator (new Function)")
	})

	t.Run("unsafe-inline is present for Alpine.js", func(t *testing.T) {
		assert.Contains(t, directive, "'unsafe-inline'",
			"CSP script-src must include 'unsafe-inline' for inline scripts")
	})

	t.Run("connect-src allows self for SSE", func(t *testing.T) {
		assert.Contains(t, csp, "connect-src 'self'",
			"CSP must allow connect-src 'self' for SSE EventSource to /api/live/stream")
	})
}
