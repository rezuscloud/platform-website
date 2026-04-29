package mac

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/internal/platform"
	viewapps "github.com/rezuscloud/platform-website/views/apps"
)

func Register(router fiber.Router, runtime platform.Runtime, basePath string) {
	router.Get(basePath, page(runtime, basePath))
	router.Get(basePath+"/", page(runtime, basePath))
	router.Get(basePath+"/embed", embed(runtime, basePath))
}

func page(runtime platform.Runtime, basePath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		state, err := sessionState(c, runtime)
		if err != nil {
			return err
		}

		return render(c, viewapps.MacPage(state, basePath))
	}
}

func embed(runtime platform.Runtime, basePath string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		state, err := sessionState(c, runtime)
		if err != nil {
			return err
		}

		return render(c, viewapps.MacEmbed(state, true, basePath, "/apps/terminal"))
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
