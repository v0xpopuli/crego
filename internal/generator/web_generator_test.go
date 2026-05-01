package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
)

type WebGeneratorTestSuite struct {
	suite.Suite
}

func TestWebGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(WebGeneratorTestSuite))
}

func (s *WebGeneratorTestSuite) TestGeneratesWebServiceMatrix() {
	cases := []struct {
		name           string
		layout         string
		server         string
		configFormat   string
		logging        string
		expectedFiles  []string
		absentFiles    []string
		expectedModule []string
		absentModule   []string
	}{
		{
			name:         "nethttp env slog",
			layout:       recipe.LayoutStyleMinimal,
			server:       recipe.ServerFrameworkNetHTTP,
			configFormat: recipe.ConfigurationFormatEnv,
			logging:      recipe.LoggingFrameworkSlog,
			expectedFiles: []string{
				"cmd/orders-api/main.go",
				"internal/app/app.go",
				"internal/app/config.go",
				"internal/app/logger.go",
				"internal/app/server.go",
				"internal/app/recover.go",
				"internal/app/readiness.go",
				"internal/app/health.go",
				"internal/app/ready.go",
			},
			absentFiles: []string{
				"configs/config.yaml",
				"internal/config/config.go",
				"internal/logging/logger.go",
				"internal/server/server.go",
			},
			absentModule: []string{
				"github.com/go-chi/chi/v5",
				"github.com/gin-gonic/gin",
				"github.com/labstack/echo/v4",
				"github.com/gofiber/fiber/v2",
				"go.uber.org/zap",
				"github.com/rs/zerolog",
				"github.com/sirupsen/logrus",
				"gopkg.in/yaml.v3",
				"github.com/pelletier/go-toml/v2",
			},
		},
		{
			name:         "chi yaml zap",
			layout:       recipe.LayoutStyleLayered,
			server:       recipe.ServerFrameworkChi,
			configFormat: recipe.ConfigurationFormatYAML,
			logging:      recipe.LoggingFrameworkZap,
			expectedFiles: []string{
				"cmd/orders-api/main.go",
				"configs/config.yaml",
				"internal/app/app.go",
				"internal/config/config.go",
				"internal/logging/logger.go",
				"internal/server/server.go",
				"internal/server/middleware/recover.go",
				"internal/server/readiness.go",
				"internal/server/handler/health.go",
				"internal/server/ready.go",
			},
			absentFiles: []string{
				"configs/config.json",
				"configs/config.toml",
				"internal/app/server.go",
			},
			expectedModule: []string{
				"github.com/go-chi/chi/v5",
				"go.uber.org/zap",
				"gopkg.in/yaml.v3",
			},
			absentModule: []string{
				"github.com/gin-gonic/gin",
				"github.com/labstack/echo/v4",
				"github.com/gofiber/fiber/v2",
				"github.com/rs/zerolog",
				"github.com/sirupsen/logrus",
				"github.com/pelletier/go-toml/v2",
			},
		},
		{
			name:         "gin json zerolog",
			layout:       recipe.LayoutStyleLayered,
			server:       recipe.ServerFrameworkGin,
			configFormat: recipe.ConfigurationFormatJSON,
			logging:      recipe.LoggingFrameworkZerolog,
			expectedFiles: []string{
				"cmd/orders-api/main.go",
				"configs/config.json",
				"internal/app/app.go",
				"internal/config/config.go",
				"internal/logging/logger.go",
				"internal/server/server.go",
				"internal/server/middleware/recover.go",
				"internal/server/readiness.go",
			},
			absentFiles: []string{
				"configs/config.yaml",
				"configs/config.toml",
				"internal/app/server.go",
			},
			expectedModule: []string{
				"github.com/gin-gonic/gin",
				"github.com/rs/zerolog",
			},
			absentModule: []string{
				"github.com/go-chi/chi/v5",
				"github.com/labstack/echo/v4",
				"github.com/gofiber/fiber/v2",
				"go.uber.org/zap",
				"github.com/sirupsen/logrus",
				"gopkg.in/yaml.v3",
				"github.com/pelletier/go-toml/v2",
			},
		},
		{
			name:         "echo toml logrus",
			layout:       recipe.LayoutStyleLayered,
			server:       recipe.ServerFrameworkEcho,
			configFormat: recipe.ConfigurationFormatTOML,
			logging:      recipe.LoggingFrameworkLogrus,
			expectedFiles: []string{
				"cmd/orders-api/main.go",
				"configs/config.toml",
				"internal/app/app.go",
				"internal/config/config.go",
				"internal/logging/logger.go",
				"internal/server/server.go",
				"internal/server/middleware/recover.go",
				"internal/server/readiness.go",
			},
			absentFiles: []string{
				"configs/config.yaml",
				"configs/config.json",
				"internal/app/server.go",
			},
			expectedModule: []string{
				"github.com/labstack/echo/v4",
				"github.com/sirupsen/logrus",
				"github.com/pelletier/go-toml/v2",
			},
			absentModule: []string{
				"github.com/go-chi/chi/v5",
				"github.com/gin-gonic/gin",
				"github.com/gofiber/fiber/v2",
				"go.uber.org/zap",
				"github.com/rs/zerolog",
				"gopkg.in/yaml.v3",
			},
		},
		{
			name:         "fiber env slog",
			layout:       recipe.LayoutStyleMinimal,
			server:       recipe.ServerFrameworkFiber,
			configFormat: recipe.ConfigurationFormatEnv,
			logging:      recipe.LoggingFrameworkSlog,
			expectedFiles: []string{
				"cmd/orders-api/main.go",
				"internal/app/app.go",
				"internal/app/config.go",
				"internal/app/logger.go",
				"internal/app/server.go",
				"internal/app/recover.go",
				"internal/app/readiness.go",
			},
			absentFiles: []string{
				"configs/config.yaml",
				"configs/config.json",
				"configs/config.toml",
				"internal/logging/logger.go",
				"internal/server/server.go",
			},
			expectedModule: []string{
				"github.com/gofiber/fiber/v2",
			},
			absentModule: []string{
				"github.com/go-chi/chi/v5",
				"github.com/gin-gonic/gin",
				"github.com/labstack/echo/v4",
				"go.uber.org/zap",
				"github.com/rs/zerolog",
				"github.com/sirupsen/logrus",
				"gopkg.in/yaml.v3",
				"github.com/pelletier/go-toml/v2",
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			r := webRecipe(tc.layout, tc.server, tc.configFormat, tc.logging)
			plan, err := Resolve(component.NewRegistry(), r)
			s.Require().NoError(err)

			outDir := s.T().TempDir()
			result, err := NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})
			s.Require().NoError(err)
			s.Require().Contains(result.FilesWritten, "cmd/orders-api/main.go")

			for _, expected := range tc.expectedFiles {
				s.Require().FileExists(filepath.Join(outDir, expected))
			}
			for _, absent := range tc.absentFiles {
				s.requireNoGeneratedFile(outDir, absent)
			}

			goMod := s.readGenerated(outDir, "go.mod")
			for _, expected := range tc.expectedModule {
				s.Require().Contains(goMod, expected)
			}
			for _, absent := range tc.absentModule {
				s.Require().NotContains(goMod, absent)
			}

			readme := s.readGenerated(outDir, "README.md")
			s.Require().Contains(readme, "Server: `"+tc.server+"`")
			s.Require().Contains(readme, "Configuration: `"+tc.configFormat+"`")
			s.Require().Contains(readme, "Logging: `"+tc.logging+"`")

			makefile := s.readGenerated(outDir, "Makefile")
			s.Require().Contains(makefile, "MAIN_PACKAGE ?= ./cmd/orders-api")

			mainGo := s.readGenerated(outDir, "cmd/orders-api/main.go")
			s.Require().Contains(mainGo, "application, err := app.New(ctx, cfg)")
			s.Require().NotContains(mainGo, "NewServer")
			s.Require().NotContains(mainGo, "Shutdown(")

			appGo := s.readGenerated(outDir, "internal/app/app.go")
			s.Require().Contains(appGo, "type Application struct")
			s.Require().Contains(appGo, "func (a *Application) Run() error")
		})
	}
}

