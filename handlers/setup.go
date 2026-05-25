package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SetupApp creates a Fiber app with production-equivalent wiring:
// error handler, routes, and all handlers. Tests and main.go both
// call this function, ensuring the handler chain is always identical.
func SetupApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "platform-website",
		ErrorHandler: ErrorHandler,
	})

	app.Get("/", Home)
	app.Get("/sections/:name", Section)
	app.Get("/docs", DocsIndex)
	app.Get("/docs/:repo", DocsRepoIndex)
	app.Get("/docs/:repo/*", DocsPage)
	app.Get("/api/version", APIVersion)
	app.Get("/api/live/stream", LiveSSE)

	return app
}
