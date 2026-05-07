package app

import (
	"context"
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

	logger       logging.Logger
	server       *server.Server
	sqliteClient *database.SQLiteClient
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("configure logger: %w", err)
	}

	sqliteClient, err := database.NewSQLiteClient(ctx, cfg.Database.SQLite, logger)
	if err != nil {
		return nil, fmt.Errorf("connect sqlite database: %w", err)
	}
	return &Application{
		ctx:       ctx,
		startedAt: time.Now(),
		logger:    logger,
		server: server.NewServer(cfg.Server, logger, readinessChecks(
			sqliteClient,
		)),
		sqliteClient: sqliteClient,
	}, nil
}

func (a *Application) Run() error {
	defer a.logger.Sync()

	if err := a.sqliteClient.RunMigrations(a.ctx); err != nil {
		return fmt.Errorf("run sqlite migrations: %w", err)
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		return err
	}
	a.logger.Info("application shut down gracefully")
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("shutdown failed", "error", err)
		return fmt.Errorf("shutdown server: %w", err)
	}

	if a.sqliteClient != nil {
		if err := a.sqliteClient.Shutdown(ctx); err != nil {
			a.logger.Error("sqlite database shutdown failed", "error", err)
			return fmt.Errorf("shutdown sqlite database: %w", err)
		}
	}
	return nil
}

func readinessChecks(checkers ...handler.Checker) handler.Checks {
	checks := make(handler.Checks, 0, len(checkers))
	for _, checker := range checkers {
		checks = append(checks, checker)
	}
	return checks
}