func (s *WebGeneratorTestSuite) TestGeneratesDatabaseMatrix() {
	cases := []struct {
		name             string
		driver           string
		framework        string
		migrations       string
		expectedFiles    []string
		absentFiles      []string
		expectedModule   []string
		absentModule     []string
		expectedConfig   []string
		absentConfig     []string
		expectedReadme   []string
		expectedMakefile []string
		absentMakefile   []string
	}{
		{
			name:       "postgres pgx goose",
			driver:     recipe.DatabaseDriverPostgres,
			framework:  recipe.DatabaseFrameworkPGX,
			migrations: recipe.DatabaseMigrationsGoose,
			expectedFiles: []string{
				"internal/database/postgres.go",
				"internal/database/migrations.go",
				"scripts/migrations/000001_init.sql",
			},
			absentFiles: []string{
				"internal/database/redis.go",
				"cmd/orders-api-migrate/main.go",
				"scripts/migrations/000001_init.up.sql",
			},
			expectedModule: []string{
				"github.com/jackc/pgx/v5",
				"github.com/pressly/goose/v3",
			},
			absentModule: []string{
				"github.com/redis/go-redis/v9",
				"go.mongodb.org/mongo-driver/v2",
				"github.com/golang-migrate/migrate/v4",
			},
			expectedConfig: []string{
				"PostgresURL",
				"DatabaseMigrations string",
			},
			expectedReadme: []string{
				"Database: `postgres`",
				"SQL framework: `pgx`",
				"SQL migrations: `goose`",
				"automatically during application startup",
			},
			absentMakefile: []string{
				"migrate-up:",
				"MIGRATE_PACKAGE",
			},
		},
		{
			name:       "mysql gorm migrate",
			driver:     recipe.DatabaseDriverMySQL,
			framework:  recipe.DatabaseFrameworkGORM,
			migrations: recipe.DatabaseMigrationsMigrate,
			expectedFiles: []string{
				"internal/database/mysql.go",
				"internal/database/migrations.go",
				"scripts/migrations/000001_init.up.sql",
				"scripts/migrations/000001_init.down.sql",
			},
			expectedModule: []string{
				"gorm.io/gorm",
				"gorm.io/driver/mysql",
				"github.com/golang-migrate/migrate/v4",
			},
			absentModule: []string{
				"github.com/go-sql-driver/mysql",
				"github.com/pressly/goose/v3",
			},
			expectedConfig: []string{
				"mysql://root:root@tcp(localhost:3306)/app?parseTime=true",
				"MySQLURL",
				"DatabaseMigrations string",
			},
			absentMakefile: []string{"migrate-status:", "MIGRATE_PACKAGE"},
		},
		{
			name:       "redis",
			driver:     recipe.DatabaseDriverRedis,
			framework:  recipe.DatabaseFrameworkNone,
			migrations: recipe.DatabaseMigrationsNone,
			expectedFiles: []string{
				"internal/database/redis.go",
			},
			absentFiles: []string{
				"internal/database/migrations.go",
				"cmd/orders-api-migrate/main.go",
				"scripts/migrations/000001_init.sql",
			},
			expectedModule: []string{"github.com/redis/go-redis/v9"},
			absentModule: []string{
				"gorm.io/gorm",
				"github.com/pressly/goose/v3",
				"github.com/golang-migrate/migrate/v4",
				"go.mongodb.org/mongo-driver/v2",
			},
			expectedConfig: []string{"RedisAddress"},
			absentConfig:   []string{"DatabaseMigrations", "PostgresURL", "MySQLURL", "SQLiteURL"},
			expectedReadme: []string{"Database: `redis`", "No SQL framework or migration settings are generated."},
			absentMakefile: []string{"migrate-up:", "MIGRATE_PACKAGE"},
		},
		{
			name:       "mongodb",
			driver:     recipe.DatabaseDriverMongoDB,
			framework:  recipe.DatabaseFrameworkNone,
			migrations: recipe.DatabaseMigrationsNone,
			expectedFiles: []string{
				"internal/database/mongodb.go",
			},
			absentFiles: []string{
				"internal/database/migrations.go",
				"cmd/orders-api-migrate/main.go",
				"scripts/migrations/000001_init.sql",
			},
			expectedModule: []string{"go.mongodb.org/mongo-driver/v2"},
			absentModule: []string{
				"gorm.io/gorm",
				"github.com/pressly/goose/v3",
				"github.com/golang-migrate/migrate/v4",
				"github.com/redis/go-redis/v9",
			},
			expectedConfig: []string{"MongoDBURI"},
			absentConfig:   []string{"DatabaseMigrations", "PostgresURL", "MySQLURL", "SQLiteURL"},
			expectedReadme: []string{"Database: `mongodb`", "No SQL framework or migration settings are generated."},
			absentMakefile: []string{"migrate-up:", "MIGRATE_PACKAGE"},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			r := webRecipe(recipe.LayoutStyleLayered, recipe.ServerFrameworkChi, recipe.ConfigurationFormatYAML, recipe.LoggingFrameworkSlog)
			r.Database.Driver = tc.driver
			r.Database.Framework = tc.framework
			r.Database.Migrations = tc.migrations
			plan, err := Resolve(component.NewRegistry(), r)
			s.Require().NoError(err)

			outDir := s.T().TempDir()
			_, err = NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})
			s.Require().NoError(err)

			for _, expected := range tc.expectedFiles {
				s.Require().FileExists(filepath.Join(outDir, expected))
			}
			for _, absent := range tc.absentFiles {
				s.requireNoGeneratedFile(outDir, absent)
			}

			goMod := s.readGenerated(outDir, "go.mod")
			for _, expected := range tc.expectedModule {
				s.Require().Contains(goMod, expected)
			}
			for _, absent := range tc.absentModule {
				s.Require().NotContains(goMod, absent)
			}

			configGo := s.readGenerated(outDir, "internal/config/config.go")
			for _, expected := range tc.expectedConfig {
				s.Require().Contains(configGo, expected)
			}
			for _, absent := range tc.absentConfig {
				s.Require().NotContains(configGo, absent)
			}

			readyGo := s.readGenerated(outDir, "internal/server/ready.go")
			s.Require().Contains(readyGo, "StatusServiceUnavailable")
			s.Require().Contains(readyGo, "Check")

			routesGo := s.readGenerated(outDir, "internal/server/routes.go")
			s.Require().Contains(routesGo, "router.Use(middleware.Recover(s.logger))")

			readme := s.readGenerated(outDir, "README.md")
			for _, expected := range tc.expectedReadme {
				s.Require().Contains(readme, expected)
			}

			makefile := s.readGenerated(outDir, "Makefile")
			for _, expected := range tc.expectedMakefile {
				s.Require().Contains(makefile, expected)
			}
			for _, absent := range tc.absentMakefile {
				s.Require().NotContains(makefile, absent)
			}
		})
	}
}

