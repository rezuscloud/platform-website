package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/internal/platform"
	"github.com/rezuscloud/platform-website/internal/server"
	"github.com/rezuscloud/platform-website/internal/telemetry"
)

func main() {
	mode := strings.TrimSpace(os.Getenv("PLATFORM_MODE"))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appName, app := buildApp(mode)

	shutdown, err := telemetry.Init(ctx, appName)
	if err != nil {
		log.Printf("telemetry init: %v", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Listen(app)
	}()

	select {
	case <-ctx.Done():
		log.Println("Received shutdown signal")
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Server shutdown: %v", err)
	}

	if err := shutdown(shutdownCtx); err != nil {
		log.Printf("Telemetry shutdown: %v", err)
	}
}

func buildApp(mode string) (string, *fiber.App) {
	rt := platform.NewRuntimeFromEnv()

	switch mode {
	case "shell":
		return platform.ShellAppID, server.NewShellApp(rt)
	case "terminal":
		return platform.TerminalAppID, server.NewTerminalApp(rt)
	case "mac":
		return platform.MacAppID, server.NewMacApp(rt)
	case "linux":
		return platform.LinuxAppID, server.NewLinuxApp(rt)
	default:
		return "platform-website", server.NewGatewayApp(platform.NewLocalRuntime())
	}
}
