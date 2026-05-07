package middleware

import (
	"time"

	"github.com/example/orders-api/internal/logging"
	"github.com/labstack/echo/v4"
)

func Logging(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			status := c.Response().Status
			if status == 0 {
				status = httpStatusOK
			}

			logger.Info(
				"request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"duration", time.Since(start).String(),
				"request_id", requestIDFromEcho(c),
			)
			return err
		}
	}
}

const httpStatusOK = 200