func (s *WebGeneratorTestSuite) TestRendersRepresentativeDatabaseCompileFixtures() {
	cases := []struct {
		name       string
		driver     string
		framework  string
		migrations string
	}{
		{"postgres pgx goose", recipe.DatabaseDriverPostgres, recipe.DatabaseFrameworkPGX, recipe.DatabaseMigrationsGoose},
		{"postgres sql migrate", recipe.DatabaseDriverPostgres, recipe.DatabaseFrameworkDatabaseSQL, recipe.DatabaseMigrationsMigrate},
		{"mysql sql goose", recipe.DatabaseDriverMySQL, recipe.DatabaseFrameworkDatabaseSQL, recipe.DatabaseMigrationsGoose},
		{"mysql gorm migrate", recipe.DatabaseDriverMySQL, recipe.DatabaseFrameworkGORM, recipe.DatabaseMigrationsMigrate},
		{"sqlite sql goose", recipe.DatabaseDriverSQLite, recipe.DatabaseFrameworkDatabaseSQL, recipe.DatabaseMigrationsGoose},
		{"sqlite gorm migrate", recipe.DatabaseDriverSQLite, recipe.DatabaseFrameworkGORM, recipe.DatabaseMigrationsMigrate},
		{"redis", recipe.DatabaseDriverRedis, recipe.DatabaseFrameworkNone, recipe.DatabaseMigrationsNone},
		{"mongodb", recipe.DatabaseDriverMongoDB, recipe.DatabaseFrameworkNone, recipe.DatabaseMigrationsNone},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			r := webRecipe(recipe.LayoutStyleLayered, recipe.ServerFrameworkChi, recipe.ConfigurationFormatEnv, recipe.LoggingFrameworkSlog)
			r.Database.Driver = tc.driver
			r.Database.Framework = tc.framework
			r.Database.Migrations = tc.migrations
			plan, err := Resolve(component.NewRegistry(), r)
			s.Require().NoError(err)

			outDir := s.T().TempDir()
			_, err = NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})

			s.Require().NoError(err)
			s.Require().FileExists(filepath.Join(outDir, "cmd/orders-api/main.go"))
			if tc.driver != recipe.DatabaseDriverNone {
				s.Require().DirExists(filepath.Join(outDir, "internal/database"))
			}
		})
	}
}

