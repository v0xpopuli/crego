package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		AppEnv   string         `yaml:"app_env"`
		Server   ServerConfig   `yaml:"server"`
		Logging  LoggingConfig  `yaml:"logging"`
		Database DatabaseConfig `yaml:"database"`
	}

	ServerConfig struct {
		Port int `yaml:"port"`
	}

	LoggingConfig struct {
		Format string `yaml:"format"`
	}

	DatabaseConfig struct {
		MongoDB MongoDBConfig `yaml:"mongodb"`
	}

	MongoDBConfig struct {
		UserName string `yaml:"user-name"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Database string `yaml:"database"`
	}

	Options struct {
		Path string
	}
)

func LoadConfig(opts Options) (Config, error) {
	cfg := defaultConfig()

	path := resolveConfigPath(opts.Path, "configs/config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
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
			Format: "text",
		},
		Database: DatabaseConfig{
			MongoDB: MongoDBConfig{
				UserName: "",
				Password: "",
				Host:     "localhost:27017",
				Database: "app",
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

	if value := os.Getenv("DATABASE_MONGODB_USER_NAME"); value != "" {
		cfg.Database.MongoDB.UserName = value
	}
	if value := os.Getenv("DATABASE_MONGODB_PASSWORD"); value != "" {
		cfg.Database.MongoDB.Password = value
	}
	if value := os.Getenv("DATABASE_MONGODB_HOST"); value != "" {
		cfg.Database.MongoDB.Host = value
	}
	if value := os.Getenv("DATABASE_MONGODB_DATABASE"); value != "" {
		cfg.Database.MongoDB.Database = value
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

	if cfg.Database.MongoDB.Host == "" {
		return fmt.Errorf("database.mongodb.host is required")
	}
	if cfg.Database.MongoDB.Database == "" {
		return fmt.Errorf("database.mongodb.database is required")
	}
	return nil
}
