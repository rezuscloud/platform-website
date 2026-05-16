// Package middleware provides Fiber middleware for the platform website.
package middleware

import "github.com/gofiber/fiber/v2"

// CSP is the Content Security Policy for the website.
// Alpine.js v3 requires 'unsafe-eval' for its expression evaluator (new Function()).
// Without it, ALL Alpine expressions fail silently: theme toggle, nav, SSE, dark mode.
// See: https://github.com/alpinejs/alpine/issues/346
const CSP = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"font-src 'self'; " +
	"img-src 'self' data: https:; " +
	"connect-src 'self'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'"

// SecurityHeaders adds production security headers to every response.
func SecurityHeaders(c *fiber.Ctx) error {
	err := c.Next()

	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	c.Set("Content-Security-Policy", CSP)

	return err
}
