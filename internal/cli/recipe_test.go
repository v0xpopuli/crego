package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/v0xpopuli/crego/internal/recipe"
)

func (s *CliTestSuite) TestRecipeInitCommand() {
	for _, tc := range []struct {
		preset string
		driver string
	}{
		{preset: recipe.PresetWebPostgres, driver: recipe.DatabaseDriverPostgres},
		{preset: recipe.PresetWebMySQL, driver: recipe.DatabaseDriverMySQL},
		{preset: recipe.PresetWebSQLite, driver: recipe.DatabaseDriverSQLite},
	} {
		s.Run(tc.preset+" writes valid web recipe", func() {
			path := filepath.Join(s.T().TempDir(), "crego.yaml")

			out, _, err := s.executeCLI("recipe", "init", "--preset", tc.preset, "--out", path)

			s.Require().NoError(err)
			s.Require().Contains(out, "created recipe: "+path)

			r, err := recipe.Load(path)
			s.Require().NoError(err)
			s.Require().Equal(recipe.ProjectTypeWeb, r.Project.Type)
			s.Require().Equal(tc.driver, r.Database.Driver)
			s.Require().Equal(recipe.ConfigurationFormatEnv, r.Configuration.Format)
			s.Require().Equal(recipe.DatabaseMigrationsMigrate, r.Database.Migrations)
		})
	}

	for _, tc := range []struct {
		preset string
		driver string
	}{
		{preset: recipe.PresetWebRedis, driver: recipe.DatabaseDriverRedis},
		{preset: recipe.PresetWebMongoDB, driver: recipe.DatabaseDriverMongoDB},
	} {
		s.Run(tc.preset+" writes valid nosql web recipe", func() {
			path := filepath.Join(s.T().TempDir(), "crego.yaml")

			_, _, err := s.executeCLI("recipe", "init", "--preset", tc.preset, "--out", path)

			s.Require().NoError(err)
			r, err := recipe.Load(path)
			s.Require().NoError(err)
			s.Require().Equal(tc.driver, r.Database.Driver)
			s.Require().Equal(recipe.DatabaseFrameworkNone, r.Database.Framework)
			s.Require().Equal(recipe.DatabaseMigrationsNone, r.Database.Migrations)

			data, err := os.ReadFile(path)
			s.Require().NoError(err)
			output := string(data)
			s.Require().Contains(output, "database:")
			s.Require().Contains(output, "  sql: none")
			s.Require().Contains(output, "  nosql: "+tc.driver)
			s.Require().NotContains(output, "orm_framework:")
			s.Require().NotContains(output, "migrations:")
		})
	}

	s.Run("does not overwrite existing file without flag", func() {
		path := filepath.Join(s.T().TempDir(), "crego.yaml")
		s.Require().NoError(os.WriteFile(path, []byte("sentinel"), 0o644))

		_, _, err := s.executeCLI("recipe", "init", "--out", path)

		s.Require().Error(err)
		s.Require().Contains(err.Error(), "already exists")
		data, readErr := os.ReadFile(path)
		s.Require().NoError(readErr)
		s.Require().Equal("sentinel", string(data))
	})

	s.Run("overrides module and derives name", func() {
		path := filepath.Join(s.T().TempDir(), "crego.yaml")

		_, _, err := s.executeCLI(
			"recipe", "init",
			"--module", "github.com/example/orders-api",
			"--out", path,
		)

		s.Require().NoError(err)
		r, err := recipe.Load(path)
		s.Require().NoError(err)
		s.Require().Equal("github.com/example/orders-api", r.Project.Module)
		s.Require().Equal("orders-api", r.Project.Name)
	})

	s.Run("explicit name wins over module basename", func() {
		path := filepath.Join(s.T().TempDir(), "crego.yaml")

		_, _, err := s.executeCLI(
			"recipe", "init",
			"--module", "github.com/example/orders-api",
			"--name", "orders",
			"--out", path,
		)

		s.Require().NoError(err)
		r, err := recipe.Load(path)
		s.Require().NoError(err)
		s.Require().Equal("orders", r.Project.Name)
	})
}

func (s *CliTestSuite) TestRecipeValidateCommand() {
	s.Run("valid recipe succeeds", func() {
		path := s.writeStarterRecipe()

		out, _, err := s.executeCLI("recipe", "validate", path)

		s.Require().NoError(err)
		s.Require().Contains(out, "recipe valid: "+path)
	})

	s.Run("invalid recipe returns exit code 3", func() {
		path := s.writeInvalidRecipe()

		out, _, err := s.executeCLI("recipe", "validate", path)

		s.Require().Error(err)
		s.Require().Equal(3, ExitCode(err))
		s.Require().Contains(out, "recipe invalid: "+path)
		s.Require().Contains(out, "project.module is required")
	})

	s.Run("json output is machine readable", func() {
		path := s.writeStarterRecipe()

		out, _, err := s.executeCLI("recipe", "validate", path, "--json")

		s.Require().NoError(err)
		var result recipeValidateResult
		s.Require().NoError(json.Unmarshal([]byte(out), &result))
		s.Require().True(result.Valid)
		s.Require().Empty(result.Errors)
		s.Require().Empty(result.Warnings)
	})
}

