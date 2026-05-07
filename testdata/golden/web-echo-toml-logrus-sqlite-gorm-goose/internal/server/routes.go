package server

import (
	"github.com/example/orders-api/internal/server/handler"
	"github.com/example/orders-api/internal/server/middleware"
)

func (s *Server) registerRoutes() {
	s.echo.Use(middleware.Recover(s.logger))
	s.echo.Use(middleware.RequestID)
	s.echo.Use(middleware.Logging(s.logger))

	s.echo.GET("/health", handler.Health)

	s.echo.GET("/ready", handler.ReadyHandler(s.readiness))

}
