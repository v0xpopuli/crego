package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/recipe"
)

type RecipeEditorTestSuite struct {
	suite.Suite
}

func TestRecipeEditorTestSuite(t *testing.T) {
	suite.Run(t, new(RecipeEditorTestSuite))
}

func (s *RecipeEditorTestSuite) TestRecipePreservesIdentityAndEditsFullMatrix() {
	source, err := recipe.NewPreset(recipe.PresetWebPostgres)
	s.Require().NoError(err)
	source.Project.Name = "orders"
	source.Project.Module = "github.com/example/orders"

	state := NewRecipeEditorState(source, RecipeEditorOptions{RecipePath: "crego.yaml"})
	state.ConfigurationFormat = recipe.ConfigurationFormatTOML
	state.Logging = recipe.LoggingFrameworkLogrus
	state.Server = recipe.ServerFrameworkFiber
	state.Database = recipe.DatabaseDriverPostgres
	state.DatabaseFramework = recipe.DatabaseFrameworkGORM
	state.Migrations = recipe.DatabaseMigrationsGoose
	state.TaskScheduler = recipe.TaskSchedulerGocron
	state.GitLabCI = true

	r := state.Recipe()

	s.Require().NoError(recipe.Validate(r))
	s.Require().Equal("orders", r.Project.Name)
	s.Require().Equal("github.com/example/orders", r.Project.Module)
	s.Require().Equal(recipe.ConfigurationFormatTOML, r.Configuration.Format)
	s.Require().Equal(recipe.LoggingFrameworkLogrus, r.Logging.Framework)
	s.Require().Equal(recipe.ServerFrameworkFiber, r.Server.Framework)
	s.Require().Equal(recipe.DatabaseFrameworkGORM, r.Database.Framework)
	s.Require().Equal(recipe.DatabaseMigrationsGoose, r.Database.Migrations)
	s.Require().Equal(recipe.TaskSchedulerGocron, r.TaskScheduler)
	s.Require().True(r.CI.GitLabCI)
}

func (s *RecipeEditorTestSuite) TestViewShowsLiveComponentsAndFiles() {
	source, err := recipe.NewPreset(recipe.PresetWebPostgres)
	s.Require().NoError(err)
	state := NewRecipeEditorState(source, RecipeEditorOptions{})
	state.ConfigurationFormat = recipe.ConfigurationFormatYAML
	state.Logging = recipe.LoggingFrameworkZap
	state.TaskScheduler = recipe.TaskSchedulerGocron
	state.GitLabCI = true

	screen := newRecipeEditorScreen(NewStyles(nil, true), state)
	view := screen.View()

	s.Require().Contains(view, "Config: yaml")
	s.Require().Contains(view, "Logging: zap")
	s.Require().Contains(view, "Task scheduler: gocron")
	s.Require().Contains(view, "GitLab CI: true")
	s.Require().Contains(view, "Components:")
	s.Require().Contains(view, "configuration.yaml")
	s.Require().Contains(view, "logging.zap")
	s.Require().Contains(view, "task_scheduler.gocron")
	s.Require().Contains(view, "ci.gitlab_ci")
	s.Require().Contains(view, "Files:")
	s.Require().Contains(view, ".gitlab-ci.yml")
}

func (s *RecipeEditorTestSuite) TestChangingDatabaseBlocksInvalidFrameworkChoices() {
	state := NewRecipeEditorState(nil, RecipeEditorOptions{})
	state.Database = recipe.DatabaseDriverMySQL
	state.DatabaseFramework = recipe.DatabaseFrameworkPGX

	state.applyDatabaseDefaults()

	s.Require().Equal(recipe.DatabaseFrameworkDatabaseSQL, state.DatabaseFramework)
	screen := newRecipeEditorScreen(NewStyles(nil, true), state)
	for _, field := range screen.fields() {
		if field.id == editorFramework {
			s.Require().NotContains(field.values, recipe.DatabaseFrameworkPGX)
		}
	}
}

