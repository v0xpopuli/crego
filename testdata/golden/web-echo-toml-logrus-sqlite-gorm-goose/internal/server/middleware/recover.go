package middleware

import (
	"net/http"

	"github.com/example/orders-api/internal/logging"
	"github.com/labstack/echo/v4"
)

func Recover(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "error", recovered, "method", c.Request().Method, "path", c.Request().URL.Path)
					_ = c.JSON(http.StatusInternalServerError, map[string]string{"status": "error"})
				}
			}()
			return next(c)
		}
	}
}
