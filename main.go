package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
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

	if os.Getenv("PPROF_ENABLED") == "true" {
		app.Use(pprof.New())
		go func() {
			log.Println("Starting pprof server on :6060")
			log.Fatal(http.ListenAndServe(":6060", nil))
		}()
	}

	app.Static("/assets", "./assets", fiber.Static{
		CacheDuration: -1,
	})

	app.Get("/manifest.webmanifest", func(c *fiber.Ctx) error {
		return c.SendFile("./assets/manifest.webmanifest")
	})

	app.Get("/", handlers.Home)
	app.Get("/sections/:name", handlers.Section)
	app.Get("/api/version", handlers.APIVersion)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	log.Fatal(app.Listener(ln))
}
