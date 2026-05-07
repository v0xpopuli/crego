package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ReadyHandler(checker Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := Check(c.Request.Context(), checker); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
