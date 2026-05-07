package middleware

import (
	"time"

	"github.com/example/orders-api/internal/logging"
	"github.com/gin-gonic/gin"
)

func Logging(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info(
			"request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
			"request_id", requestIDFromGin(c),
		)
	}
}
