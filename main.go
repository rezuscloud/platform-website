package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/rezuscloud/platform-website/handlers"
	"github.com/rezuscloud/platform-website/obs"
)

// securityHeaders adds production security headers to every response.
func securityHeaders(c *fiber.Ctx) error {
	err := c.Next()

	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

	// CSP allows self + inline scripts/styles (needed for Alpine.js + Tailwind).
	// connect-src self is needed for SSE /api/live/stream.
	c.Set("Content-Security-Policy",
		"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"font-src 'self'; "+
			"img-src 'self' data: https:; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self'")

	return err
}

func main() {
	meterProvider := obs.InitTelemetry()

	app := fiber.New(fiber.Config{
		AppName:      "platform-website",
		ServerHeader: "",
		ErrorHandler: handlers.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(obs.OTelFiberMiddleware(meterProvider))
	app.Use(securityHeaders)
	app.Use(logger.New())
	app.Use(compress.New())

	if os.Getenv("PPROF_ENABLED") == "true" {
		app.Use(pprof.New())
	}

	app.Static("/assets", "./assets", fiber.Static{
		MaxAge: 86400, // 24h for CSS/JS/SVG
	})

	// Long-lived cache for fonts (content-addressed, rarely change)
	app.Static("/assets/fonts", "./assets/fonts", fiber.Static{
		MaxAge: 31536000, // 1 year
	})

	app.Get("/manifest.webmanifest", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/manifest+json")
		return c.SendFile("./assets/manifest.webmanifest")
	})

	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; charset=utf-8")
		return c.SendString("User-agent: *\nAllow: /\n\nSitemap: https://rezus.cloud/sitemap.xml\n")
	})

	app.Get("/sitemap.xml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/xml")
		return c.SendString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://rezus.cloud/</loc><changefreq>weekly</changefreq><priority>1.0</priority></url>
</urlset>`)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Get("/", handlers.Home)
	app.Get("/sections/:name", handlers.Section)
	app.Get("/api/version", handlers.APIVersion)
	app.Get("/api/live/stream", handlers.LiveSSE)

	// Wire SigNoz client — try Dapr secrets first, fall back to env vars
	obs.LoadSecretsFromDapr()
	if signoz := obs.NewSigNozClientFromEnv(); signoz != nil {
		handlers.SetLiveClient(signoz)
		log.Printf("Live section using SigNoz metrics from %s", os.Getenv("SIGNOZ_URL"))
	} else {
		log.Println("SIGNOZ_URL/SIGNOZ_API_KEY not set, live section using mock data")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":3000"
	log.Printf("Starting server on %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		obs.ShutdownTelemetry(context.Background())
		log.Fatalf("Failed to create listener: %v", err)
	}
	log.Fatal(app.Listener(ln))
}