func (s *WebGeneratorTestSuite) TestGeneratesMultipleDatabaseProject() {
	r := webRecipe(recipe.LayoutStyleLayered, recipe.ServerFrameworkChi, recipe.ConfigurationFormatYAML, recipe.LoggingFrameworkSlog)
	r.Database.Drivers = []string{recipe.DatabaseDriverPostgres, recipe.DatabaseDriverRedis, recipe.DatabaseDriverMongoDB}
	r.Database.Framework = recipe.DatabaseFrameworkDatabaseSQL
	r.Database.Migrations = recipe.DatabaseMigrationsMigrate
	plan, err := Resolve(component.NewRegistry(), r)
	s.Require().NoError(err)

	outDir := s.T().TempDir()
	_, err = NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})
	s.Require().NoError(err)

	s.Require().FileExists(filepath.Join(outDir, "internal/database/postgres.go"))
	s.Require().FileExists(filepath.Join(outDir, "internal/database/redis.go"))
	s.Require().FileExists(filepath.Join(outDir, "internal/database/mongodb.go"))
	s.Require().FileExists(filepath.Join(outDir, "scripts/migrations/000001_init.up.sql"))
	s.requireNoGeneratedFile(outDir, "cmd/orders-api-migrate/main.go")
	configGo := s.readGenerated(outDir, "internal/config/config.go")
	s.Require().Contains(configGo, "PostgresURL")
	s.Require().Contains(configGo, "RedisAddress")
	s.Require().Contains(configGo, "MongoDBURI")
	appGo := s.readGenerated(outDir, "internal/app/app.go")
	s.Require().Contains(appGo, "postgresClient *database.PostgresClient")
	s.Require().Contains(appGo, "redisClient *database.RedisClient")
	s.Require().Contains(appGo, "mongoDBClient *database.MongoDBClient")
	s.Require().Contains(appGo, "readinessChecks(")
	s.Require().Contains(appGo, "a.postgresClient.RunMigrations(a.ctx)")
	s.Require().Contains(appGo, "postgresClient:")
	s.Require().Contains(appGo, "redisClient:")
	s.Require().Contains(appGo, "mongoDBClient:")
	s.Require().NotContains(appGo, "newPostgresClient")
}

