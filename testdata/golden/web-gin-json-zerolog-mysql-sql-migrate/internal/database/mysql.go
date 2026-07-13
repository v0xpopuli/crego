package database

import (
	"context"
	"database/sql"
	"fmt"

	"net/url"

	"time"

	mysqldriver "github.com/go-sql-driver/mysql"

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
	driverConfig := mysqldriver.NewConfig()
	driverConfig.User = cfg.UserName
	driverConfig.Passwd = cfg.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = cfg.Host
	driverConfig.DBName = cfg.Database
	driverConfig.ParseTime = cfg.ParseTime
	return driverConfig.FormatDSN()
}

func mysqlMigrationURL(cfg config.MySQLConfig) string {
	credentials := ""
	if cfg.UserName != "" {
		credentials = url.UserPassword(cfg.UserName, cfg.Password).String() + "@"
	}
	address := fmt.Sprintf("mysql://%stcp(%s)/%s", credentials, cfg.Host, url.PathEscape(cfg.Database))
	if cfg.ParseTime {
		address += "?parseTime=true"
	}
	return address
}
