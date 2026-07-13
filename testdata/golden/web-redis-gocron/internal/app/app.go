package app

import (
	"context"
	"errors"
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

	logger        logging.Logger
	server        *server.Server
	redisClient   *database.RedisClient
	taskScheduler *scheduler.TaskScheduler
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

	redisClient, err := database.NewRedisClient(ctx, cfg.Database.Redis, logger)
	if err != nil {
		return nil, fmt.Errorf("connect redis database: %w", err)
	}
	cleanup = append(cleanup, func() { _ = redisClient.Shutdown(context.Background()) })

	taskScheduler, err := scheduler.NewTaskScheduler(logger, cfg.TaskScheduler.Worker)
	if err != nil {
		return nil, fmt.Errorf("configure task scheduler: %w", err)
	}
	cleanup = append(cleanup, func() { taskScheduler.Shutdown(context.Background()) })

	taskScheduler.AddTasks(tasks.NewExampleCleanupTask(logger, tasks.ExampleCleanupTaskConfig{
		Name:                   cfg.TaskScheduler.Tasks.ExampleCleanup.Name,
		Cron:                   cfg.TaskScheduler.Tasks.ExampleCleanup.Cron,
		ShouldStartImmediately: cfg.TaskScheduler.Tasks.ExampleCleanup.ShouldStartImmediately,
		BatchSize:              cfg.TaskScheduler.Tasks.ExampleCleanup.BatchSize,
		RetentionPeriod:        cfg.TaskScheduler.Tasks.ExampleCleanup.RetentionPeriod,
	}))

	application := &Application{
		ctx:       ctx,
		startedAt: time.Now(),
		logger:    logger,
		server: server.NewServer(cfg.Server, logger, readinessChecks(
			redisClient,
		)),
		redisClient:   redisClient,
		taskScheduler: taskScheduler,
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
	return nil

}

func (a *Application) Shutdown(ctx context.Context) error {
	var shutdownErrors []error

	if a.taskScheduler != nil {
		a.taskScheduler.Shutdown(ctx)
	}
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("shutdown failed", "error", err)
		shutdownErrors = append(shutdownErrors, fmt.Errorf("shutdown server: %w", err))
	}

	if a.redisClient != nil {
		if err := a.redisClient.Shutdown(ctx); err != nil {
			a.logger.Error("redis database shutdown failed", "error", err)
			shutdownErrors = append(shutdownErrors, fmt.Errorf("shutdown redis database: %w", err))
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