func (s *WebGeneratorTestSuite) readGenerated(outDir string, target string) string {
	s.T().Helper()

	data, err := os.ReadFile(filepath.Join(outDir, target))
	s.Require().NoError(err)
	return string(data)
}

func (s *WebGeneratorTestSuite) requireNoGeneratedFile(outDir string, target string) {
	s.T().Helper()

	_, err := os.Stat(filepath.Join(outDir, target))
	s.Require().True(os.IsNotExist(err), "expected %s to be absent", target)
}

func webRecipe(layout string, server string, configFormat string, loggingFramework string) *recipe.Recipe {
	return &recipe.Recipe{
		Version: recipe.VersionV1,
		Project: recipe.ProjectConfig{
			Name:   "orders-api",
			Module: "github.com/acme/orders-api",
			Type:   recipe.ProjectTypeWeb,
		},
		Go: recipe.GoConfig{
			Version: "1.24",
		},
		Layout: recipe.LayoutConfig{
			Style: layout,
		},
		Server: recipe.ServerConfig{
			Framework: server,
			Port:      8080,
		},
		Configuration: recipe.ConfigurationConfig{
			Format: configFormat,
		},
		Logging: recipe.LoggingConfig{
			Framework: loggingFramework,
			Format:    recipe.LoggingFormatJSON,
		},
		Observability: recipe.ObservabilityConfig{
			Health:    true,
			Readiness: true,
		},
	}
}
