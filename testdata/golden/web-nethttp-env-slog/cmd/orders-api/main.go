package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/orders-api/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "run application: %v\n", err)
		os.Exit(1)
	}
}

func loadConfig() (app.Config, error) {
	return app.LoadConfig(app.Options{})
}
