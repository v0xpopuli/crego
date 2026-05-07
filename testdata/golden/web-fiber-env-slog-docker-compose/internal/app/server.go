package app

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app       *fiber.App
	addr      string
	readiness Checker
	logger    Logger
}

func NewServer(cfg ServerConfig, logger Logger, readiness Checker) *Server {
	srv := &Server{
		app: fiber.New(fiber.Config{
			DisableStartupMessage: true,
		}),
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
	return s.app.Listen(s.addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