func (s *CliTestSuite) TestRecipePrintCommand() {
	s.Run("prints normalized snake case yaml", func() {
		path := s.writeStarterRecipe()

		out, _, err := s.executeCLI("recipe", "print", path)

		s.Require().NoError(err)
		s.Require().Contains(out, "graceful_shutdown:")
		s.Require().Contains(out, "configuration:")
		s.Require().Contains(out, "format: env")
		s.Require().Contains(out, "request_logging:")
		s.Require().Contains(out, "github_actions:")
		s.Require().Contains(out, "gitlab_ci:")
		s.Require().Contains(out, "azure_pipelines:")
		s.Require().Contains(out, "database:")
		s.Require().Contains(out, "  sql: postgres")
		s.Require().Contains(out, "  orm_framework:")
		s.Require().Contains(out, "  nosql: none")
		s.Require().Contains(out, "  migrations: migrate")
		s.Require().Contains(out, "task_scheduler: none")
		s.Require().Contains(out, "framework: slog")
		s.Require().Contains(out, "version: v1\n\nproject:")
		s.Require().Contains(out, "\nproject:\n  name:")
		s.Require().NotContains(out, "\nsql_database:")
		s.Require().NotContains(out, "    name:")
		s.Require().NotContains(out, "provider:")
		s.Require().NotContains(out, "requestLogging")
		s.Require().NotContains(out, "gitlabCI")
	})

	s.Run("prints valid json", func() {
		path := s.writeStarterRecipe()

		out, _, err := s.executeCLI("recipe", "print", path, "--json")

		s.Require().NoError(err)
		var result map[string]any
		s.Require().NoError(json.Unmarshal([]byte(out), &result))
		s.Require().Contains(result, "project")
		s.Require().NotContains(result, "Project")
	})
}

func (s *CliTestSuite) TestRecipeCommandHelpExamples() {
	cmd := NewRootCommandWithWriters(VersionInfo{}, &bytes.Buffer{}, &bytes.Buffer{})
	for _, args := range [][]string{
		{"recipe", "init"},
		{"recipe", "validate"},
		{"recipe", "print"},
		{"recipe", "edit"},
	} {
		found, _, err := cmd.Find(args)

		s.Require().NoError(err)
		s.Require().NotEmpty(strings.TrimSpace(found.Example))
		s.Require().Contains(found.Example, "crego recipe "+found.Name())
	}
}

func (s *CliTestSuite) TestRecipeEditCommandFlags() {
	cmd := NewRootCommandWithWriters(VersionInfo{}, &bytes.Buffer{}, &bytes.Buffer{})

	found, _, err := cmd.Find([]string{"recipe", "edit"})

	s.Require().NoError(err)
	s.Require().NotNil(found.Flag("save-as"))
	s.Require().NotNil(found.Flag("readonly"))
	s.Require().Contains(found.Example, "crego recipe edit crego.yaml --save-as company-web.yaml")
	s.Require().Contains(found.Example, "crego recipe edit crego.yaml --readonly")
}

func (s *CliTestSuite) TestRecipeCommandRequiresSubcommand() {
	out, _, err := s.executeCLI("recipe")

	s.Require().Error(err)
	s.Require().EqualError(err, "recipe requires a subcommand: init, validate, print, or edit")
	s.Require().Contains(out, "Work with crego recipes")
	s.Require().Contains(out, "init")
	s.Require().Contains(out, "validate")
	s.Require().Contains(out, "print")
	s.Require().Contains(out, "edit")
}

func (s *CliTestSuite) executeCLI(args ...string) (string, string, error) {
	s.T().Helper()

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommandWithWriters(VersionInfo{}, &out, &errOut)
	cmd.SetArgs(args)

	err := cmd.Execute()

	return out.String(), errOut.String(), err
}

func (s *CliTestSuite) writeStarterRecipe() string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	_, _, err := s.executeCLI(
		"recipe", "init",
		"--preset", recipe.PresetWebPostgres,
		"--out", path,
	)
	s.Require().NoError(err)
	return path
}

func (s *CliTestSuite) writeInvalidRecipe() string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	err := os.WriteFile(path, []byte(`version: v1
project:
  name: orders
  type: web
`), 0o644)
	s.Require().NoError(err)
	return path
}
