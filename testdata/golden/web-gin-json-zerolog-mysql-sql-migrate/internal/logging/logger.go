package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/example/orders-api/internal/config"
)

type Logger interface {
	Info(message string, args ...any)
	Error(message string, args ...any)
	Sync()
}

type defaultLogger struct {
	inner zerolog.Logger
}

func NewLogger(cfg config.LoggingConfig) (Logger, error) {
	var logger zerolog.Logger
	if cfg.Format == "json" {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		logger = zerolog.New(writer).With().Timestamp().Logger()
	}
	return &defaultLogger{inner: logger}, nil
}

func (l *defaultLogger) Info(message string, args ...any) {
	l.inner.Info().Fields(logFields(args...)).Msg(message)
}

func (l *defaultLogger) Error(message string, args ...any) {
	l.inner.Error().Fields(logFields(args...)).Msg(message)
}

func (l *defaultLogger) Sync() {}

func logFields(args ...any) map[string]any {
	fields := make(map[string]any, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok || key == "" {
			continue
		}
		fields[key] = args[i+1]
	}
	return fields
}
