package middleware

import (
	"net/http"

	"github.com/example/orders-api/internal/logging"
	"github.com/gin-gonic/gin"
)

func Recover(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered", "error", recovered, "method", c.Request.Method, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error"})
			}
		}()
		c.Next()
	}
}
