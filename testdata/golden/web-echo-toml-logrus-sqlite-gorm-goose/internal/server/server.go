package server

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/logging"
	"github.com/example/orders-api/internal/server/handler"
)

type Server struct {
	echo      *echo.Echo
	addr      string
	readiness handler.Checker
	logger    logging.Logger
}

func NewServer(cfg config.ServerConfig, logger logging.Logger, readiness handler.Checker) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	srv := &Server{
		echo:      e,
		addr:      ":" + strconv.Itoa(cfg.Port),
		readiness: readiness,
		logger:    logger,
	}
	srv.registerRoutes()
	return srv
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Start() error {
	err := s.echo.Start(s.addr)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
