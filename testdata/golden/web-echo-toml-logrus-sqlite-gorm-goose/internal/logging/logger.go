package logging

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/example/orders-api/internal/config"
)

type Logger interface {
	Info(message string, args ...any)
	Error(message string, args ...any)
	Sync()
}

type defaultLogger struct {
	inner *logrus.Logger
}

func NewLogger(cfg config.LoggingConfig) (Logger, error) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}
	return &defaultLogger{inner: logger}, nil
}

func (l *defaultLogger) Info(message string, args ...any) {
	l.inner.WithFields(logFields(args...)).Info(message)
}

func (l *defaultLogger) Error(message string, args ...any) {
	l.inner.WithFields(logFields(args...)).Error(message)
}

func (l *defaultLogger) Sync() {}

func logFields(args ...any) logrus.Fields {
	fields := make(logrus.Fields, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok || key == "" {
			continue
		}
		fields[key] = args[i+1]
	}
	return fields
}
