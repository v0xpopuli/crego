package app

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func RequestID(c *fiber.Ctx) error {
	requestID := c.Get(requestIDHeader)
	if requestID == "" {
		requestID = generateRequestID()
	}
	c.Locals(requestIDKey, requestID)
	c.Set(requestIDHeader, requestID)
	return c.Next()
}

func requestIDFromFiber(c *fiber.Ctx) string {
	requestID, _ := c.Locals(requestIDKey).(string)
	return requestID
}

func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
