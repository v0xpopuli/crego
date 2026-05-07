package app

import (
	"fmt"
	"os"
	"strconv"
)

type (
	Config struct {
		AppEnv  string
		Server  ServerConfig
		Logging LoggingConfig
	}

	ServerConfig struct {
		Port int
	}

	LoggingConfig struct {
		Format string
	}

	Options struct {
		Path string
	}
)

func LoadConfig(opts Options) (Config, error) {
	cfg := defaultConfig()

	_ = opts

	if err := applyEnvOverrides(&cfg); err != nil {
		return Config{}, err
	}
	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		AppEnv: "development",
		Server: ServerConfig{
			Port: 8080,
		},
		Logging: LoggingConfig{
			Format: "text",
		},
	}
}

func applyEnvOverrides(cfg *Config) error {
	if value := os.Getenv("APP_ENV"); value != "" {
		cfg.AppEnv = value
	}
	if value := os.Getenv("SERVER_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse SERVER_PORT: %w", err)
		}
		cfg.Server.Port = port
	}
	if value := os.Getenv("LOGGING_FORMAT"); value != "" {
		cfg.Logging.Format = value
	}
	return nil
}

func validateConfig(cfg Config) error {
	if cfg.AppEnv == "" {
		return fmt.Errorf("app_env is required")
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if cfg.Logging.Format != "text" && cfg.Logging.Format != "json" {
		return fmt.Errorf("logging.format must be text or json")
	}
	return nil
}
