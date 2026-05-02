package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/generator"
	"github.com/v0xpopuli/crego/internal/recipe"
	"github.com/v0xpopuli/crego/internal/tui/components"
	"github.com/v0xpopuli/crego/internal/tui/screens"
)

type (
	RecipeEditorOptions struct {
		RecipePath string
		SavePath   string
		ReadOnly   bool
	}

	RecipeEditorState struct {
		options RecipeEditorOptions
		source  *recipe.Recipe
		saved   bool

		ProjectType         string
		Layout              string
		Server              string
		ConfigurationFormat string
		Logging             string
		Database            string
		DatabaseFramework   string
		Migrations          string
		Health              bool
		Readiness           bool
		TaskScheduler       string
		Docker              bool
		Compose             bool
		GitHubActions       bool
		GitLabCI            bool
	}

	recipeEditorScreen struct {
		styles Styles
		state  *RecipeEditorState
		cursor int
		err    error
	}

	recipeEditorField struct {
		id      string
		label   string
		current string
		values  []string
		toggle  bool
		enabled bool
	}
)

const (
	editorProjectType   = "project_type"
	editorLayout        = "layout"
	editorServer        = "server"
	editorConfiguration = "configuration"
	editorLogging       = "logging"
	editorDatabase      = "database"
	editorFramework     = "framework"
	editorMigrations    = "migrations"
	editorHealth        = "health"
	editorReadiness     = "readiness"
	editorScheduler     = "scheduler"
	editorDocker        = "docker"
	editorCompose       = "compose"
	editorGitHubActions = "github_actions"
	editorGitLabCI      = "gitlab_ci"
)

func NewRecipeEditorState(source *recipe.Recipe, opts RecipeEditorOptions) *RecipeEditorState {
	if opts.RecipePath == "" {
		opts.RecipePath = "crego.yaml"
	}
	if opts.SavePath == "" {
		opts.SavePath = opts.RecipePath
	}
	if source == nil {
		source, _ = recipe.NewPreset(recipe.PresetWebBasic)
	}

	resolved := *source
	recipe.Normalize(&resolved)
	recipe.ApplyDefaults(&resolved)

	drivers := recipe.DatabaseDrivers(resolved.Database)
	databaseDriver := recipe.DatabaseDriverNone
	if len(drivers) > 0 {
		databaseDriver = drivers[0]
	}

	state := &RecipeEditorState{
		options:             opts,
		source:              &resolved,
		ProjectType:         resolved.Project.Type,
		Layout:              resolved.Layout.Style,
		Server:              resolved.Server.Framework,
		ConfigurationFormat: resolved.Configuration.Format,
		Logging:             resolved.Logging.Framework,
		Database:            databaseDriver,
		DatabaseFramework:   resolved.Database.Framework,
		Migrations:          resolved.Database.Migrations,
		Health:              resolved.Observability.Health,
		Readiness:           resolved.Observability.Readiness,
		TaskScheduler:       resolved.TaskScheduler,
		Docker:              resolved.Deployment.Docker,
		Compose:             resolved.Deployment.Compose,
		GitHubActions:       resolved.CI.GitHubActions,
		GitLabCI:            resolved.CI.GitLabCI,
	}
	state.applyDatabaseDefaults()
	return state
}

func NewRecipeEditorApp(state *RecipeEditorState, opts AppOptions) *App {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	styles := NewStyles(opts.Out, opts.NoColor)
	return &App{
		model: NewModel(newRecipeEditorScreen(styles, state), styles),
		in:    opts.In,
		out:   opts.Out,
	}
}

func (s *RecipeEditorState) RecipePath() string {
	return s.options.RecipePath
}

func (s *RecipeEditorState) SavePath() string {
	return s.options.SavePath
}

func (s *RecipeEditorState) Saved() bool {
	return s.saved
}

