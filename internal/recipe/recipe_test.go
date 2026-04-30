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
database:
  driver: postgres
  framework: pgx
  migrations: goose
logging:
  provider: slog
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
`)

	r, err := Load(path)

	s.Require().NoError(err)
	s.Require().Equal("orders-web", r.Project.Name)
	s.Require().Equal(ServerFrameworkChi, r.Server.Framework)
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
	s.Require().Equal(DatabaseDriverNone, r.Database.Driver)
	s.Require().Equal(DatabaseFrameworkNone, r.Database.Framework)
	s.Require().Equal(DatabaseMigrationsNone, r.Database.Migrations)
	s.Require().Equal(LoggingProviderSlog, r.Logging.Provider)
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

func (s *RecipeTestSuite) TestDatabasePresetsAreValid() {
	for _, name := range []string{PresetWebPostgres, PresetWebMySQL, PresetWebSQLite} {
		s.Run(name, func() {
			r, err := NewPreset(name)

			s.Require().NoError(err)
			s.Require().NoError(Validate(r))
			s.Require().Equal(ProjectTypeWeb, r.Project.Type)
			s.Require().Equal(DatabaseMigrationsMigrate, r.Database.Migrations)
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
	s.Require().Contains(output, "request_logging:")
	s.Require().Contains(output, "github_actions:")
	s.Require().Contains(output, "version: v1\n\nproject:")
	s.Require().Contains(output, "\nproject:\n  name:")
	s.Require().Contains(output, "\nproject:\n  name: orders-web\n  module:")
	s.Require().NotContains(output, "    name:")
	s.Require().NotContains(output, "gracefulShutdown")
	s.Require().NotContains(output, "requestLogging")
	s.Require().NotContains(output, "githubActions")
}

func (s *RecipeTestSuite) writeRecipe(contents string) string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	err := os.WriteFile(path, []byte(strings.TrimLeft(contents, "\n")), 0o644)
	s.Require().NoError(err)
	return path
}
