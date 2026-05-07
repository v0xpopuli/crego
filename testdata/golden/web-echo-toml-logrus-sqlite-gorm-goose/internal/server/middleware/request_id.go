package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func RequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := c.Request().Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set(requestIDKey, requestID)
		c.Response().Header().Set(requestIDHeader, requestID)
		return next(c)
	}
}

func requestIDFromEcho(c echo.Context) string {
	requestID, _ := c.Get(requestIDKey).(string)
	return requestID
}

func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
