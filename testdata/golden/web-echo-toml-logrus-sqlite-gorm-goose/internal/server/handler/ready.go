package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ReadyHandler(checker Checker) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := Check(c.Request().Context(), checker); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	}
}