func (s *RecipeEditorState) Recipe() *recipe.Recipe {
	var r recipe.Recipe
	if s.source != nil {
		r = *s.source
	}

	r.Project.Type = s.ProjectType
	r.Layout.Style = s.Layout
	r.Configuration.Format = s.ConfigurationFormat
	r.Logging.Framework = s.Logging
	r.SQLDatabase = ""
	r.ORMFramework = ""
	r.NoSQLDatabase = nil
	r.Migrations = ""
	r.Database = recipe.DatabaseConfig{
		Driver:     s.Database,
		Framework:  s.DatabaseFramework,
		Migrations: s.Migrations,
	}
	r.TaskScheduler = s.TaskScheduler
	r.Observability.Health = s.ProjectType == recipe.ProjectTypeWeb && s.Health
	r.Observability.Readiness = s.ProjectType == recipe.ProjectTypeWeb && s.Readiness
	r.Deployment.Docker = s.Docker
	r.Deployment.Compose = s.Compose
	r.CI.GitHubActions = s.GitHubActions
	r.CI.GitLabCI = s.GitLabCI

	if s.ProjectType == recipe.ProjectTypeWeb {
		r.Server.Framework = s.Server
		if r.Server.Port == 0 {
			r.Server.Port = 8080
		}
	} else {
		r.Server = recipe.ServerConfig{}
	}

	if isSQLDatabase(s.Database) {
		r.Database.SQL = s.Database
	} else if isNoSQLDatabase(s.Database) {
		r.Database.NoSQL = recipe.NoSQLDrivers{s.Database}
		r.Database.Framework = recipe.DatabaseFrameworkNone
		r.Database.Migrations = recipe.DatabaseMigrationsNone
	} else {
		r.Database.Framework = recipe.DatabaseFrameworkNone
		r.Database.Migrations = recipe.DatabaseMigrationsNone
	}

	recipe.Normalize(&r)
	recipe.ApplyDefaults(&r)
	return &r
}

func newRecipeEditorScreen(styles Styles, state *RecipeEditorState) recipeEditorScreen {
	if state == nil {
		state = NewRecipeEditorState(nil, RecipeEditorOptions{})
	}
	return recipeEditorScreen{styles: styles, state: state}
}

func (s recipeEditorScreen) Init() tea.Cmd {
	return nil
}

func (s recipeEditorScreen) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	fields := s.fields()
	switch key.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(fields)-1 {
			s.cursor++
		}
	case "left", "h":
		s.changeCurrent(-1)
	case "right", "l", "enter", " ":
		s.changeCurrent(1)
	case "s", "ctrl+s":
		if err := s.save(); err != nil {
			s.err = err
			return s, nil
		}
		return s, tea.Quit
	}

	return s, nil
}

func (s recipeEditorScreen) View() string {
	parts := []string{
		s.styles.Title.Render("crego recipe edit"),
		s.styles.Description.Render(s.description()),
	}
	if s.err != nil {
		parts = append(parts, components.ErrorPanel(s.styles.Components(), s.err))
	}
	parts = append(parts, s.columnsView())
	parts = append(parts, s.styles.Footer.Render(s.hint()))
	return strings.Join(parts, "\n\n")
}

