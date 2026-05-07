package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		AppEnv        string              `yaml:"app_env"`
		Server        ServerConfig        `yaml:"server"`
		Logging       LoggingConfig       `yaml:"logging"`
		Database      DatabaseConfig      `yaml:"database"`
		TaskScheduler TaskSchedulerConfig `yaml:"task_scheduler"`
	}

	ServerConfig struct {
		Port int `yaml:"port"`
	}

	LoggingConfig struct {
		Format string `yaml:"format"`
	}

	DatabaseConfig struct {
		Postgres PostgresConfig `yaml:"postgres"`
	}

	PostgresConfig struct {
		UserName           string `yaml:"user-name"`
		Password           string `yaml:"password"`
		Host               string `yaml:"host"`
		Database           string `yaml:"database"`
		SSLMode            string `yaml:"ssl-mode"`
		Migrations         string `yaml:"migrations"`
		MaxOpenConnections int    `yaml:"max-open-connections"`
		MaxIdleConnections int    `yaml:"max-idle-connections"`
	}

	TaskSchedulerConfig struct {
		Worker string      `yaml:"worker"`
		Tasks  TaskConfigs `yaml:"tasks"`
	}

	TaskConfigs struct {
		ExampleCleanup ExampleCleanupTaskConfig `yaml:"example_cleanup"`
	}

	ExampleCleanupTaskConfig struct {
		Name                   string `yaml:"name"`
		Cron                   string `yaml:"cron"`
		ShouldStartImmediately bool   `yaml:"should_start_immediately"`
		BatchSize              int    `yaml:"batch_size"`
		RetentionPeriod        string `yaml:"retention_period"`
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
			Format: "json",
		},
		Database: DatabaseConfig{
			Postgres: PostgresConfig{
				UserName:           "postgres",
				Password:           "postgres",
				Host:               "localhost:5432",
				Database:           "app",
				SSLMode:            "",
				Migrations:         "file://scripts/migrations",
				MaxOpenConnections: 10,
				MaxIdleConnections: 5,
			},
		},
		TaskScheduler: TaskSchedulerConfig{
			Worker: "crego-task-executor",
			Tasks: TaskConfigs{
				ExampleCleanup: ExampleCleanupTaskConfig{
					Name:                   "example-cleanup",
					Cron:                   "0 2 * * *",
					ShouldStartImmediately: false,
					BatchSize:              500,
					RetentionPeriod:        "72h",
				},
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

	if value := os.Getenv("TASK_SCHEDULER_WORKER"); value != "" {
		cfg.TaskScheduler.Worker = value
	}
	if value := os.Getenv("TASK_SCHEDULER_EXAMPLE_CLEANUP_NAME"); value != "" {
		cfg.TaskScheduler.Tasks.ExampleCleanup.Name = value
	}
	if value := os.Getenv("TASK_SCHEDULER_EXAMPLE_CLEANUP_CRON"); value != "" {
		cfg.TaskScheduler.Tasks.ExampleCleanup.Cron = value
	}
	if value := os.Getenv("TASK_SCHEDULER_EXAMPLE_CLEANUP_SHOULD_START_IMMEDIATELY"); value != "" {
		shouldStartImmediately, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse TASK_SCHEDULER_EXAMPLE_CLEANUP_SHOULD_START_IMMEDIATELY: %w", err)
		}
		cfg.TaskScheduler.Tasks.ExampleCleanup.ShouldStartImmediately = shouldStartImmediately
	}
	if value := os.Getenv("TASK_SCHEDULER_EXAMPLE_CLEANUP_BATCH_SIZE"); value != "" {
		batchSize, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse TASK_SCHEDULER_EXAMPLE_CLEANUP_BATCH_SIZE: %w", err)
		}
		cfg.TaskScheduler.Tasks.ExampleCleanup.BatchSize = batchSize
	}
	if value := os.Getenv("TASK_SCHEDULER_EXAMPLE_CLEANUP_RETENTION_PERIOD"); value != "" {
		cfg.TaskScheduler.Tasks.ExampleCleanup.RetentionPeriod = value
	}

	if value := os.Getenv("DATABASE_POSTGRES_USER_NAME"); value != "" {
		cfg.Database.Postgres.UserName = value
	}
	if value := os.Getenv("DATABASE_POSTGRES_PASSWORD"); value != "" {
		cfg.Database.Postgres.Password = value
	}
	if value := os.Getenv("DATABASE_POSTGRES_HOST"); value != "" {
		cfg.Database.Postgres.Host = value
	}
	if value := os.Getenv("DATABASE_POSTGRES_DATABASE"); value != "" {
		cfg.Database.Postgres.Database = value
	}
	if value := os.Getenv("DATABASE_POSTGRES_SSL_MODE"); value != "" {
		cfg.Database.Postgres.SSLMode = value
	}
	if value := os.Getenv("DATABASE_POSTGRES_MAX_OPEN_CONNECTIONS"); value != "" {
		maxOpenConnections, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_POSTGRES_MAX_OPEN_CONNECTIONS: %w", err)
		}
		cfg.Database.Postgres.MaxOpenConnections = maxOpenConnections
	}
	if value := os.Getenv("DATABASE_POSTGRES_MAX_IDLE_CONNECTIONS"); value != "" {
		maxIdleConnections, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_POSTGRES_MAX_IDLE_CONNECTIONS: %w", err)
		}
		cfg.Database.Postgres.MaxIdleConnections = maxIdleConnections
	}

	if value := os.Getenv("DATABASE_MIGRATIONS"); value != "" {
		cfg.Database.Postgres.Migrations = value
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

	if cfg.TaskScheduler.Worker == "" {
		return fmt.Errorf("task_scheduler.worker is required")
	}
	if cfg.TaskScheduler.Tasks.ExampleCleanup.Name == "" {
		return fmt.Errorf("task_scheduler.tasks.example_cleanup.name is required")
	}
	if cfg.TaskScheduler.Tasks.ExampleCleanup.Cron == "" {
		return fmt.Errorf("task_scheduler.tasks.example_cleanup.cron is required")
	}
	if cfg.TaskScheduler.Tasks.ExampleCleanup.BatchSize < 1 {
		return fmt.Errorf("task_scheduler.tasks.example_cleanup.batch_size must be greater than zero")
	}
	if _, err := time.ParseDuration(cfg.TaskScheduler.Tasks.ExampleCleanup.RetentionPeriod); err != nil {
		return fmt.Errorf("parse task_scheduler.tasks.example_cleanup.retention_period: %w", err)
	}

	if cfg.Database.Postgres.UserName == "" {
		return fmt.Errorf("database.postgres.user-name is required")
	}
	if cfg.Database.Postgres.Host == "" {
		return fmt.Errorf("database.postgres.host is required")
	}
	if cfg.Database.Postgres.Database == "" {
		return fmt.Errorf("database.postgres.database is required")
	}
	if cfg.Database.Postgres.MaxOpenConnections < 1 {
		return fmt.Errorf("database.postgres.max-open-connections must be greater than zero")
	}
	if cfg.Database.Postgres.MaxIdleConnections < 0 {
		return fmt.Errorf("database.postgres.max-idle-connections must be zero or greater")
	}

	if cfg.Database.Postgres.Migrations == "" {
		return fmt.Errorf("database.postgres.migrations is required")
	}
	return nil
}
