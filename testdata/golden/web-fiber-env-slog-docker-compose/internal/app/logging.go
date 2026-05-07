package app

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logging(logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		logger.Info(
			"request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
			"request_id", requestIDFromFiber(c),
		)
		return err
	}
}
