package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/rezuscloud/platform-website/handlers"
	"github.com/rezuscloud/platform-website/middleware"
	"github.com/rezuscloud/platform-website/obs"
)

func main() {
	meterProvider := obs.InitTelemetry()

	app := handlers.SetupApp()

	// Middleware chain (applied before routes because SetupApp registers routes first,
	// but Fiber processes middleware in registration order per request)
	app.Use(recover.New())
	app.Use(obs.OTelFiberMiddleware(meterProvider))
	app.Use(middleware.SecurityHeaders)
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

	// Wire SigNoz client — try Dapr secrets first, fall back to env vars
	obs.LoadSecretsFromDapr()
	if signoz := obs.NewSigNozClientFromEnv(); signoz != nil {
		handlers.SetLiveClient(signoz)
		signozURL := os.Getenv("SIGNOZ_URL")
		log.Printf("Live section using SigNoz metrics from %s", signozURL)

		// Startup health check: warn early if SigNoz is unreachable
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if _, err := signoz.Fetch(ctx); err != nil {
			log.Printf("WARNING: SigNoz health check failed (%s): %v", signozURL, err)
		} else {
			log.Printf("SigNoz health check OK (%s)", signozURL)
		}
		cancel()
	} else {
		log.Println("SIGNOZ_URL/SIGNOZ_API_KEY not set, live section using mock data")
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
