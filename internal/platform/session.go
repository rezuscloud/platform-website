package platform

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func EnsureSessionID(c *fiber.Ctx) string {
	if sessionID := c.Cookies(SessionCookieName); sessionID != "" {
		return sessionID
	}

	sessionID := newSessionID()
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   SessionCookieMaxAge,
	})

	return sessionID
}

func newSessionID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("generate session id: %w", err))
	}

	return hex.EncodeToString(buf)
}
