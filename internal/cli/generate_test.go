package cli

import (
	"os"
	"path/filepath"
)

func (s *CliTestSuite) TestGenerateCommand() {
	s.Run("dry run prints planned files without writing output", func() {
		recipePath := s.writeGenerateRecipe()
		outDir := s.T().TempDir()
		before := tempDirEntries(outDir)

		out, _, err := s.executeCLI("generate", "--recipe", recipePath, "--out", outDir, "--dry-run")

		s.Require().NoError(err)
		s.Require().Equal(before, tempDirEntries(outDir))
		s.Require().Contains(out, "Generation plan")
		s.Require().Contains(out, "Files:")
		s.Require().Contains(out, "Go modules:")
		s.Require().Contains(out, "Hooks:")
		s.Require().Contains(out, "cmd/orders-api/main.go")
		s.Require().Contains(out, "internal/app/config.go")
	})

	s.Run("uses global config flag as recipe path", func() {
		recipePath := s.writeGenerateRecipe()
		outDir := s.T().TempDir()

		out, _, err := s.executeCLI("generate", "--config", recipePath, "--out", outDir, "--dry-run")

		s.Require().NoError(err)
		s.Require().Contains(out, "cmd/orders-api/main.go")
	})

	s.Run("derives output directory and prints next steps", func() {
		recipePath := s.writeGenerateRecipe()
		workingDir := s.T().TempDir()
		currentDir, err := os.Getwd()
		s.Require().NoError(err)
		s.Require().NoError(os.Chdir(workingDir))
		s.T().Cleanup(func() {
			s.Require().NoError(os.Chdir(currentDir))
		})

		out, _, err := s.executeCLI("generate", "--recipe", recipePath, "--skip-go-mod-tidy", "--skip-git-init")

		s.Require().NoError(err)
		s.Require().Contains(out, "Project generated successfully.")
		s.Require().Contains(out, "cd orders-api")
		s.Require().Contains(out, "make test")
		s.Require().Contains(out, "make run")
		s.Require().FileExists(filepath.Join(workingDir, "orders-api", "go.mod"))
	})
}

func (s *CliTestSuite) writeGenerateRecipe() string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	err := os.WriteFile(path, []byte(`version: v1
project:
  name: orders-api
  module: github.com/example/orders-api
  type: web
layout:
  style: minimal
server:
  framework: nethttp
configuration:
  format: env
logging:
  framework: slog
  format: text
observability:
  health: true
  readiness: true
`), 0o644)
	s.Require().NoError(err)
	return path
}
