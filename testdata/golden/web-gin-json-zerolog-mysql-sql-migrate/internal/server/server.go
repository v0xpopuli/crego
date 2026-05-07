package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/logging"
	"github.com/example/orders-api/internal/server/handler"
)

type Server struct {
	httpServer *http.Server
	readiness  handler.Checker
	logger     logging.Logger
}

func NewServer(cfg config.ServerConfig, logger logging.Logger, readiness handler.Checker) *Server {
	srv := &Server{logger: logger, readiness: readiness}
	srv.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           srv.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return srv
}

func (s *Server) Addr() string {
	return s.httpServer.Addr
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
