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
				"internal/app/server.go",
				"internal/app/health.go",
				"internal/app/ready.go",
			},
			absentFiles: []string{
				"configs/config.yaml",
				"internal/config/config.go",
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
				"internal/server/handler/health.go",
				"internal/server/handler/ready.go",
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
				"internal/app/server.go",
			},
			absentFiles: []string{
				"configs/config.yaml",
				"configs/config.json",
				"configs/config.toml",
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
