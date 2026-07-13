package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/database"
	"github.com/example/orders-api/internal/logging"
	"github.com/example/orders-api/internal/server"
	"github.com/example/orders-api/internal/server/handler"
)

type Application struct {
	ctx       context.Context
	startedAt time.Time

	logger      logging.Logger
	server      *server.Server
	mysqlClient *database.MySQLClient
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("configure logger: %w", err)
	}
	cleanup := []func(){logger.Sync}
	defer func() {
		for index := len(cleanup) - 1; index >= 0; index-- {
			cleanup[index]()
		}
	}()

	mysqlClient, err := database.NewMySQLClient(ctx, cfg.Database.MySQL, logger)
	if err != nil {
		return nil, fmt.Errorf("connect mysql database: %w", err)
	}
	cleanup = append(cleanup, func() { _ = mysqlClient.Shutdown(context.Background()) })
	application := &Application{
		ctx:       ctx,
		startedAt: time.Now(),
		logger:    logger,
		server: server.NewServer(cfg.Server, logger, readinessChecks(
			mysqlClient,
		)),
		mysqlClient: mysqlClient,
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

	if err := a.mysqlClient.RunMigrations(a.ctx); err != nil {
		return fmt.Errorf("run mysql migrations: %w", err)
	}

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

	if a.mysqlClient != nil {
		if err := a.mysqlClient.Shutdown(ctx); err != nil {
			a.logger.Error("mysql database shutdown failed", "error", err)
			shutdownErrors = append(shutdownErrors, fmt.Errorf("shutdown mysql database: %w", err))
		}
	}
	return errors.Join(shutdownErrors...)
}

func readinessChecks(checkers ...handler.Checker) handler.Checks {
	checks := make(handler.Checks, 0, len(checkers))
	for _, checker := range checkers {
		checks = append(checks, checker)
	}
	return checks
}
