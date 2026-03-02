package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/version"
)

func APIVersion(c *fiber.Ctx) error {
	return c.JSON(version.Get())
}
