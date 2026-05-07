package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/example/orders-api/internal/config"
)

type Logger interface {
	Info(message string, args ...any)
	Error(message string, args ...any)
	Sync()
}

type defaultLogger struct {
	inner *zap.SugaredLogger
}

func NewLogger(cfg config.LoggingConfig) (Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel)
	return &defaultLogger{inner: zap.New(core).Sugar()}, nil
}

func (l *defaultLogger) Info(message string, args ...any) {
	l.inner.Infow(message, args...)
}

func (l *defaultLogger) Error(message string, args ...any) {
	l.inner.Errorw(message, args...)
}

func (l *defaultLogger) Sync() {
	_ = l.inner.Sync()
}
