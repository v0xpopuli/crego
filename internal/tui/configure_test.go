package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/recipe"
	"github.com/v0xpopuli/crego/internal/tui/components"
)

type ConfigureWizardTestSuite struct {
	suite.Suite
}

func TestConfigureWizardTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigureWizardTestSuite))
}

func (s *ConfigureWizardTestSuite) TestRecipeBuildsFullWebMatrix() {
	state := NewConfigureWizardState(nil, ConfigureWizardOptions{RecipePath: "custom.yaml"})
	state.Name = "orders"
	state.Module = "github.com/example/orders"
	state.GoVersion = "1.26"
	state.ProjectType = recipe.ProjectTypeWeb
	state.Layout = recipe.LayoutStyleLayered
	state.Server = recipe.ServerFrameworkFiber
	state.ConfigurationFormat = recipe.ConfigurationFormatTOML
	state.Logging = recipe.LoggingFrameworkLogrus
	state.Database = recipe.DatabaseDriverPostgres
	state.DatabaseFramework = recipe.DatabaseFrameworkPGX
	state.Migrations = recipe.DatabaseMigrationsGoose
	state.Health = true
	state.Readiness = true
	state.TaskScheduler = recipe.TaskSchedulerGocron
	state.Docker = true
	state.Compose = true
	state.GitHubActions = true
	state.GitLabCI = true
	state.AzurePipelines = true

	r := state.Recipe()

	s.Require().NoError(recipe.Validate(r))
	s.Require().Equal("custom.yaml", state.RecipePath())
	s.Require().Equal("1.26", r.Go.Version)
	s.Require().Equal(recipe.ServerFrameworkFiber, r.Server.Framework)
	s.Require().Equal(recipe.ConfigurationFormatTOML, r.Configuration.Format)
	s.Require().Equal(recipe.LoggingFrameworkLogrus, r.Logging.Framework)
	s.Require().Equal(recipe.TaskSchedulerGocron, r.TaskScheduler)
	s.Require().True(r.CI.GitHubActions)
	s.Require().True(r.CI.GitLabCI)
	s.Require().True(r.CI.AzurePipelines)
}

func (s *ConfigureWizardTestSuite) TestGenerationModePrefillsModuleAndOutputPreview() {
	source, err := recipe.NewPreset(recipe.PresetWebPostgres)
	s.Require().NoError(err)
	source.Project.Module = "github.com/acme/orders-web"
	source.Project.Name = "orders-web"
	source.TaskScheduler = recipe.TaskSchedulerGocron

	state := NewConfigureWizardState(source, ConfigureWizardOptions{
		OutputDir:     "custom-orders",
		SkipGoModTidy: false,
		Mode:          ConfigureWizardModeGeneration,
	})
	screen := newConfigureScreen(NewStyles(nil, true), state, stepPreview)

	view := screen.preview()

	s.Require().Contains(view, "module: github.com/acme/orders-web")
	s.Require().Contains(view, "name: orders-web")
	s.Require().Contains(view, "output directory: custom-orders")
	s.Require().Contains(view, "task_scheduler: gocron")
	s.Require().Contains(view, "Components:")
	s.Require().Contains(view, "Actions:")
	s.Require().Contains(view, "write files")
	s.Require().Contains(view, "run go mod tidy")
	s.Require().NotContains(view, "Normalized recipe:")
	s.Require().NotContains(view, "Files:")

	files := screen.livePreview()

	s.Require().Contains(files, "Files tree")
	s.Require().Contains(files, "cmd/orders-web/main.go")
	s.Require().Contains(files, "internal/app/app.go")
}

func (s *ConfigureWizardTestSuite) TestGenerationPreviewActions() {
	state := NewConfigureWizardState(nil, ConfigureWizardOptions{Mode: ConfigureWizardModeGeneration})
	screen := newConfigureScreen(NewStyles(nil, true), state, stepPreview)

	s.Require().Equal([]components.SelectOption{
		{Label: "Generate project", Value: "generate"},
		{Label: "Save recipe only", Value: "save"},
		{Label: "Back", Value: "back"},
		{Label: "Cancel", Value: "cancel"},
	}, screen.selectInput.Options)
}

func (s *ConfigureWizardTestSuite) TestGenerationPreviewConfirmsNonEmptyOutputDirectory() {
	outDir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(outDir, "existing.txt"), []byte("existing"), 0o644))

	state := NewConfigureWizardState(nil, ConfigureWizardOptions{
		OutputDir: outDir,
		Mode:      ConfigureWizardModeGeneration,
	})
	screen := newConfigureScreen(NewStyles(nil, true), state, stepPreview)

	s.Require().Equal("Generate project and overwrite non-empty output directory", screen.selectInput.Options[0].Label)
	s.Require().Equal("generate_overwrite", screen.selectInput.Options[0].Value)
}

func (s *ConfigureWizardTestSuite) TestLayoutViewShowsTreeExamplesBelowOptions() {
	screen := newConfigureScreen(NewStyles(nil, true), NewConfigureWizardState(nil, ConfigureWizardOptions{}), stepLayout)

	view := screen.View()

	s.Require().Contains(view, "Minimal")
	s.Require().Contains(view, "Layered")
	s.Require().Contains(view, "Preview")
	s.Require().Contains(view, "cmd/app/main.go")
	s.Require().Contains(view, "internal/app/app.go")
	s.Require().Less(strings.Index(view, "Minimal"), strings.Index(view, "cmd/app/main.go"))
}