func (s *RecipeEditorTestSuite) TestKeyboardEditsFields() {
	state := NewRecipeEditorState(nil, RecipeEditorOptions{})
	screen := newRecipeEditorScreen(NewStyles(nil, true), state)
	screen.cursor = fieldIndex(screen.fields(), editorConfiguration)

	next, _ := screen.Update(tea.KeyMsg{Type: tea.KeyRight})

	updated := next.(recipeEditorScreen)
	s.Require().Equal(recipe.ConfigurationFormatYAML, updated.state.ConfigurationFormat)

	updated.cursor = fieldIndex(updated.fields(), editorGitLabCI)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	updated = next.(recipeEditorScreen)

	s.Require().True(updated.state.GitLabCI)
}

func (s *RecipeEditorTestSuite) TestSaveAsWritesDifferentValidatedRecipe() {
	source, err := recipe.NewPreset(recipe.PresetWebPostgres)
	s.Require().NoError(err)
	source.Project.Name = "orders"
	source.Project.Module = "github.com/example/orders"
	dir := s.T().TempDir()
	original := filepath.Join(dir, "crego.yaml")
	saveAs := filepath.Join(dir, "company-web.yaml")
	s.Require().NoError(recipe.Save(original, source))

	state := NewRecipeEditorState(source, RecipeEditorOptions{RecipePath: original, SavePath: saveAs})
	state.ConfigurationFormat = recipe.ConfigurationFormatJSON
	state.GitLabCI = true
	screen := newRecipeEditorScreen(NewStyles(nil, true), state)

	s.Require().NoError(screen.save())
	s.Require().True(state.Saved())

	originalData, err := os.ReadFile(original)
	s.Require().NoError(err)
	s.Require().NotContains(string(originalData), "format: json")

	saved, err := recipe.Load(saveAs)
	s.Require().NoError(err)
	s.Require().Equal(recipe.ConfigurationFormatJSON, saved.Configuration.Format)
	s.Require().True(saved.CI.GitLabCI)
}

func (s *RecipeEditorTestSuite) TestReadonlyPreventsWrites() {
	source, err := recipe.NewPreset(recipe.PresetWebPostgres)
	s.Require().NoError(err)
	path := filepath.Join(s.T().TempDir(), "crego.yaml")
	state := NewRecipeEditorState(source, RecipeEditorOptions{RecipePath: path, ReadOnly: true})
	screen := newRecipeEditorScreen(NewStyles(nil, true), state)

	err = screen.save()

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "readonly")
	_, statErr := os.Stat(path)
	s.Require().True(os.IsNotExist(statErr))
}

func (s *RecipeEditorTestSuite) TestReadonlyKeyboardDoesNotMutate() {
	state := NewRecipeEditorState(nil, RecipeEditorOptions{ReadOnly: true})
	screen := newRecipeEditorScreen(NewStyles(nil, true), state)
	screen.cursor = fieldIndex(screen.fields(), editorConfiguration)

	next, _ := screen.Update(tea.KeyMsg{Type: tea.KeyRight})

	s.Require().Equal(recipe.ConfigurationFormatEnv, next.(recipeEditorScreen).state.ConfigurationFormat)
}

func fieldIndex(fields []recipeEditorField, id string) int {
	for index, field := range fields {
		if field.id == id {
			return index
		}
	}
	return 0
}

func (s *RecipeEditorTestSuite) TestSchedulerYAMLOmitsTaskLevelSettings() {
	state := NewRecipeEditorState(nil, RecipeEditorOptions{})
	state.TaskScheduler = recipe.TaskSchedulerGocron

	data, err := recipe.MarshalYAML(state.Recipe())

	s.Require().NoError(err)
	output := string(data)
	s.Require().Contains(output, "task_scheduler: gocron")
	s.Require().False(strings.Contains(output, "cron:"))
	s.Require().False(strings.Contains(output, "batch_size:"))
	s.Require().False(strings.Contains(output, "retention_period:"))
}
