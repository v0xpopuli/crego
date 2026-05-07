package app

import (
	"context"
	"fmt"
	"time"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/database"
	"github.com/example/orders-api/internal/logging"
	"github.com/example/orders-api/internal/scheduler"
	"github.com/example/orders-api/internal/scheduler/tasks"
	"github.com/example/orders-api/internal/server"
	"github.com/example/orders-api/internal/server/handler"
)

type Application struct {
	ctx       context.Context
	startedAt time.Time

	logger         logging.Logger
	server         *server.Server
	postgresClient *database.PostgresClient
	taskScheduler  *scheduler.TaskScheduler
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger, err := logging.NewLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("configure logger: %w", err)
	}

	postgresClient, err := database.NewPostgresClient(ctx, cfg.Database.Postgres, logger)
	if err != nil {
		return nil, fmt.Errorf("connect postgres database: %w", err)
	}

	taskScheduler, err := scheduler.NewTaskScheduler(logger, cfg.TaskScheduler)
	if err != nil {
		return nil, fmt.Errorf("configure task scheduler: %w", err)
	}

	taskScheduler.AddTasks(tasks.NewExampleCleanupTask(logger, cfg.TaskScheduler.Tasks.ExampleCleanup))

	return &Application{
		ctx:       ctx,
		startedAt: time.Now(),
		logger:    logger,
		server: server.NewServer(cfg.Server, logger, readinessChecks(
			postgresClient,
		)),
		postgresClient: postgresClient,
		taskScheduler:  taskScheduler,
	}, nil
}

func (a *Application) Run() error {
	defer a.logger.Sync()

	if err := a.postgresClient.RunMigrations(a.ctx); err != nil {
		return fmt.Errorf("run postgres migrations: %w", err)
	}

	if a.taskScheduler != nil {
		a.taskScheduler.Run(a.ctx)
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

	if a.taskScheduler != nil {
		a.taskScheduler.Shutdown(ctx)
	}
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("shutdown failed", "error", err)
		return fmt.Errorf("shutdown server: %w", err)
	}

	if a.postgresClient != nil {
		if err := a.postgresClient.Shutdown(ctx); err != nil {
			a.logger.Error("postgres database shutdown failed", "error", err)
			return fmt.Errorf("shutdown postgres database: %w", err)
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
