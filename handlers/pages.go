package handlers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/views/pages"
	"github.com/rezuscloud/platform-website/views/sections"
)

// Render adapts a templ.Component to a Fiber response.
func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// Home renders the full landing page.
func Home(c *fiber.Ctx) error {
	return render(c, pages.Home())
}

// Section renders an individual section for HTMX partial swaps.
func Section(c *fiber.Ctx) error {
	name := c.Params("name")

	sectionMap := map[string]templ.Component{
		"hero":         sections.Hero(),
		"challenge":    sections.Challenge(),
		"architecture": sections.Architecture(),
		"features":     sections.Features(),
		"networking":   sections.Networking(),
		"comparison":   sections.Comparison(),
		"usecases":     sections.UseCases(),
		"getstarted":   sections.GetStarted(),
	}

	component, ok := sectionMap[name]
	if !ok {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return render(c, component)
}
