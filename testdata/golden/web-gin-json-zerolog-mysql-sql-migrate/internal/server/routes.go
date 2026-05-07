package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/orders-api/internal/server/handler"
	"github.com/example/orders-api/internal/server/middleware"
)

func (s *Server) routes() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.Recover(s.logger), middleware.RequestID(), middleware.Logging(s.logger))

	router.GET("/health", handler.Health)

	router.GET("/ready", handler.ReadyHandler(s.readiness))

	return router
}
