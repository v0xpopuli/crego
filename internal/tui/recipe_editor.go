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
		options  RecipeEditorOptions
		source   *recipe.Recipe
		saved    bool
		modified bool

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
		AzurePipelines      bool
	}

	recipeEditorScreen struct {
		styles Styles
		state  *RecipeEditorState
		cursor int
		err    error

		width  int
		height int
	}

	recipeEditorField struct {
		id      string
		label   string
		current string
		values  []string
		toggle  bool
		enabled bool
	}

	editorSection struct {
		title  string
		fields []string
	}
)

const (
	editorProjectType    = "project_type"
	editorLayout         = "layout"
	editorServer         = "server"
	editorConfiguration  = "configuration"
	editorLogging        = "logging"
	editorDatabase       = "database"
	editorFramework      = "framework"
	editorMigrations     = "migrations"
	editorHealth         = "health"
	editorReadiness      = "readiness"
	editorScheduler      = "scheduler"
	editorDocker         = "docker"
	editorCompose        = "compose"
	editorGitHubActions  = "github_actions"
	editorGitLabCI       = "gitlab_ci"
	editorAzurePipelines = "azure_pipelines"
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
		AzurePipelines:      resolved.CI.AzurePipelines,
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
		model: NewModel(new(newRecipeEditorScreen(styles, state)), styles),
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
	r.CI.AzurePipelines = s.AzurePipelines

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

func (s *recipeEditorScreen) UsesShellLayout() bool {
	return true
}

func (s *recipeEditorScreen) Init() tea.Cmd {
	return nil
}

func (s *recipeEditorScreen) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tea.KeyMsg:
		key := msg.String()

		fields := s.fields()

		switch key {
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
	default:
		return s, nil
	}
}

func (s *recipeEditorScreen) View() string {
	r := s.state.Recipe()

	errorText := ""
	if s.err != nil {
		errorText = components.ErrorPanel(s.styles.Components(), s.err)
	}

	return RenderShell(s.styles, LayoutProps{
		Title:     "crego recipe edit",
		Subtitle:  s.editorStatusLine(),
		Sidebar:   s.sectionsView(),
		Body:      s.editorView(),
		Preview:   s.structuredPreview(),
		StackLine: s.styles.Description.Render("Stack: ") + editorStackSummary(r),
		Help:      s.editorHelpLine(),
		Error:     errorText,
		Width:     s.width,
		Height:    s.height,
	})
}

func editorStackSummary(r *recipe.Recipe) string {
	if r == nil {
		return ""
	}

	values := make([]string, 0, 12)

	appendIf := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" ||
			value == recipe.DatabaseDriverNone ||
			value == recipe.DatabaseFrameworkNone ||
			value == recipe.DatabaseMigrationsNone ||
			value == recipe.TaskSchedulerNone {
			return
		}

		values = append(values, value)
	}

	appendIf(r.Project.Type)
	appendIf(r.Layout.Style)

	if r.Go.Version != "" {
		appendIf("Go " + r.Go.Version)
	}

	if r.Project.Type == recipe.ProjectTypeWeb {
		appendIf(r.Server.Framework)
	}

	appendIf(r.Configuration.Format)
	appendIf(r.Logging.Framework)

	drivers := recipe.DatabaseDrivers(r.Database)
	if len(drivers) > 0 {
		if r.Database.Framework != "" && r.Database.Framework != recipe.DatabaseFrameworkNone {
			appendIf(drivers[0] + "/" + r.Database.Framework)
		} else {
			appendIf(drivers[0])
		}
	}

	appendIf(r.Database.Migrations)
	appendIf(r.TaskScheduler)

	if r.Deployment.Docker {
		appendIf("docker")
	}
	if r.Deployment.Compose {
		appendIf("compose")
	}
	if r.CI.GitHubActions {
		appendIf("github actions")
	}
	if r.CI.GitLabCI {
		appendIf("gitlab ci")
	}
	if r.CI.AzurePipelines {
		appendIf("azure pipelines")
	}

	return strings.Join(values, " · ")
}

