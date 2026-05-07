package logging

import (
	"log/slog"
	"os"

	"github.com/example/orders-api/internal/config"
)

type Logger interface {
	Info(message string, args ...any)
	Error(message string, args ...any)
	Sync()
}

type defaultLogger struct {
	inner *slog.Logger
}

func NewLogger(cfg config.LoggingConfig) (Logger, error) {
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	return &defaultLogger{inner: slog.New(handler)}, nil
}

func (l *defaultLogger) Info(message string, args ...any) {
	l.inner.Info(message, args...)
}

func (l *defaultLogger) Error(message string, args ...any) {
	l.inner.Error(message, args...)
}

func (l *defaultLogger) Sync() {}
