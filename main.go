package main

import (
	"os"
	"strings"

	"github.com/rezuscloud/platform-website/internal/platform"
	"github.com/rezuscloud/platform-website/internal/server"
)

func main() {
	mode := strings.TrimSpace(os.Getenv("PLATFORM_MODE"))

	switch mode {
	case "shell":
		server.Listen(server.NewShellApp(platform.NewRuntimeFromEnv()))
	case "terminal":
		server.Listen(server.NewTerminalApp(platform.NewRuntimeFromEnv()))
	case "mac":
		server.Listen(server.NewMacApp(platform.NewRuntimeFromEnv()))
	case "linux":
		server.Listen(server.NewLinuxApp(platform.NewRuntimeFromEnv()))
	default:
		server.Listen(server.NewGatewayApp(platform.NewLocalRuntime()))
	}
}
