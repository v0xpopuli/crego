package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/orders-api/internal/logging"
	"github.com/go-co-op/gocron/v2"

	"github.com/example/orders-api/internal/config"
)

type (
	TaskScheduler struct {
		logger    logging.Logger
		scheduler gocron.Scheduler
		tasks     []Task
	}

	Task interface {
		Name() string
		Cron() string
		ShouldStartImmediately() bool
		Runnable(context.Context)
	}
)

func NewTaskScheduler(logger logging.Logger, cfg config.TaskSchedulerConfig) (*TaskScheduler, error) {
	_ = cfg
	inner, err := gocron.NewScheduler()

	if err != nil {
		return nil, fmt.Errorf("configure task scheduler: %w", err)
	}
	return &TaskScheduler{
		logger:    logger,
		scheduler: inner,
	}, nil
}

func (s *TaskScheduler) AddTasks(tasks ...Task) {
	s.tasks = append(s.tasks, tasks...)
}

func (s *TaskScheduler) Run(ctx context.Context) {
	if s == nil || s.scheduler == nil {
		return
	}
	for _, task := range s.tasks {
		jobOptions := []gocron.JobOption{
			gocron.WithName(task.Name()),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		}
		if task.ShouldStartImmediately() {
			jobOptions = append(jobOptions, gocron.WithStartAt(gocron.WithStartImmediately()))
		}

		_, err := s.scheduler.NewJob(
			gocron.CronJob(task.Cron(), cronHasSeconds(task.Cron())),
			gocron.NewTask(func(task Task) func(context.Context) {
				return func(jobCtx context.Context) {
					s.runTask(jobCtx, task)
				}
			}(task)),
			jobOptions...,
		)
		if err != nil {
			s.logError("task registration failed", "task", task.Name(), "error", err)
			continue
		}
		s.logInfo("task registered", "task", task.Name(), "cron", task.Cron(), "start_immediately", task.ShouldStartImmediately())
	}

	s.scheduler.Start()
	s.logInfo("task scheduler started", "tasks", len(s.tasks))
	_ = ctx
}

func (s *TaskScheduler) Shutdown(ctx context.Context) {
	if s == nil || s.scheduler == nil {
		return
	}
	if err := s.scheduler.ShutdownWithContext(ctx); err != nil {
		s.logError("task scheduler shutdown failed", "error", err)
		return
	}

	s.logInfo("task scheduler stopped")
}

func (s *TaskScheduler) runTask(ctx context.Context, task Task) {
	startedAt := time.Now()
	s.logInfo("task started", "task", task.Name())
	task.Runnable(ctx)
	s.logInfo("task completed", "task", task.Name(), "took", time.Since(startedAt).String())
}

func cronHasSeconds(expression string) bool {
	return len(strings.Fields(expression)) == 6
}

func (s *TaskScheduler) logInfo(message string, args ...any) {
	if s.logger != nil {
		s.logger.Info(message, args...)
	}
}

func (s *TaskScheduler) logError(message string, args ...any) {
	if s.logger != nil {
		s.logger.Error(message, args...)
	}
}
