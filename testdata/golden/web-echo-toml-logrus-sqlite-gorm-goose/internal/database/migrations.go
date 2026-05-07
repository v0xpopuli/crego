package database

import (
	"context"

	"database/sql"

	"fmt"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/pressly/goose/v3"

	"github.com/example/orders-api/internal/logging"
)

func (c *SQLiteClient) RunMigrations(ctx context.Context) error {
	return runSQLMigrations(ctx, "sqlite", c.address, c.config.Migrations, c.logger)
}

func (c *SQLiteClient) RollbackMigration(ctx context.Context) error {
	return rollbackSQLMigration(ctx, "sqlite", c.address, c.config.Migrations, c.logger)
}

func (c *SQLiteClient) MigrationStatus(ctx context.Context) error {
	return sqlMigrationStatus(ctx, "sqlite", c.address, c.config.Migrations, c.logger)
}

func runSQLMigrations(ctx context.Context, driver string, address string, source string, logger logging.Logger) error {
	logInfo(logger, "running database migrations", "driver", driver, "source", source)
	db, err := openMigrationDB(driver, address)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := goose.SetDialect(gooseDialect(driver)); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, migrationDirectory(source)); err != nil {
		return fmt.Errorf("run goose migrations: %w", err)
	}
	logInfo(logger, "database migrations applied", "driver", driver)
	return nil
}

func rollbackSQLMigration(ctx context.Context, driver string, address string, source string, logger logging.Logger) error {
	logInfo(logger, "rolling back database migration", "driver", driver, "source", source)
	db, err := openMigrationDB(driver, address)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := goose.SetDialect(gooseDialect(driver)); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.DownContext(ctx, db, migrationDirectory(source)); err != nil {
		return fmt.Errorf("rollback goose migration: %w", err)
	}
	logInfo(logger, "database migration rolled back", "driver", driver)
	return nil
}

func sqlMigrationStatus(ctx context.Context, driver string, address string, source string, logger logging.Logger) error {
	logInfo(logger, "checking database migration status", "driver", driver, "source", source)
	db, err := openMigrationDB(driver, address)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := goose.SetDialect(gooseDialect(driver)); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	return goose.StatusContext(ctx, db, migrationDirectory(source))
}

func openMigrationDB(driver string, address string) (*sql.DB, error) {
	db, err := sql.Open(gooseDriverName(driver), address)
	if err != nil {
		return nil, fmt.Errorf("open migration database: %w", err)
	}
	return db, nil
}

func gooseDriverName(driver string) string {
	if driver == "postgres" {
		return "pgx"
	}
	if driver == "sqlite" {
		return "sqlite"
	}
	return driver
}

func gooseDialect(driver string) string {
	if driver == "sqlite" {
		return "sqlite3"
	}
	return driver
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
