package database

import (
	"context"

	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/logging"
)

type PostgresClient struct {
	config  config.PostgresConfig
	address string
	logger  logging.Logger

	Pool *pgxpool.Pool
}

func NewPostgresClient(ctx context.Context, cfg config.PostgresConfig, logger logging.Logger) (*PostgresClient, error) {
	address := postgresAddress(cfg)

	poolConfig, err := pgxpool.ParseConfig(address)
	if err != nil {
		return nil, fmt.Errorf("parse postgres address: %w", err)
	}
	poolConfig.MaxConns = int32(cfg.MaxOpenConnections)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}
	client := &PostgresClient{config: cfg, address: address, logger: logger, Pool: pool}
	if err := client.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres database: %w", err)
	}

	if logger != nil {
		logger.Info("database connected", "driver", "postgres", "framework", "pgx")
	}
	return client, nil
}

func (c *PostgresClient) Ping(ctx context.Context) error {

	return c.Pool.Ping(ctx)

}

func (c *PostgresClient) Shutdown(ctx context.Context) error {

	_ = ctx
	c.Pool.Close()

	if c.logger != nil {
		c.logger.Info("database connection closed", "driver", "postgres")
	}
	return nil
}

func postgresAddress(cfg config.PostgresConfig) string {
	address := fmt.Sprintf("postgres://%s:%s@%s/%s", cfg.UserName, cfg.Password, cfg.Host, cfg.Database)
	if cfg.SSLMode != "" {
		address += "?sslmode=" + cfg.SSLMode
	}
	return address
}
