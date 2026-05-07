package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set(requestIDKey, requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

func requestIDFromGin(c *gin.Context) string {
	requestID, _ := c.Get(requestIDKey)
	value, _ := requestID.(string)
	return value
}

func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
