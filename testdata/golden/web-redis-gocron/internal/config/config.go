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
		Redis RedisConfig `yaml:"redis"`
	}

	RedisConfig struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
		Database int    `yaml:"database"`
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
			Format: "text",
		},
		Database: DatabaseConfig{
			Redis: RedisConfig{
				Host:     "localhost:6379",
				Password: "",
				Database: 0,
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

	if value := os.Getenv("DATABASE_REDIS_HOST"); value != "" {
		cfg.Database.Redis.Host = value
	}
	if value := os.Getenv("DATABASE_REDIS_PASSWORD"); value != "" {
		cfg.Database.Redis.Password = value
	}
	if value := os.Getenv("DATABASE_REDIS_DATABASE"); value != "" {
		database, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("parse DATABASE_REDIS_DATABASE: %w", err)
		}
		cfg.Database.Redis.Database = database
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

	if cfg.Database.Redis.Host == "" {
		return fmt.Errorf("database.redis.host is required")
	}
	if cfg.Database.Redis.Database < 0 {
		return fmt.Errorf("database.redis.database must be zero or greater")
	}
	return nil
}
