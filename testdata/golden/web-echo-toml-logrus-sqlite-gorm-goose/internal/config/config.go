package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

type (
	Config struct {
		AppEnv   string         `toml:"app_env"`
		Server   ServerConfig   `toml:"server"`
		Logging  LoggingConfig  `toml:"logging"`
		Database DatabaseConfig `toml:"database"`
	}

	ServerConfig struct {
		Port int `toml:"port"`
	}

	LoggingConfig struct {
		Format string `toml:"format"`
	}

	DatabaseConfig struct {
		SQLite SQLiteConfig `toml:"sqlite"`
	}

	SQLiteConfig struct {
		Path               string `toml:"path"`
		ForeignKeys        bool   `toml:"foreign-keys"`
		Migrations         string `toml:"migrations"`
		MaxOpenConnections int    `toml:"max-open-connections"`
		MaxIdleConnections int    `toml:"max-idle-connections"`
	}

	Options struct {
		Path string
	}
)

func LoadConfig(opts Options) (Config, error) {
	cfg := defaultConfig()

	path := resolveConfigPath(opts.Path, "configs/config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}

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
			Format: "json",
		},
		Database: DatabaseConfig{
			SQLite: SQLiteConfig{
				Path:               "app.db",
				ForeignKeys:        true,
				Migrations:         "file://scripts/migrations",
				MaxOpenConnections: 1,
				MaxIdleConnections: 1,
			},
		},
	}
}

func resolveConfigPath(explicit string, fallback string) string {
	if explicit != "" {
		return explicit
	}
	if value := os.Getenv("CONFIG_PATH"); value != "" {
		return value
	}
	return fallback
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

	if value := os.Getenv("DATABASE_SQLITE_PATH"); value != "" {
		cfg.Database.SQLite.Path = value
	}
	if value := os.Getenv("DATABASE_SQLITE_FOREIGN_KEYS"); value != "" {
		foreignKeys, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_SQLITE_FOREIGN_KEYS: %w", err)
		}
		cfg.Database.SQLite.ForeignKeys = foreignKeys
	}
	if value := os.Getenv("DATABASE_SQLITE_MAX_OPEN_CONNECTIONS"); value != "" {
		maxOpenConnections, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_SQLITE_MAX_OPEN_CONNECTIONS: %w", err)
		}
		cfg.Database.SQLite.MaxOpenConnections = maxOpenConnections
	}
	if value := os.Getenv("DATABASE_SQLITE_MAX_IDLE_CONNECTIONS"); value != "" {
		maxIdleConnections, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_SQLITE_MAX_IDLE_CONNECTIONS: %w", err)
		}
		cfg.Database.SQLite.MaxIdleConnections = maxIdleConnections
	}

	if value := os.Getenv("DATABASE_MIGRATIONS"); value != "" {
		cfg.Database.SQLite.Migrations = value
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

	if cfg.Database.SQLite.Path == "" {
		return fmt.Errorf("database.sqlite.path is required")
	}
	if cfg.Database.SQLite.MaxOpenConnections < 1 {
		return fmt.Errorf("database.sqlite.max-open-connections must be greater than zero")
	}
	if cfg.Database.SQLite.MaxIdleConnections < 0 {
		return fmt.Errorf("database.sqlite.max-idle-connections must be zero or greater")
	}

	if cfg.Database.SQLite.Migrations == "" {
		return fmt.Errorf("database.sqlite.migrations is required")
	}
	return nil
}
