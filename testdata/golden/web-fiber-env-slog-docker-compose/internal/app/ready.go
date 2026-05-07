package app

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func ReadyHandler(checker Checker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := Check(context.Background(), checker); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ready"})
	}
}
