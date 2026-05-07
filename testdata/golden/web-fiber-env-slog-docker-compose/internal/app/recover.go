package app

import (
	"github.com/gofiber/fiber/v2"
)

func Recover(logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered", "error", recovered, "method", c.Method(), "path", c.Path())
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error"})
			}
		}()
		return c.Next()
	}
}
