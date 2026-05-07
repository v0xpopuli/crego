package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/example/orders-api/internal/config"
	"github.com/example/orders-api/internal/logging"
)

type MongoDBClient struct {
	config config.MongoDBConfig
	logger logging.Logger
	Client *mongo.Client
}

func NewMongoDBClient(ctx context.Context, cfg config.MongoDBConfig, logger logging.Logger) (*MongoDBClient, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("mongodb_host is required")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(mongoDBURI(cfg)))
	if err != nil {
		return nil, fmt.Errorf("open mongodb: %w", err)
	}
	result := &MongoDBClient{config: cfg, logger: logger, Client: client}
	if err := result.Ping(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}
	if logger != nil {
		logger.Info("database connected", "driver", "mongodb")
	}
	return result, nil
}

func (c *MongoDBClient) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx, readpref.Primary())
}

func (c *MongoDBClient) Shutdown(ctx context.Context) error {
	if err := c.Client.Disconnect(ctx); err != nil {
		return err
	}
	if c.logger != nil {
		c.logger.Info("database connection closed", "driver", "mongodb")
	}
	return nil
}

func mongoDBURI(cfg config.MongoDBConfig) string {
	credentials := ""
	if cfg.UserName != "" {
		credentials = cfg.UserName + ":" + cfg.Password + "@"
	}
	return fmt.Sprintf("mongodb://%s%s/%s", credentials, cfg.Host, cfg.Database)
}
