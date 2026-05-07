package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/orders-api/internal/app"
	"github.com/example/orders-api/internal/config"
)

const (
	configArgumentName = "config"
	configDefaultPath  = "configs/config.toml"
	configUsageMessage = "Path to the configuration file"
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

func loadConfig() (config.Config, error) {
	return config.LoadConfig(config.Options{Path: getConfigPath()})
}

func getConfigPath() string {
	configPath := flag.String(configArgumentName, "", configUsageMessage)
	flag.Parse()
	if *configPath != "" {
		return *configPath
	}
	if os.Getenv("CONFIG_PATH") != "" {
		return ""
	}
	return configDefaultPath
}
