package server

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
	linuxapp "github.com/rezuscloud/platform-website/internal/apps/linux"
	macapp "github.com/rezuscloud/platform-website/internal/apps/mac"
	shellapp "github.com/rezuscloud/platform-website/internal/apps/shell"
	terminalapp "github.com/rezuscloud/platform-website/internal/apps/terminal"
	"github.com/rezuscloud/platform-website/internal/platform"
)

func NewGatewayApp(runtime platform.Runtime) *fiber.App {
	app := newBaseApp("platform-website")
	shellapp.Register(app, runtime)
	terminalapp.Register(app, runtime, "/apps/terminal")
	macapp.Register(app, runtime, "/apps/mac")
	linuxapp.Register(app, runtime, "/apps/linux")

	return app
}

func NewShellApp(runtime platform.Runtime) *fiber.App {
	app := newBaseApp(platform.ShellAppID)
	shellapp.Register(app, runtime)
	return app
}

func NewTerminalApp(runtime platform.Runtime) *fiber.App {
	app := newBaseApp(platform.TerminalAppID)
	terminalapp.Register(app, runtime, "/apps/terminal")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/apps/terminal", fiber.StatusTemporaryRedirect)
	})
	return app
}

func NewMacApp(runtime platform.Runtime) *fiber.App {
	app := newBaseApp(platform.MacAppID)
	macapp.Register(app, runtime, "/apps/mac")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/apps/mac", fiber.StatusTemporaryRedirect)
	})
	return app
}

func NewLinuxApp(runtime platform.Runtime) *fiber.App {
	app := newBaseApp(platform.LinuxAppID)
	linuxapp.Register(app, runtime, "/apps/linux")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/apps/linux", fiber.StatusTemporaryRedirect)
	})
	return app
}

func Listen(app *fiber.App) {
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

func newBaseApp(appName string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      appName,
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

	app.Static("/assets", "./assets", fiber.Static{CacheDuration: -1})
	app.Get("/manifest.webmanifest", func(c *fiber.Ctx) error {
		return c.SendFile("./assets/manifest.webmanifest")
	})
	app.Get("/api/version", handlers.APIVersion)

	return app
}