func (s *recipeEditorScreen) editorHelpLine() string {
	if s.state.options.ReadOnly {
		return "↑↓ move · tab preview · readonly · esc back · q quit"
	}

	return "↑↓ field · ←→ change · space toggle · s save · esc back · q quit"
}

func (s *recipeEditorScreen) editorView() string {
	fields := s.fields()
	if len(fields) == 0 || s.cursor >= len(fields) {
		return ""
	}

	field := fields[s.cursor]

	lines := []string{
		s.styles.Title.Render(field.label),
		s.styles.Description.Render(fieldDescription(field.id, s.state)),
		"",
	}

	if !field.enabled {
		lines = append(lines, s.styles.Description.Render(inactiveReason(field.id, s.state)))
		return strings.Join(lines, "\n")
	}

	if field.toggle {
		options := []struct {
			value       bool
			label       string
			description string
		}{
			{true, "enabled", enabledDescription(field.id)},
			{false, "disabled", disabledDescription(field.id)},
		}

		current := field.currentBool()
		for _, option := range options {
			prefix := "  "
			line := fmt.Sprintf("%s%-10s %s", prefix, option.label, option.description)
			if option.value == current {
				line = s.styles.Selected.Render("› " + fmt.Sprintf("%-10s %s", option.label, option.description))
			} else {
				line = s.styles.Description.Render(line)
			}
			lines = append(lines, line)
		}

		return strings.Join(lines, "\n")
	}

	for _, value := range field.values {
		description := optionDescription(field.id, value)
		line := fmt.Sprintf("  %-12s %s", value, description)

		if value == field.current {
			line = s.styles.Selected.Render("› " + fmt.Sprintf("%-12s %s", value, description))
		} else {
			line = s.styles.Description.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (s *recipeEditorScreen) structuredPreview() string {
	r := s.state.Recipe()

	lines := []string{s.styles.Title.Render("Preview")}

	if err := recipe.Validate(r); err != nil {
		lines = append(lines, "", s.styles.Error.Render("Validation"))
		for _, problem := range recipeErrorLines(err) {
			lines = append(lines, s.styles.Error.Render("! "+problem))
		}
		return strings.Join(lines, "\n")
	}

	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		lines = append(lines, "", s.styles.Error.Render("Plan"))
		lines = append(lines, s.styles.Error.Render("! "+err.Error()))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "", s.styles.Description.Render("Components"))
	for _, current := range plan.Components {
		lines = append(lines, "  "+current.ID)
	}

	files, err := generator.RenderFileTargets(r, plan)
	if err != nil {
		lines = append(lines, "", s.styles.Error.Render("Files"))
		lines = append(lines, s.styles.Error.Render("! "+err.Error()))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "", s.styles.Description.Render("Files"))
	for _, file := range files {
		lines = append(lines, "  "+file.Target)
	}

	return strings.Join(lines, "\n")
}

func inactiveReason(fieldID string, state *RecipeEditorState) string {
	switch fieldID {
	case editorServer:
		return "Inactive because project type is CLI.\nWeb server is only used for web projects."
	case editorFramework:
		if isNoSQLDatabase(state.Database) {
			return fmt.Sprintf("Inactive because database driver is %s.\nNo ORM/framework is used for NoSQL databases.", state.Database)
		}
		return "Inactive because no SQL database is selected."
	case editorMigrations:
		if isNoSQLDatabase(state.Database) {
			return fmt.Sprintf("Inactive because database driver is %s.\nSQL migration tools are not used for NoSQL databases.", state.Database)
		}
		return "Inactive because no SQL database is selected."
	case editorHealth, editorReadiness:
		return "Inactive because project type is CLI.\nHTTP endpoints are only generated for web projects."
	default:
		return "This field is inactive for the current recipe."
	}
}

func fieldDescription(id string, state *RecipeEditorState) string {
	switch id {
	case editorProjectType:
		return "Choose generated project kind."
	case editorLayout:
		return "Choose how much structure the generated project should have."
	case editorServer:
		return "Choose HTTP server/router implementation."
	case editorConfiguration:
		return "Choose generated application configuration format."
	case editorLogging:
		return "Choose logging provider."
	case editorDatabase:
		return "Choose database integration."
	case editorFramework:
		return "Choose SQL database access layer."
	case editorMigrations:
		return "Choose SQL migration tool."
	case editorHealth:
		return "Generate basic health endpoint."
	case editorReadiness:
		return "Generate dependency readiness endpoint."
	case editorScheduler:
		return "Generate scheduled task executor support."
	case editorDocker:
		return "Generate Dockerfile."
	case editorCompose:
		return "Generate docker-compose.yml."
	case editorGitHubActions:
		return "Generate GitHub Actions workflow."
	case editorGitLabCI:
		return "Generate GitLab CI pipeline."
	case editorAzurePipelines:
		return "Generate Azure Pipelines workflow."
	default:
		return ""
	}
}

func optionDescription(fieldID string, value string) string {
	switch fieldID {
	case editorProjectType:
		switch value {
		case recipe.ProjectTypeWeb:
			return "HTTP API/application scaffold"
		case recipe.ProjectTypeCLI:
			return "Command-line application scaffold"
		}
	case editorLayout:
		switch value {
		case recipe.LayoutStyleMinimal:
			return "Small apps, fewer packages"
		case recipe.LayoutStyleLayered:
			return "Separated app, config, server, database packages"
		}
	case editorServer:
		switch value {
		case recipe.ServerFrameworkNetHTTP:
			return "Standard library, zero external router"
		case recipe.ServerFrameworkChi:
			return "Lightweight router, clean APIs"
		case recipe.ServerFrameworkGin:
			return "Popular web framework"
		case recipe.ServerFrameworkEcho:
			return "Minimal framework with middleware"
		case recipe.ServerFrameworkFiber:
			return "Express-like API, fasthttp-based"
		}
	case editorConfiguration:
		switch value {
		case recipe.ConfigurationFormatEnv:
			return "Environment variables only"
		case recipe.ConfigurationFormatYAML:
			return "Human-friendly config files"
		case recipe.ConfigurationFormatJSON:
			return "Machine-friendly config files"
		case recipe.ConfigurationFormatTOML:
			return "Compact structured config"
		}
	case editorLogging:
		switch value {
		case recipe.LoggingFrameworkSlog:
			return "Standard library, simple and stable"
		case recipe.LoggingFrameworkZap:
			return "High-performance structured logging"
		case recipe.LoggingFrameworkZerolog:
			return "Minimal allocations, JSON-first"
		case recipe.LoggingFrameworkLogrus:
			return "Mature, widely known"
		}
	case editorDatabase:
		switch value {
		case recipe.DatabaseDriverNone:
			return "No database integration"
		case recipe.DatabaseDriverPostgres:
			return "SQL database, production default"
		case recipe.DatabaseDriverMySQL:
			return "SQL database, common web stack"
		case recipe.DatabaseDriverSQLite:
			return "Local/file database"
		case recipe.DatabaseDriverRedis:
			return "Key-value NoSQL database"
		case recipe.DatabaseDriverMongoDB:
			return "Document NoSQL database"
		}
	case editorFramework:
		switch value {
		case recipe.DatabaseFrameworkNone:
			return "No SQL framework"
		case recipe.DatabaseFrameworkPGX:
			return "PostgreSQL-native driver"
		case recipe.DatabaseFrameworkDatabaseSQL:
			return "Standard SQL abstraction"
		case recipe.DatabaseFrameworkGORM:
			return "ORM with models and relations"
		}
	case editorMigrations:
		switch value {
		case recipe.DatabaseMigrationsNone:
			return "No SQL migrations"
		case recipe.DatabaseMigrationsGoose:
			return "Goose migrations"
		case recipe.DatabaseMigrationsMigrate:
			return "golang-migrate migrations"
		}
	case editorScheduler:
		switch value {
		case recipe.TaskSchedulerNone:
			return "No scheduled task executor"
		case recipe.TaskSchedulerGocron:
			return "Generate gocron scheduler and example task"
		}
	}

	return ""
}

func enabledDescription(fieldID string) string {
	switch fieldID {
	case editorHealth:
		return "Generate GET /health"
	case editorReadiness:
		return "Generate GET /ready"
	case editorDocker:
		return "Generate Dockerfile"
	case editorCompose:
		return "Generate docker-compose.yml"
	case editorGitHubActions:
		return "Generate GitHub Actions workflow"
	case editorGitLabCI:
		return "Generate GitLab CI pipeline"
	case editorAzurePipelines:
		return "Generate Azure Pipelines workflow"
	default:
		return "Enable this option"
	}
}

func disabledDescription(fieldID string) string {
	switch fieldID {
	case editorHealth:
		return "Do not generate health endpoint"
	case editorReadiness:
		return "Do not generate readiness endpoint"
	case editorDocker:
		return "Do not generate Dockerfile"
	case editorCompose:
		return "Do not generate docker-compose.yml"
	case editorGitHubActions:
		return "Do not generate GitHub Actions workflow"
	case editorGitLabCI:
		return "Do not generate GitLab CI pipeline"
	case editorAzurePipelines:
		return "Do not generate Azure Pipelines workflow"
	default:
		return "Disable this option"
	}
}

func (s *recipeEditorScreen) sectionsView() string {
	fields := s.fields()
	indexByID := make(map[string]int, len(fields))
	fieldByID := make(map[string]recipeEditorField, len(fields))

	for i, field := range fields {
		indexByID[field.id] = i
		fieldByID[field.id] = field
	}

	sections := []editorSection{
		{
			title: "Project",
			fields: []string{
				editorProjectType,
				editorLayout,
			},
		},
		{
			title: "Web stack",
			fields: []string{
				editorServer,
				editorConfiguration,
				editorLogging,
			},
		},
		{
			title: "Data layer",
			fields: []string{
				editorDatabase,
				editorFramework,
				editorMigrations,
			},
		},
		{
			title: "Features",
			fields: []string{
				editorHealth,
				editorReadiness,
				editorScheduler,
			},
		},
		{
			title: "Delivery",
			fields: []string{
				editorDocker,
				editorCompose,
				editorGitHubActions,
				editorGitLabCI,
				editorAzurePipelines,
			},
		},
	}

	lines := []string{s.styles.Title.Render("Recipe")}

	for _, section := range sections {
		lines = append(lines, "", s.styles.Description.Render(section.title))

		for _, fieldID := range section.fields {
			idx, ok := indexByID[fieldID]
			if !ok {
				continue
			}

			field := fieldByID[fieldID]
			marker := "○"
			if idx == s.cursor {
				marker = "●"
			} else if field.enabled {
				marker = "✓"
			}

			value := field.current
			if field.toggle {
				if field.currentBool() {
					value = "enabled"
				} else {
					value = "disabled"
				}
			}

			label := fmt.Sprintf("%s %s: %s", marker, field.label, value)
			if !field.enabled {
				label = s.styles.Description.Render(label + " (inactive)")
			} else if idx == s.cursor {
				label = s.styles.Selected.Render(label)
			}
			lines = append(lines, "  "+label)
		}
	}

	return strings.Join(lines, "\n")
}

func (s *recipeEditorScreen) editorStatusLine() string {
	mode := "clean"
	if s.state.modified {
		mode = "modified •"
	}
	if s.state.saved {
		mode = "saved ✓"
	}

	validation := "valid ✓"
	if err := recipe.Validate(s.state.Recipe()); err != nil {
		validation = "invalid !"
	}

	if s.state.options.ReadOnly {
		mode = "readonly"
	}

	source := s.state.RecipePath()
	target := s.state.SavePath()
	if source == target {
		return fmt.Sprintf("Editing: %s  %s · %s", source, mode, validation)
	}

	return fmt.Sprintf("Editing: %s → %s  %s · %s", source, target, mode, validation)
}

func (s *recipeEditorScreen) fields() []recipeEditorField {
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
		{id: editorAzurePipelines, label: "Azure Pipelines", current: boolValue(s.state.AzurePipelines), toggle: true, enabled: true},
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
	s.state.modified = true
	s.state.saved = false

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
	case editorAzurePipelines:
		s.state.AzurePipelines = value.(bool)
	}
}

func (s *recipeEditorScreen) save() error {
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
	s.state.modified = false
	return nil
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
