package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// LiveServiceHistory returns the sparkline history for a single service, on
// demand. The detail panel calls this when it opens (and polls while open) so
// history is never shipped on the SSE tick or in the initial page render.
//
// Reuses the same cached topology the SSE stream reads from, so it adds no
// Prometheus load: a cache hit plus a linear scan over the service list.
func LiveServiceHistory(c *fiber.Ctx) error {
	c.Set("Cache-Control", "no-cache")

	namespace := c.Query("namespace")
	name := c.Query("name")
	host := c.Query("host")
	if namespace == "" || name == "" || host == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "namespace, name, and host query params are required"})
	}

	data, err := liveClient.Fetch(c.Context())
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(fiber.Map{"error": "live data unavailable"})
	}

	hist, ok := data.HistoryFor(namespace, name, host)
	if !ok {
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"error": "service not found"})
	}
	return c.JSON(hist)
}
