package cli

import (
	"io"
	"os"
	"path/filepath"
)

func (s *CliTestSuite) TestNewCommand() {
	s.Run("non-interactive flags generate full web matrix", func() {
		outDir := filepath.Join(s.T().TempDir(), "orders-web")

		out, _, err := s.executeCLI(
			"new", "github.com/example/orders-web",
			"--type", "web",
			"--layout", "layered",
			"--server", "chi",
			"--configuration", "yaml",
			"--logging", "zap",
			"--database", "postgres",
			"--framework", "pgx",
			"--migrations", "goose",
			"--docker",
			"--compose",
			"--github-actions",
			"--gitlab-ci",
			"--health",
			"--readiness",
			"--out", outDir,
			"--skip-go-mod-tidy",
			"--skip-git-init",
			"--non-interactive",
		)

		s.Require().NoError(err)
		s.Require().Contains(out, "Project generated successfully.")
		s.Require().Contains(out, "cd "+outDir)
		s.Require().Contains(out, "make test")
		s.Require().Contains(out, "make run")
		s.Require().FileExists(filepath.Join(outDir, "go.mod"))
		s.Require().FileExists(filepath.Join(outDir, "configs", "config.yaml"))
		s.Require().FileExists(filepath.Join(outDir, ".github", "workflows", "test.yml"))
		s.Require().FileExists(filepath.Join(outDir, ".gitlab-ci.yml"))
		s.Require().FileExists(filepath.Join(outDir, "deployments", "docker-compose.yml"))
	})

	s.Run("dry run writes no files", func() {
		outDir := s.T().TempDir()
		before := tempDirEntries(outDir)

		out, _, err := s.executeCLI(
			"new", "github.com/example/orders-web",
			"--server", "gin",
			"--configuration", "json",
			"--logging", "zerolog",
			"--out", outDir,
			"--dry-run",
			"--non-interactive",
		)

		s.Require().NoError(err)
		s.Require().Equal(before, tempDirEntries(outDir))
		s.Require().Contains(out, "Generation plan")
		s.Require().Contains(out, "github.com/gin-gonic/gin")
		s.Require().Contains(out, "github.com/rs/zerolog")
	})

	s.Run("recipe flag generates from recipe without non-interactive flag", func() {
		recipePath := s.writeGenerateRecipe()
		outDir := filepath.Join(s.T().TempDir(), "from-recipe")

		out, _, err := s.executeCLI("new", "--recipe", recipePath, "--out", outDir, "--skip-go-mod-tidy", "--skip-git-init", "--non-interactive")

		s.Require().NoError(err)
		s.Require().Contains(out, "Project generated successfully.")
		s.Require().FileExists(filepath.Join(outDir, "go.mod"))
	})

	s.Run("missing module in non-interactive mode returns clear error", func() {
		_, _, err := s.executeCLI("new", "--non-interactive")

		s.Require().EqualError(err, "module path is required for non-interactive new")
	})

	s.Run("has interactive starter flags", func() {
		cmd := newNewCommand(io.Discard, &globalOptions{})

		s.Require().NotNil(cmd.Flags().Lookup("preset"))
		s.Require().NotNil(cmd.Flags().Lookup("recipe"))
		s.Require().NotNil(cmd.Flags().Lookup("overwrite"))
		s.Require().Contains(cmd.Long, "opens the TUI wizard")
	})

	s.Run("preset builds interactive base recipe", func() {
		r, err := newBaseRecipe(&newOptions{preset: "web-postgres"})

		s.Require().NoError(err)
		s.Require().Equal("postgres", r.Database.Driver)
	})

	s.Run("invalid database framework combination returns validation error", func() {
		_, _, err := s.executeCLI(
			"new", "github.com/example/orders-web",
			"--database", "mysql",
			"--framework", "pgx",
			"--non-interactive",
		)

		s.Require().Error(err)
		s.Require().Contains(err.Error(), "database.framework=pgx is only supported with database.driver=postgres")
	})

	s.Run("derives output directory from module basename", func() {
		workingDir := s.T().TempDir()
		currentDir, err := os.Getwd()
		s.Require().NoError(err)
		s.Require().NoError(os.Chdir(workingDir))
		s.T().Cleanup(func() {
			s.Require().NoError(os.Chdir(currentDir))
		})

		out, _, err := s.executeCLI("new", "github.com/example/orders-cli", "--type", "cli", "--skip-go-mod-tidy", "--skip-git-init", "--non-interactive")

		s.Require().NoError(err)
		s.Require().Contains(out, "cd orders-cli")
		s.Require().FileExists(filepath.Join(workingDir, "orders-cli", "README.md"))
	})
}
