package handlers

import (
	"context"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/obs"
	"github.com/rezuscloud/platform-website/views/pages"
	"github.com/rezuscloud/platform-website/views/sections"
)

// SectionHandler renders a named section. Dynamic sections capture
// their own data source at wiring time; static sections ignore ctx.
type SectionHandler func(ctx context.Context) templ.Component

// sectionRegistry maps section names to their handlers.
// Built once at startup — dynamic sections close over their data source.
var sectionRegistry map[string]SectionHandler

func init() {
	InitSections()
}

// InitSections builds the section registry. Call after SetLiveClient.
// Dynamic sections close over the current liveClient; static sections ignore it.
func InitSections() {
	sectionRegistry = map[string]SectionHandler{
		"hero":         func(_ context.Context) templ.Component { return sections.Hero() },
		"challenge":    func(_ context.Context) templ.Component { return sections.Challenge() },
		"architecture": func(_ context.Context) templ.Component { return sections.Architecture() },
		"features":     func(_ context.Context) templ.Component { return sections.Features() },
		"networking":   func(_ context.Context) templ.Component { return sections.Networking() },
		"comparison":   func(_ context.Context) templ.Component { return sections.Comparison() },
		"usecases":     func(_ context.Context) templ.Component { return sections.UseCases() },
		"getstarted":   func(_ context.Context) templ.Component { return sections.GetStarted() },
		"live": func(ctx context.Context) templ.Component {
			data, _ := liveClient.Fetch(ctx)
			if len(data.Hosts) == 0 && len(data.Services) == 0 {
				data = obs.DefaultMockData()
			}
			return sections.Live(data)
		},
	}
}

// liveClient returns the data source for the live section.
// Used by Home (full page) and SSE stream.
// Defaults to mock if InitSections hasn't been called.
var liveClient obs.Client = &obs.MockClient{Data: obs.DefaultMockData()}

// SetLiveClient configures the data source for the live section.
func SetLiveClient(c obs.Client) {
	if c != nil {
		liveClient = c
	}
}

// Render adapts a templ.Component to a Fiber response.
func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// Home renders the full landing page.
// Falls back to mock data when SigNoz returns empty results.
func Home(c *fiber.Ctx) error {
	data, _ := liveClient.Fetch(c.Context())
	if len(data.Hosts) == 0 && len(data.Services) == 0 {
		data = obs.DefaultMockData()
	}
	return render(c, pages.Home(data))
}

// Section renders an individual section for HTMX partial swaps.
// All sections — static and dynamic — go through the same code path.
func Section(c *fiber.Ctx) error {
	name := c.Params("name")

	handler, ok := sectionRegistry[name]
	if !ok {
		return fiber.NewError(fiber.StatusNotFound)
	}

	return render(c, handler(c.Context()))
}
