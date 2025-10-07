package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// IPAllowlist creates a middleware that restricts access to allowed IPs
func IPAllowlist(allowedIPs []string) fiber.Handler {
	// If no IPs configured, skip IP check
	if len(allowedIPs) == 0 {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	// Create a map for faster lookup
	ipMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipMap[strings.TrimSpace(ip)] = true
	}

	return func(c *fiber.Ctx) error {
		clientIP := c.IP()

		// Check if IP is in allowlist
		if !ipMap[clientIP] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied: IP not allowed",
			})
		}

		return c.Next()
	}
}
