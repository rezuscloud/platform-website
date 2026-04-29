package terminal

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/internal/platform"
	viewapps "github.com/rezuscloud/platform-website/views/apps"
	pages "github.com/rezuscloud/platform-website/views/pages"
)

func Register(router fiber.Router, runtime platform.Runtime, basePath string) {
	router.Get(basePath, page(runtime, basePath))
	router.Get(basePath+"/", page(runtime, basePath))
	router.Get(basePath+"/embed", embed(runtime, basePath))
	router.Post(basePath+"/actions/run", run(runtime, basePath))
}

func page(runtime platform.Runtime, basePath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		state, err := sessionState(c, runtime)
		if err != nil {
			return err
		}

		return render(c, viewapps.TerminalPage(state, basePath))
	}
}

func embed(runtime platform.Runtime, basePath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		state, err := sessionState(c, runtime)
		if err != nil {
			return err
		}

		return render(c, viewapps.TerminalEmbed(state, true, basePath, basePath))
	}
}

func run(runtime platform.Runtime, basePath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := platform.EnsureSessionID(c)
		command := strings.TrimSpace(c.FormValue("command"))
		parentRoute := resolvedParentRoute(c.FormValue("parent"), basePath)
		if command == "" {
			command = strings.TrimSpace(c.FormValue("preset"))
		}

		response, err := runtime.RunCommand(c.Context(), sessionID, command)
		if err != nil {
			return err
		}

		c.Set("HX-Trigger", "session-updated")
		if c.Get("HX-Request") == "true" {
			return render(c, viewapps.TerminalEmbed(response.State, true, basePath, parentRoute))
		}

		switch parentRoute {
		case "/":
			return render(c, pages.Home(response.State))
		case "/apps/mac":
			return render(c, viewapps.MacPage(response.State, "/apps/mac"))
		case "/apps/linux":
			return render(c, viewapps.LinuxPage(response.State, "/apps/linux"))
		default:
			return render(c, viewapps.TerminalPage(response.State, basePath))
		}
	}
}

func resolvedParentRoute(parent string, terminalBasePath string) string {
	switch strings.TrimSpace(parent) {
	case "/", "/apps/mac", "/apps/linux":
		return strings.TrimSpace(parent)
	default:
		return terminalBasePath
	}
}

func sessionState(c *fiber.Ctx, runtime platform.Runtime) (platform.SessionState, error) {
	sessionID := platform.EnsureSessionID(c)
	return runtime.LoadSession(c.Context(), sessionID)
}

func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}
