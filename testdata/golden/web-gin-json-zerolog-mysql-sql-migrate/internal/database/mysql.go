package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/logging"
)

type MySQLClient struct {
	config  config.MySQLConfig
	address string
	logger  logging.Logger

	DB *sql.DB
}

func NewMySQLClient(ctx context.Context, cfg config.MySQLConfig, logger logging.Logger) (*MySQLClient, error) {
	address := mysqlDSN(cfg)

	db, err := sql.Open("mysql", address)
	if err != nil {
		return nil, fmt.Errorf("open mysql database: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(30 * time.Minute)
	client := &MySQLClient{config: cfg, address: address, logger: logger, DB: db}
	if err := client.Ping(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql database: %w", err)
	}

	if logger != nil {
		logger.Info("database connected", "driver", "mysql", "framework", "sql")
	}
	return client, nil
}

func (c *MySQLClient) Ping(ctx context.Context) error {

	return c.DB.PingContext(ctx)

}

func (c *MySQLClient) Shutdown(ctx context.Context) error {
	_ = ctx

	if err := c.DB.Close(); err != nil {
		return err
	}

	if c.logger != nil {
		c.logger.Info("database connection closed", "driver", "mysql")
	}
	return nil
}

func mysqlDSN(cfg config.MySQLConfig) string {
	address := fmt.Sprintf("%s:%s@tcp(%s)/%s", cfg.UserName, cfg.Password, cfg.Host, cfg.Database)
	if cfg.ParseTime {
		address += "?parseTime=true"
	}
	return address
}
