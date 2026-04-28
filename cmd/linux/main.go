package main

import (
	"github.com/rezuscloud/platform-website/internal/platform"
	"github.com/rezuscloud/platform-website/internal/server"
)

func main() {
	server.Listen(server.NewLinuxApp(platform.NewRuntimeFromEnv()))
}
