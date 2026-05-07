package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/logging"
)

type SQLiteClient struct {
	config  config.SQLiteConfig
	address string
	logger  logging.Logger

	Database *gorm.DB
	sqlDB    *sql.DB
}

func NewSQLiteClient(ctx context.Context, cfg config.SQLiteConfig, logger logging.Logger) (*SQLiteClient, error) {
	address := sqliteDSN(cfg)

	database, err := gorm.Open(sqlite.Open(address), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("open sqlite sql database: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	client := &SQLiteClient{config: cfg, address: address, logger: logger, Database: database, sqlDB: sqlDB}
	if err := client.Ping(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if logger != nil {
		logger.Info("database connected", "driver", "sqlite", "framework", "gorm")
	}
	return client, nil
}

func (c *SQLiteClient) Ping(ctx context.Context) error {

	return c.sqlDB.PingContext(ctx)

}

func (c *SQLiteClient) Shutdown(ctx context.Context) error {
	_ = ctx

	if err := c.sqlDB.Close(); err != nil {
		return err
	}

	if c.logger != nil {
		c.logger.Info("database connection closed", "driver", "sqlite")
	}
	return nil
}

func sqliteDSN(cfg config.SQLiteConfig) string {
	address := "file:" + cfg.Path
	if cfg.ForeignKeys {
		address += "?_foreign_keys=on"
	}
	return address
}
