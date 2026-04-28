package shell

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/internal/platform"
	viewapps "github.com/rezuscloud/platform-website/views/apps"
	pages "github.com/rezuscloud/platform-website/views/pages"
)

func Register(router fiber.Router, runtime platform.Runtime) {
	router.Get("/", home(runtime))
	router.Get("/shell/summary", summary(runtime))
	router.Get("/dapr/subscribe", subscribe(runtime))
	router.Post("/events/shell", observe(runtime))
}

func home(runtime platform.Runtime) fiber.Handler {
	return func(c *fiber.Ctx) error {
		state, err := sessionState(c, runtime)
		if err != nil {
			return err
		}

		return render(c, pages.Home(state))
	}
}

func summary(runtime platform.Runtime) fiber.Handler {
	return func(c *fiber.Ctx) error {
		state, err := sessionState(c, runtime)
		if err != nil {
			return err
		}

		return render(c, viewapps.ShellSummary(state, true))
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

func subscribe(runtime platform.Runtime) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON([]platform.DaprSubscription{{
			PubsubName: runtime.PubsubName(),
			Topic:      platform.SessionEventsTopic,
			Route:      "/events/shell",
		}})
	}
}

func observe(runtime platform.Runtime) fiber.Handler {
	type envelope struct {
		Data platform.SessionEvent `json:"data"`
	}

	return func(c *fiber.Ctx) error {
		var event platform.SessionEvent
		var wrapped envelope
		if err := c.BodyParser(&wrapped); err == nil && wrapped.Data.Type != "" {
			event = wrapped.Data
		} else if err := c.BodyParser(&event); err != nil {
			return err
		}

		if event.SessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "DROP"})
		}

		if _, err := runtime.ObserveEvent(c.Context(), event); err != nil {
			return err
		}

		return c.JSON(fiber.Map{"status": "SUCCESS"})
	}
}
