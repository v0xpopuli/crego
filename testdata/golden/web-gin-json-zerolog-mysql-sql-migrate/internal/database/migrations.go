package database

import (
	"context"

	"errors"

	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/example/orders-api/internal/logging"
)

func (c *MySQLClient) RunMigrations(ctx context.Context) error {
	return runSQLMigrations(ctx, "mysql", mysqlMigrationURL(c.config), c.config.Migrations, c.logger)
}

func (c *MySQLClient) RollbackMigration(ctx context.Context) error {
	return rollbackSQLMigration(ctx, "mysql", mysqlMigrationURL(c.config), c.config.Migrations, c.logger)
}

func (c *MySQLClient) MigrationStatus(ctx context.Context) error {
	return sqlMigrationStatus(ctx, "mysql", mysqlMigrationURL(c.config), c.config.Migrations, c.logger)
}

func runSQLMigrations(ctx context.Context, driver string, address string, source string, logger logging.Logger) error {
	_ = ctx
	logInfo(logger, "running database migrations", "driver", driver, "source", source)
	runner, err := newMigrateRunner(driver, address, source)
	if err != nil {
		return err
	}
	defer runner.Close()
	if err := runner.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logInfo(logger, "database migrations already current", "driver", driver)
			return nil
		}
		return fmt.Errorf("run migrate migrations: %w", err)
	}
	logInfo(logger, "database migrations applied", "driver", driver)
	return nil
}

func rollbackSQLMigration(ctx context.Context, driver string, address string, source string, logger logging.Logger) error {
	_ = ctx
	logInfo(logger, "rolling back database migration", "driver", driver, "source", source)
	runner, err := newMigrateRunner(driver, address, source)
	if err != nil {
		return err
	}
	defer runner.Close()
	if err := runner.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logInfo(logger, "no database migration to roll back", "driver", driver)
			return nil
		}
		return fmt.Errorf("rollback migrate migration: %w", err)
	}
	logInfo(logger, "database migration rolled back", "driver", driver)
	return nil
}

func sqlMigrationStatus(ctx context.Context, driver string, address string, source string, logger logging.Logger) error {
	_ = ctx
	logInfo(logger, "checking database migration status", "driver", driver, "source", source)
	runner, err := newMigrateRunner(driver, address, source)
	if err != nil {
		return err
	}
	defer runner.Close()
	version, dirty, err := runner.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			logInfo(logger, "database migrations not applied", "driver", driver)
			return nil
		}
		return fmt.Errorf("read migration status: %w", err)
	}
	logInfo(logger, "database migration status", "driver", driver, "version", version, "dirty", dirty)
	return nil
}

func newMigrateRunner(driver string, address string, source string) (*migrate.Migrate, error) {
	runner, err := migrate.New(migrationSource(source), migrateDatabaseURL(driver, address))
	if err != nil {
		return nil, fmt.Errorf("create migrate runner: %w", err)
	}
	return runner, nil
}

func migrateDatabaseURL(driver string, address string) string {
	if driver == "sqlite" && strings.HasPrefix(address, "file:") {
		return "sqlite://" + strings.TrimPrefix(address, "file:")
	}
	return address
}

func migrationSource(source string) string {
	if source == "" {
		return "file://scripts/migrations"
	}
	return source
}

func migrationDirectory(source string) string {
	source = migrationSource(source)
	return strings.TrimPrefix(source, "file://")
}

func logInfo(logger logging.Logger, message string, args ...any) {
	if logger != nil {
		logger.Info(message, args...)
	}
}
