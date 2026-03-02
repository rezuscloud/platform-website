package main

import (
	"log"
	"net"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/rezuscloud/platform-website/handlers"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "platform-website",
		ServerHeader: "Fiber",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(compress.New())

	app.Static("/assets", "./assets", fiber.Static{
		CacheDuration: -1,
	})

	app.Get("/manifest.webmanifest", func(c *fiber.Ctx) error {
		return c.SendFile("./assets/manifest.webmanifest")
	})

	app.Get("/", handlers.Home)
	app.Get("/sections/:name", handlers.Section)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":3000"
	log.Printf("Starting server on %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	log.Fatal(app.Listener(ln))
}