func (s recipeEditorScreen) fields() []recipeEditorField {
	fields := []recipeEditorField{
		{id: editorProjectType, label: "Project", current: s.state.ProjectType, values: []string{recipe.ProjectTypeWeb, recipe.ProjectTypeCLI}, enabled: true},
		{id: editorLayout, label: "Layout", current: s.state.Layout, values: []string{recipe.LayoutStyleMinimal, recipe.LayoutStyleLayered}, enabled: true},
		{id: editorServer, label: "Server", current: s.state.Server, values: []string{recipe.ServerFrameworkNetHTTP, recipe.ServerFrameworkChi, recipe.ServerFrameworkGin, recipe.ServerFrameworkEcho, recipe.ServerFrameworkFiber}, enabled: s.state.ProjectType == recipe.ProjectTypeWeb},
		{id: editorConfiguration, label: "Config", current: s.state.ConfigurationFormat, values: []string{recipe.ConfigurationFormatEnv, recipe.ConfigurationFormatYAML, recipe.ConfigurationFormatJSON, recipe.ConfigurationFormatTOML}, enabled: true},
		{id: editorLogging, label: "Logging", current: s.state.Logging, values: []string{recipe.LoggingFrameworkSlog, recipe.LoggingFrameworkZap, recipe.LoggingFrameworkZerolog, recipe.LoggingFrameworkLogrus}, enabled: true},
		{id: editorDatabase, label: "Database", current: s.state.Database, values: []string{recipe.DatabaseDriverNone, recipe.DatabaseDriverPostgres, recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite, recipe.DatabaseDriverRedis, recipe.DatabaseDriverMongoDB}, enabled: true},
		{id: editorFramework, label: "DB framework", current: s.state.DatabaseFramework, values: compatibleFrameworks(s.state.Database), enabled: isSQLDatabase(s.state.Database)},
		{id: editorMigrations, label: "Migrations", current: s.state.Migrations, values: []string{recipe.DatabaseMigrationsNone, recipe.DatabaseMigrationsGoose, recipe.DatabaseMigrationsMigrate}, enabled: isSQLDatabase(s.state.Database)},
		{id: editorHealth, label: "Health", current: boolValue(s.state.Health), toggle: true, enabled: s.state.ProjectType == recipe.ProjectTypeWeb},
		{id: editorReadiness, label: "Readiness", current: boolValue(s.state.Readiness), toggle: true, enabled: s.state.ProjectType == recipe.ProjectTypeWeb},
		{id: editorScheduler, label: "Task scheduler", current: s.state.TaskScheduler, values: []string{recipe.TaskSchedulerNone, recipe.TaskSchedulerGocron}, enabled: true},
		{id: editorDocker, label: "Docker", current: boolValue(s.state.Docker), toggle: true, enabled: true},
		{id: editorCompose, label: "Compose", current: boolValue(s.state.Compose), toggle: true, enabled: true},
		{id: editorGitHubActions, label: "GitHub Actions", current: boolValue(s.state.GitHubActions), toggle: true, enabled: true},
		{id: editorGitLabCI, label: "GitLab CI", current: boolValue(s.state.GitLabCI), toggle: true, enabled: true},
	}
	return fields
}

func (s *RecipeEditorState) applyDatabaseDefaults() {
	switch s.Database {
	case recipe.DatabaseDriverPostgres:
		if !contains(compatibleFrameworks(s.Database), s.DatabaseFramework) {
			s.DatabaseFramework = recipe.DatabaseFrameworkPGX
		}
		if s.Migrations == "" {
			s.Migrations = recipe.DatabaseMigrationsNone
		}
	case recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite:
		if !contains(compatibleFrameworks(s.Database), s.DatabaseFramework) {
			s.DatabaseFramework = recipe.DatabaseFrameworkDatabaseSQL
		}
		if s.Migrations == "" {
			s.Migrations = recipe.DatabaseMigrationsNone
		}
	default:
		s.DatabaseFramework = recipe.DatabaseFrameworkNone
		s.Migrations = recipe.DatabaseMigrationsNone
	}
}

func (s *recipeEditorScreen) changeCurrent(delta int) {
	if s.state.options.ReadOnly {
		return
	}

	fields := s.fields()
	if len(fields) == 0 || s.cursor >= len(fields) {
		return
	}
	field := fields[s.cursor]
	if !field.enabled {
		return
	}
	if field.toggle {
		s.setField(field.id, !field.currentBool())
		s.err = nil
		return
	}
	if len(field.values) == 0 {
		return
	}
	next := cycleValue(field.values, field.current, delta)
	s.setField(field.id, next)
	if field.id == editorDatabase {
		s.state.applyDatabaseDefaults()
	}
	s.err = nil
}

func (s *recipeEditorScreen) setField(id string, value any) {
	switch id {
	case editorProjectType:
		s.state.ProjectType = value.(string)
	case editorLayout:
		s.state.Layout = value.(string)
	case editorServer:
		s.state.Server = value.(string)
	case editorConfiguration:
		s.state.ConfigurationFormat = value.(string)
	case editorLogging:
		s.state.Logging = value.(string)
	case editorDatabase:
		s.state.Database = value.(string)
	case editorFramework:
		s.state.DatabaseFramework = value.(string)
	case editorMigrations:
		s.state.Migrations = value.(string)
	case editorHealth:
		s.state.Health = value.(bool)
	case editorReadiness:
		s.state.Readiness = value.(bool)
	case editorScheduler:
		s.state.TaskScheduler = value.(string)
	case editorDocker:
		s.state.Docker = value.(bool)
	case editorCompose:
		s.state.Compose = value.(bool)
	case editorGitHubActions:
		s.state.GitHubActions = value.(bool)
	case editorGitLabCI:
		s.state.GitLabCI = value.(bool)
	}
}