func (s *ConfigureWizardTestSuite) TestGoVersionOptionsAreDescending() {
	screen := newConfigureScreen(NewStyles(nil, true), NewConfigureWizardState(nil, ConfigureWizardOptions{}), stepGoVersion)

	s.Require().Equal([]components.SelectOption{
		{Label: "Go 1.26", Value: "1.26", Description: "Latest toolchain target"},
		{Label: "Go 1.25", Value: "1.25", Description: "Stable modern runtime"},
		{Label: "Go 1.24", Value: "1.24", Description: "Conservative compatibility"},
	}, screen.selectInput.Options)
}

func (s *ConfigureWizardTestSuite) TestConfigurationFormatUsesENVLabel() {
	screen := newConfigureScreen(NewStyles(nil, true), NewConfigureWizardState(nil, ConfigureWizardOptions{}), stepConfiguration)

	s.Require().Equal("ENV", screen.selectInput.Options[0].Label)
	s.Require().Equal(recipe.ConfigurationFormatEnv, screen.selectInput.Options[0].Value)
}

func (s *ConfigureWizardTestSuite) TestBackShortcutReturnsToPreviousStep() {
	screen := newConfigureScreen(NewStyles(nil, true), NewConfigureWizardState(nil, ConfigureWizardOptions{}), stepConfiguration)

	next, _ := screen.Update(tea.KeyMsg{Type: tea.KeyCtrlB})

	s.Require().Equal(stepServer, next.(configureScreen).step)
}

func (s *ConfigureWizardTestSuite) TestMultiSelectScreensShowSelectionHint() {
	for _, step := range []configureStep{stepObservability, stepDeployment} {
		s.Run(fmt.Sprintf("step %d", step), func() {
			screen := newConfigureScreen(NewStyles(nil, true), NewConfigureWizardState(nil, ConfigureWizardOptions{}), step)

			s.Require().Contains(screen.View(), "space toggle")
			s.Require().Contains(screen.View(), "esc back")
		})
	}
}

func (s *ConfigureWizardTestSuite) TestSchedulerRecipeYAMLUsesRecipeLevelSelectionOnly() {
	state := NewConfigureWizardState(nil, ConfigureWizardOptions{})
	state.Name = "orders"
	state.Module = "github.com/example/orders"
	state.TaskScheduler = recipe.TaskSchedulerGocron

	data, err := recipe.MarshalYAML(state.Recipe())

	s.Require().NoError(err)
	s.Require().Contains(string(data), "task_scheduler: gocron")
	s.Require().NotContains(string(data), "cron:")
	s.Require().NotContains(string(data), "batch_size:")
	s.Require().NotContains(string(data), "retention_period:")
	s.Require().NotContains(string(data), "worker:")
}

func (s *ConfigureWizardTestSuite) TestRecipeForNoSQLForcesSafeDatabaseDefaults() {
	for _, driver := range []string{recipe.DatabaseDriverRedis, recipe.DatabaseDriverMongoDB} {
		s.Run(driver, func() {
			state := NewConfigureWizardState(nil, ConfigureWizardOptions{})
			state.Name = "cache"
			state.Module = "github.com/example/cache"
			state.Database = driver
			state.DatabaseFramework = recipe.DatabaseFrameworkGORM
			state.Migrations = recipe.DatabaseMigrationsMigrate

			r := state.Recipe()

			s.Require().NoError(recipe.Validate(r))
			s.Require().Equal(driver, r.Database.Driver)
			s.Require().Equal(recipe.DatabaseFrameworkNone, r.Database.Framework)
			s.Require().Equal(recipe.DatabaseMigrationsNone, r.Database.Migrations)
		})
	}
}

func (s *ConfigureWizardTestSuite) TestFrameworkOptionsRejectInvalidPGXCombinations() {
	s.Require().True(hasFrameworkOption(frameworkOptions(recipe.DatabaseDriverPostgres), recipe.DatabaseFrameworkPGX))
	s.Require().False(hasFrameworkOption(frameworkOptions(recipe.DatabaseDriverMySQL), recipe.DatabaseFrameworkPGX))
	s.Require().False(hasFrameworkOption(frameworkOptions(recipe.DatabaseDriverSQLite), recipe.DatabaseFrameworkPGX))
}

func (s *ConfigureWizardTestSuite) TestSkippedSteps() {
	cliState := NewConfigureWizardState(nil, ConfigureWizardOptions{})
	cliState.ProjectType = recipe.ProjectTypeCLI
	cliState.Database = recipe.DatabaseDriverNone

	s.Require().False(shouldShowConfigureStep(stepServer, cliState))
	s.Require().False(shouldShowConfigureStep(stepObservability, cliState))
	s.Require().False(shouldShowConfigureStep(stepDatabaseFramework, cliState))
	s.Require().False(shouldShowConfigureStep(stepMigrations, cliState))

	sqlState := NewConfigureWizardState(nil, ConfigureWizardOptions{})
	sqlState.Database = recipe.DatabaseDriverMySQL

	s.Require().True(shouldShowConfigureStep(stepDatabaseFramework, sqlState))
	s.Require().True(shouldShowConfigureStep(stepMigrations, sqlState))
}

func hasFrameworkOption(options []components.SelectOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}
