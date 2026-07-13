package tasks

import (
	"context"
	"time"
)

type (
	ExampleCleanupTask struct {
		logger Logger
		config ExampleCleanupTaskConfig
	}

	ExampleCleanupTaskConfig struct {
		Name                   string
		Cron                   string
		ShouldStartImmediately bool
		BatchSize              int
		RetentionPeriod        string
	}

	Logger interface {
		Info(message string, args ...any)
		Error(message string, args ...any)
	}
)

func NewExampleCleanupTask(logger Logger, cfg ExampleCleanupTaskConfig) *ExampleCleanupTask {
	return &ExampleCleanupTask{logger: logger, config: cfg}
}

func (t *ExampleCleanupTask) Name() string {
	return t.config.Name
}

func (t *ExampleCleanupTask) Cron() string {
	return t.config.Cron
}

func (t *ExampleCleanupTask) ShouldStartImmediately() bool {
	return t.config.ShouldStartImmediately
}

func (t *ExampleCleanupTask) Runnable(ctx context.Context) {
	startedAt := time.Now()
	retentionPeriod, err := time.ParseDuration(t.config.RetentionPeriod)
	if err != nil && t.logger != nil {
		t.logger.Error("example cleanup retention period is invalid", "task", t.Name(), "retention_period", t.config.RetentionPeriod, "error", err)
	}
	if t.logger != nil {
		t.logger.Info(
			"example cleanup started",
			"task", t.Name(),
			"batch_size", t.config.BatchSize,
			"retention_period", retentionPeriod.String(),
		)
	}

	select {
	case <-ctx.Done():
		if t.logger != nil {
			t.logger.Error("example cleanup canceled", "task", t.Name(), "error", ctx.Err())
		}
		return
	default:
	}

	if t.logger != nil {
		t.logger.Info(
			"example cleanup completed",
			"task", t.Name(),
			"batch_size", t.config.BatchSize,
			"retention_period", retentionPeriod.String(),
			"took", time.Since(startedAt).String(),
		)
	}
}
