package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// APIKeyAuth creates a middleware that validates API key from X-API-Key header
func APIKeyAuth(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		key := c.Get("X-API-Key")

		// Validate API key
		if key == "" || key != apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or missing API key",
			})
		}

		return c.Next()
	}
}
