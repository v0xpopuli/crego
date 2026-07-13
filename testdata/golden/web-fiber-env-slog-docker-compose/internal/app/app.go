package app

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Application struct {
	ctx       context.Context
	startedAt time.Time

	logger Logger
	server *Server
}

func New(ctx context.Context, cfg Config) (*Application, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger, err := NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("configure logger: %w", err)
	}
	cleanup := []func(){logger.Sync}
	defer func() {
		for index := len(cleanup) - 1; index >= 0; index-- {
			cleanup[index]()
		}
	}()

	application := &Application{
		ctx:       ctx,
		startedAt: time.Now(),
		logger:    logger,
		server:    NewServer(cfg.Server, logger, nil),
	}
	cleanup = nil
	return application, nil
}

func (a *Application) Run() (runErr error) {
	defer a.logger.Sync()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			if runErr == nil {
				runErr = err
				return
			}
			a.logger.Error("application shutdown failed", "error", err)
			return
		}
		a.logger.Info("application shut down gracefully")
	}()

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("starting server", "addr", a.server.Addr())
		errCh <- a.server.Start()
	}()

	a.logger.Info("application started", "took", time.Since(a.startedAt).String())

	select {
	case err := <-errCh:
		if err != nil {
			a.logger.Error("server failed", "error", err)
			return fmt.Errorf("run server: %w", err)
		}
		return nil
	case <-a.ctx.Done():
	}

	a.logger.Info("application shutdown initiated")
	return nil

}

func (a *Application) Shutdown(ctx context.Context) error {
	var shutdownErrors []error
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("shutdown failed", "error", err)
		shutdownErrors = append(shutdownErrors, fmt.Errorf("shutdown server: %w", err))
	}
	return errors.Join(shutdownErrors...)
}
