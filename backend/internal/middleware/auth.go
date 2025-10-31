package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func APIKeyAuth(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization format. Use: Bearer <api_key>",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		if token != apiKey {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		return c.Next()
	}
}
