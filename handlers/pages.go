package handlers

import (
	"context"
	"sync"

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
		"architecture": func(_ context.Context) templ.Component { return sections.Architecture() },
		"features":     func(_ context.Context) templ.Component { return sections.Features() },
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

// lastGoodClient wraps an obs.Client and serves the last successful,
// non-empty snapshot when the inner client errors or returns an empty
// fetch. A "Live" section must degrade to stale data, not to nothing:
// under transient Prometheus/K8s API pressure an empty result would blank
// the whole section (nodes, metrics and all) for a page load or an SSE tick.
type lastGoodClient struct {
	inner obs.Client
	mu    sync.RWMutex
	last  obs.LiveData
	have  bool
}

func (l *lastGoodClient) Fetch(ctx context.Context) (obs.LiveData, error) {
	data, err := l.inner.Fetch(ctx)
	if err == nil && (len(data.Hosts) > 0 || len(data.Services) > 0) {
		l.mu.Lock()
		l.last, l.have = data, true
		l.mu.Unlock()
		return data, nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.have {
		return l.last, nil
	}
	return data, err
}

// SetLiveClient configures the data source for the live section.
func SetLiveClient(c obs.Client) {
	if c != nil {
		liveClient = &lastGoodClient{inner: c}
	}
}

// Render adapts a templ.Component to a Fiber response.
//
// HTML responses use Cache-Control: no-cache so browsers always revalidate
// against the origin after a deploy. Without it (and with no ETag/Last-Modified
// on these responses) a browser can hold a pre-deploy page across deploys.
// Static assets keep their own long cache via app.Static in main.go.
func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-cache")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// Home renders the full landing page.
// Falls back to mock data when the live source returns empty results.
func Home(c *fiber.Ctx) error {
	data, _ := liveClient.Fetch(c.Context())
	if len(data.Hosts) == 0 && len(data.Services) == 0 {
		data = obs.DefaultMockData()
	}
	return render(c, pages.Home(data))
}

// Privacy renders the privacy notice. The site collects no user data, so this
// is a plain declaration, not a consent flow.
func Privacy(c *fiber.Ctx) error {
	return render(c, pages.Privacy())
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
