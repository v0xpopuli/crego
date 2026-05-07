package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/orders-api/internal/server/handler"
	"github.com/example/orders-api/internal/server/middleware"
)

func (s *Server) routes() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Recover(s.logger))
	router.Use(middleware.RequestID)
	router.Use(middleware.Logging(s.logger))

	router.Get("/health", handler.Health)

	router.Get("/ready", handler.ReadyHandler(s.readiness))

	return router
}
