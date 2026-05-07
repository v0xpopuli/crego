package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type (
	Config struct {
		AppEnv   string         `json:"app_env"`
		Server   ServerConfig   `json:"server"`
		Logging  LoggingConfig  `json:"logging"`
		Database DatabaseConfig `json:"database"`
	}

	ServerConfig struct {
		Port int `json:"port"`
	}

	LoggingConfig struct {
		Format string `json:"format"`
	}

	DatabaseConfig struct {
		MySQL MySQLConfig `json:"mysql"`
	}

	MySQLConfig struct {
		UserName           string `json:"user-name"`
		Password           string `json:"password"`
		Host               string `json:"host"`
		Database           string `json:"database"`
		ParseTime          bool   `json:"parse-time"`
		Migrations         string `json:"migrations"`
		MaxOpenConnections int    `json:"max-open-connections"`
		MaxIdleConnections int    `json:"max-idle-connections"`
	}

	Options struct {
		Path string
	}
)

func LoadConfig(opts Options) (Config, error) {
	cfg := defaultConfig()

	path := resolveConfigPath(opts.Path, "configs/config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
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
			MySQL: MySQLConfig{
				UserName:           "root",
				Password:           "root",
				Host:               "localhost:3306",
				Database:           "app",
				ParseTime:          true,
				Migrations:         "file://scripts/migrations",
				MaxOpenConnections: 10,
				MaxIdleConnections: 5,
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

	if value := os.Getenv("DATABASE_MYSQL_USER_NAME"); value != "" {
		cfg.Database.MySQL.UserName = value
	}
	if value := os.Getenv("DATABASE_MYSQL_PASSWORD"); value != "" {
		cfg.Database.MySQL.Password = value
	}
	if value := os.Getenv("DATABASE_MYSQL_HOST"); value != "" {
		cfg.Database.MySQL.Host = value
	}
	if value := os.Getenv("DATABASE_MYSQL_DATABASE"); value != "" {
		cfg.Database.MySQL.Database = value
	}
	if value := os.Getenv("DATABASE_MYSQL_PARSE_TIME"); value != "" {
		parseTime, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_MYSQL_PARSE_TIME: %w", err)
		}
		cfg.Database.MySQL.ParseTime = parseTime
	}
	if value := os.Getenv("DATABASE_MYSQL_MAX_OPEN_CONNECTIONS"); value != "" {
		maxOpenConnections, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_MYSQL_MAX_OPEN_CONNECTIONS: %w", err)
		}
		cfg.Database.MySQL.MaxOpenConnections = maxOpenConnections
	}
	if value := os.Getenv("DATABASE_MYSQL_MAX_IDLE_CONNECTIONS"); value != "" {
		maxIdleConnections, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_MYSQL_MAX_IDLE_CONNECTIONS: %w", err)
		}
		cfg.Database.MySQL.MaxIdleConnections = maxIdleConnections
	}

	if value := os.Getenv("DATABASE_MIGRATIONS"); value != "" {
		cfg.Database.MySQL.Migrations = value
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

	if cfg.Database.MySQL.UserName == "" {
		return fmt.Errorf("database.mysql.user-name is required")
	}
	if cfg.Database.MySQL.Host == "" {
		return fmt.Errorf("database.mysql.host is required")
	}
	if cfg.Database.MySQL.Database == "" {
		return fmt.Errorf("database.mysql.database is required")
	}
	if cfg.Database.MySQL.MaxOpenConnections < 1 {
		return fmt.Errorf("database.mysql.max-open-connections must be greater than zero")
	}
	if cfg.Database.MySQL.MaxIdleConnections < 0 {
		return fmt.Errorf("database.mysql.max-idle-connections must be zero or greater")
	}

	if cfg.Database.MySQL.Migrations == "" {
		return fmt.Errorf("database.mysql.migrations is required")
	}
	return nil
}
