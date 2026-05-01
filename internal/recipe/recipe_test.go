package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RecipeTestSuite struct {
	suite.Suite
}

func TestRecipeTestSuite(t *testing.T) {
	suite.Run(t, new(RecipeTestSuite))
}

func (s *RecipeTestSuite) TestLoadFullWebRecipe() {
	path := s.writeRecipe(`version: "v1"
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
go:
  version: "1.24"
layout:
  style: layered
server:
  framework: chi
  port: 8080
  graceful_shutdown: true
configuration:
  format: yaml
database:
  driver: postgres
  framework: pgx
  migrations: goose
logging:
  framework: slog
  format: json
  request_logging: true
observability:
  health: true
  readiness: true
  metrics: false
  tracing: false
deployment:
  docker: true
  compose: true
ci:
  github_actions: true
  gitlab_ci: false
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal("orders-web", r.Project.Name)
	s.Require().Equal(ServerFrameworkChi, r.Server.Framework)
	s.Require().Equal(ConfigurationFormatYAML, r.Configuration.Format)
	s.Require().Equal(DatabaseDriverPostgres, r.Database.Driver)
	s.Require().Equal(LoggingFormatJSON, r.Logging.Format)
}

func (s *RecipeTestSuite) TestLoadMinimalWebRecipeAppliesDefaults() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal("1.24", r.Go.Version)
	s.Require().Equal(LayoutStyleMinimal, r.Layout.Style)
	s.Require().Equal(ServerFrameworkNetHTTP, r.Server.Framework)
	s.Require().Equal(8080, r.Server.Port)
	s.Require().True(r.Server.GracefulShutdown)
	s.Require().Equal(ConfigurationFormatEnv, r.Configuration.Format)
	s.Require().Equal(DatabaseDriverNone, r.Database.Driver)
	s.Require().Equal(DatabaseFrameworkNone, r.Database.Framework)
	s.Require().Equal(DatabaseMigrationsNone, r.Database.Migrations)
	s.Require().Equal(LoggingFrameworkSlog, r.Logging.Framework)
	s.Require().Equal(LoggingFormatText, r.Logging.Format)
}

func (s *RecipeTestSuite) TestLoadInvalidEnumValue() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
server:
  framework: martini
`)

	_, err := Load(path)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "server.framework=martini is invalid")
	s.Require().Contains(err.Error(), "allowed values: nethttp, chi, gin, echo, fiber")
}

func (s *RecipeTestSuite) TestLoadRejectsLoggingProviderField() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
logging:
  provider: slog
`)

	_, err := Load(path)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), `unknown logging field "provider"`)
}

func (s *RecipeTestSuite) TestLoadAcceptsLoggingFrameworks() {
	for _, framework := range []string{LoggingFrameworkSlog, LoggingFrameworkZap, LoggingFrameworkZerolog, LoggingFrameworkLogrus} {
		s.Run(framework, func() {
			path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
logging:
  framework: ` + framework + `
`)

			r, err := Load(path)

			s.Require().NoError(err)
			s.Require().Equal(framework, r.Logging.Framework)
		})
	}
}

func (s *RecipeTestSuite) TestLoadAcceptsDatabaseDrivers() {
	for _, driver := range []string{DatabaseDriverNone, DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite, DatabaseDriverRedis, DatabaseDriverMongoDB} {
		s.Run(driver, func() {
			path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  driver: ` + driver + `
`)

			r, err := Load(path)

			s.Require().NoError(err)
			s.Require().Equal(driver, r.Database.Driver)
		})
	}
}

func (s *RecipeTestSuite) TestLoadAcceptsMultipleDatabaseDrivers() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
sql_database: postgres
orm_framework: sql
nosql_database:
  - redis
  - mongodb
migrations: migrate
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal(DatabaseDriverPostgres, r.Database.Driver)
	s.Require().Equal([]string{DatabaseDriverPostgres, DatabaseDriverRedis, DatabaseDriverMongoDB}, r.Database.Drivers)
	s.Require().Equal(DatabaseFrameworkDatabaseSQL, r.Database.Framework)
}

func (s *RecipeTestSuite) TestLoadAcceptsLegacyDatabaseRoot() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  driver: postgres
  framework: pgx
  migrations: goose
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal(DatabaseDriverPostgres, r.Database.Driver)
	s.Require().Equal(DatabaseFrameworkPGX, r.Database.Framework)
	s.Require().Equal(DatabaseMigrationsGoose, r.Database.Migrations)
}

func (s *RecipeTestSuite) TestLoadCanonicalizesDatabaseSQLFramework() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  driver: postgres
  framework: database/sql
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal(DatabaseFrameworkDatabaseSQL, r.Database.Framework)
}

func (s *RecipeTestSuite) TestLoadMissingModule() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  type: web
`)

	_, err := Load(path)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "project.module is required")
}

func (s *RecipeTestSuite) TestLoadMigrationsWithoutDatabase() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  driver: none
  migrations: goose
`)

	_, err := Load(path)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "database.migrations=goose requires database.driver to be postgres, mysql, or sqlite")
}

func (s *RecipeTestSuite) TestLoadDatabaseFrameworkCompatibility() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  driver: mysql
  framework: pgx
`)

	_, err := Load(path)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "database.framework=pgx is only supported with database.driver=postgres")
}