func (s recipeEditorScreen) save() error {
	if s.state.options.ReadOnly {
		return fmt.Errorf("readonly mode prevents saving")
	}
	r := s.state.Recipe()
	if err := recipe.Validate(r); err != nil {
		return err
	}
	if _, err := generator.Resolve(component.NewRegistry(), r); err != nil {
		return err
	}
	if err := recipe.Save(s.state.options.SavePath, r); err != nil {
		return err
	}
	s.state.saved = true
	return nil
}

func (s recipeEditorScreen) columnsView() string {
	left := s.recipeLines()
	right := s.previewLines()
	width := 38
	lines := make([]string, 0, maxInt(len(left), len(right)))
	count := maxInt(len(left), len(right))
	for i := 0; i < count; i++ {
		var leftLine, rightLine string
		if i < len(left) {
			leftLine = left[i]
		}
		if i < len(right) {
			rightLine = right[i]
		}
		lines = append(lines, padRight(leftLine, width)+"  "+rightLine)
	}
	return strings.Join(lines, "\n")
}

func (s recipeEditorScreen) recipeLines() []string {
	fields := s.fields()
	lines := []string{"Recipe"}
	for i, field := range fields {
		prefix := "  "
		if i == s.cursor {
			prefix = "> "
		}
		value := field.current
		if !field.enabled {
			value += " (inactive)"
		}
		line := fmt.Sprintf("%s%s: %s", prefix, field.label, value)
		if i == s.cursor {
			line = s.styles.Selected.Render(line)
		}
		lines = append(lines, line)
	}
	return lines
}

func (s recipeEditorScreen) previewLines() []string {
	r := s.state.Recipe()
	lines := []string{"Preview"}
	if err := recipe.Validate(r); err != nil {
		lines = append(lines, "Validation errors:")
		for _, problem := range recipeErrorLines(err) {
			lines = append(lines, "  "+problem)
		}
		return lines
	}

	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		lines = append(lines, "Plan errors:", "  "+err.Error())
		return lines
	}

	lines = append(lines, "Components:")
	for _, current := range plan.Components {
		lines = append(lines, "  "+current.ID)
	}
	files, err := generator.RenderFileTargets(r, plan)
	if err != nil {
		lines = append(lines, "", "File target errors:", "  "+err.Error())
		return lines
	}
	lines = append(lines, "", "Files:")
	for _, file := range files {
		lines = append(lines, "  "+file.Target)
	}
	return lines
}

func (s recipeEditorScreen) description() string {
	mode := fmt.Sprintf("Editing %s -> %s", s.state.RecipePath(), s.state.SavePath())
	if s.state.options.ReadOnly {
		mode += " (readonly)"
	}
	return mode
}

func (s recipeEditorScreen) hint() string {
	if s.state.options.ReadOnly {
		return "up/down move • readonly preview only • q/esc cancel"
	}
	return "up/down move • left/right change • space toggle • s save • q/esc cancel"
}

func recipeErrorLines(err error) []string {
	messages := strings.Split(err.Error(), "\n")
	result := make([]string, 0, len(messages))
	for _, message := range messages {
		message = strings.TrimSpace(message)
		if message != "" {
			result = append(result, message)
		}
	}
	return result
}

func cycleValue(values []string, current string, delta int) string {
	if len(values) == 0 {
		return current
	}
	index := 0
	for i, value := range values {
		if value == current {
			index = i
			break
		}
	}
	index = (index + delta) % len(values)
	if index < 0 {
		index += len(values)
	}
	return values[index]
}

func (f recipeEditorField) currentBool() bool {
	return f.current == "true"
}

func boolValue(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
