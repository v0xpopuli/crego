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
		s.Require().Contains(out, "planned files:")
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
}

func (s *CliTestSuite) writeGenerateRecipe() string {
	s.T().Helper()

	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	err := os.WriteFile(path, []byte(`version: v1
project:
  name: orders-api
  module: github.com/acme/orders-api
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