func (s *RecipeTestSuite) TestLoadRejectsNoSQLFrameworkAndMigrations() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  driver: redis
  framework: gorm
  migrations: migrate
`)

	_, err := Load(path)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "database.framework=gorm is only supported with SQL database drivers")
	s.Require().Contains(err.Error(), "database.migrations=migrate is only supported with SQL database drivers")
}

func (s *RecipeTestSuite) TestLoadAcceptsMultipleSQLDrivers() {
	path := s.writeRecipe(`version: v1
project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web
database:
  drivers:
    - postgres
    - mysql
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal([]string{DatabaseDriverPostgres, DatabaseDriverMySQL}, r.Database.Drivers)
	s.Require().Equal(DatabaseFrameworkDatabaseSQL, r.Database.Framework)
}

func (s *RecipeTestSuite) TestDatabasePresetsAreValid() {
	for _, name := range []string{PresetWebPostgres, PresetWebMySQL, PresetWebSQLite, PresetWebRedis, PresetWebMongoDB} {
		s.Run(name, func() {
			r, err := NewPreset(name)

			s.Require().NoError(err)
			s.Require().NoError(Validate(r))
			s.Require().Equal(ProjectTypeWeb, r.Project.Type)
			if r.Database.Driver == DatabaseDriverRedis || r.Database.Driver == DatabaseDriverMongoDB {
				s.Require().Equal(DatabaseFrameworkNone, r.Database.Framework)
				s.Require().Equal(DatabaseMigrationsNone, r.Database.Migrations)
			} else {
				s.Require().Equal(DatabaseMigrationsMigrate, r.Database.Migrations)
			}
		})
	}
}

func (s *RecipeTestSuite) TestSaveUsesSnakeCaseYAMLKeys() {
	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	r := &Recipe{
		Version: VersionV1,
		Project: ProjectConfig{
			Name:   "orders-web",
			Module: "github.com/acme/orders-web",
			Type:   ProjectTypeWeb,
		},
		Logging: LoggingConfig{
			RequestLogging: true,
		},
		CI: CIConfig{
			GitHubActions: true,
		},
	}

	err := Save(path, r)

	s.Require().NoError(err)
	data, err := os.ReadFile(path)
	s.Require().NoError(err)
	output := string(data)
	s.Require().Contains(output, "graceful_shutdown:")
	s.Require().Contains(output, "framework: slog")
	s.Require().Contains(output, "sql_database: none")
	s.Require().Contains(output, "nosql_database: none")
	s.Require().Contains(output, "request_logging:")
	s.Require().Contains(output, "github_actions:")
	s.Require().Contains(output, "gitlab_ci:")
	s.Require().Contains(output, "version: v1\n\nproject:")
	s.Require().Contains(output, "\nproject:\n  name:")
	s.Require().Contains(output, "\nproject:\n  name: orders-web\n  module:")
	s.Require().NotContains(output, "\ndatabase:")
	s.Require().NotContains(output, "orm_framework: none")
	s.Require().NotContains(output, "    name:")
	s.Require().NotContains(output, "gracefulShutdown")
	s.Require().NotContains(output, "provider:")
	s.Require().NotContains(output, "requestLogging")
	s.Require().NotContains(output, "githubActions")
}

func (s *RecipeTestSuite) TestSaveOmitsNoSQLFrameworkAndMigrations() {
	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	r := &Recipe{
		Version: VersionV1,
		Project: ProjectConfig{
			Name:   "orders-web",
			Module: "github.com/acme/orders-web",
			Type:   ProjectTypeWeb,
		},
		Database: DatabaseConfig{
			Driver: DatabaseDriverRedis,
		},
	}

	err := Save(path, r)

	s.Require().NoError(err)
	data, err := os.ReadFile(path)
	s.Require().NoError(err)
	output := string(data)
	s.Require().Contains(output, "sql_database: none")
	s.Require().Contains(output, "nosql_database: redis")
	s.Require().NotContains(output, "\ndatabase:")
	s.Require().NotContains(output, "orm_framework:")
	s.Require().NotContains(output, "migrations:")
}

func (s *RecipeTestSuite) writeRecipe(contents string) string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	err := os.WriteFile(path, []byte(strings.TrimLeft(contents, "\n")), 0o644)
	s.Require().NoError(err)
	return path
}
