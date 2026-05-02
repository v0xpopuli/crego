package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/generator"
	"github.com/v0xpopuli/crego/internal/recipe"
	"github.com/v0xpopuli/crego/internal/tui/components"
	"github.com/v0xpopuli/crego/internal/tui/screens"
)

type (
	ConfigureWizardOptions struct {
		RecipePath    string
		OutputDir     string
		Minimal       bool
		Overwrite     bool
		Force         bool
		SkipGoModTidy bool
		Mode          ConfigureWizardMode
	}

	ConfigureWizardState struct {
		options ConfigureWizardOptions

		Name                string
		Module              string
		GoVersion           string
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

		Action         ConfigureWizardAction
		ForceOverwrite bool
		saved          bool
	}

	configureScreen struct {
		styles Styles
		state  *ConfigureWizardState
		step   configureStep

		selectInput components.Select
		multiInput  components.MultiSelect
		nameInput   textinput.Model
		moduleInput textinput.Model
		focus       int
		err         error
		width       int
		height      int
	}

	configureStep int

	ConfigureWizardMode string

	ConfigureWizardAction string
)

const (
	ConfigureWizardModeRecipe     ConfigureWizardMode = "recipe"
	ConfigureWizardModeGeneration ConfigureWizardMode = "generation"

	ConfigureWizardActionGenerate ConfigureWizardAction = "generate"
	ConfigureWizardActionSave     ConfigureWizardAction = "save"

	stepWelcome configureStep = iota
	stepProjectIdentity
	stepGoVersion
	stepProjectType
	stepLayout
	stepServer
	stepConfiguration
	stepLogging
	stepDatabase
	stepDatabaseFramework
	stepMigrations
	stepObservability
	stepTaskScheduler
	stepDeployment
	stepCI
	stepPreview
	stepSave
)

func NewConfigureWizardState(source *recipe.Recipe, opts ConfigureWizardOptions) *ConfigureWizardState {
	if opts.RecipePath == "" {
		opts.RecipePath = "crego.yaml"
	}
	if opts.Mode == "" {
		opts.Mode = ConfigureWizardModeRecipe
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

	return &ConfigureWizardState{
		options:             opts,
		Name:                resolved.Project.Name,
		Module:              resolved.Project.Module,
		GoVersion:           resolved.Go.Version,
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
}

func NewConfigureApp(state *ConfigureWizardState, opts AppOptions) *App {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	styles := NewStyles(opts.Out, opts.NoColor)
	return &App{
		model: NewModel(newConfigureScreen(styles, state, stepWelcome), styles),
		in:    opts.In,
		out:   opts.Out,
	}
}

func (s *ConfigureWizardState) RecipePath() string {
	return s.options.RecipePath
}

func (s *ConfigureWizardState) Saved() bool {
	return s.saved
}

func (s *ConfigureWizardState) OutputDirectory() string {
	if s == nil {
		return "."
	}
	if strings.TrimSpace(s.options.OutputDir) != "" {
		return strings.TrimSpace(s.options.OutputDir)
	}
	r := s.Recipe()
	if strings.TrimSpace(r.Project.Name) != "" {
		return strings.TrimSpace(r.Project.Name)
	}
	module := strings.TrimSuffix(strings.TrimSpace(r.Project.Module), "/")
	if module == "" {
		return "."
	}
	parts := strings.Split(module, "/")
	name := strings.TrimSpace(parts[len(parts)-1])
	if name == "" || name == "." {
		return "."
	}
	return name
}

func (s *ConfigureWizardState) Recipe() *recipe.Recipe {
	r := &recipe.Recipe{
		Version: recipe.VersionV1,
		Project: recipe.ProjectConfig{
			Name:   s.Name,
			Module: s.Module,
			Type:   s.ProjectType,
		},
		Go: recipe.GoConfig{
			Version: s.GoVersion,
		},
		Layout: recipe.LayoutConfig{
			Style: s.Layout,
		},
		Configuration: recipe.ConfigurationConfig{
			Format: s.ConfigurationFormat,
		},
		Database: recipe.DatabaseConfig{
			Driver:     s.Database,
			Framework:  s.DatabaseFramework,
			Migrations: s.Migrations,
		},
		TaskScheduler: s.TaskScheduler,
		Logging: recipe.LoggingConfig{
			Framework: s.Logging,
			Format:    recipe.LoggingFormatText,
		},
		Observability: recipe.ObservabilityConfig{
			Health:    s.ProjectType == recipe.ProjectTypeWeb && s.Health,
			Readiness: s.ProjectType == recipe.ProjectTypeWeb && s.Readiness,
		},
		Deployment: recipe.DeploymentConfig{
			Docker:  s.Docker,
			Compose: s.Compose,
		},
		CI: recipe.CIConfig{
			GitHubActions:  s.GitHubActions,
			GitLabCI:       s.GitLabCI,
			AzurePipelines: s.AzurePipelines,
		},
	}

	if s.ProjectType == recipe.ProjectTypeWeb {
		r.Server = recipe.ServerConfig{
			Framework:        s.Server,
			Port:             8080,
			GracefulShutdown: true,
		}
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

	recipe.Normalize(r)
	recipe.ApplyDefaults(r)
	return r
}

func newConfigureScreen(styles Styles, state *ConfigureWizardState, step configureStep) configureScreen {
	if state == nil {
		state = NewConfigureWizardState(nil, ConfigureWizardOptions{})
	}
	screen := configureScreen{
		styles: styles,
		state:  state,
		step:   step,
	}
	return screen.configureInputs()
}

func (s configureScreen) Init() tea.Cmd {
	if s.step == stepProjectIdentity {
		return s.nameInput.Focus()
	}
	return nil
}

func (s configureScreen) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		s.width = size.Width
		s.height = size.Height
		return s, nil
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		switch s.step {
		case stepProjectIdentity:
			return s.updateIdentity(key, msg)
		case stepDeployment, stepCI, stepObservability:
			return s.updateMultiSelect(key, msg)
		default:
			return s.updateSelect(key, msg)
		}
	}
	if s.step == stepProjectIdentity {
		return s.updateText(msg)
	}
	return s, nil
}

func (s configureScreen) View() string {
	var bodyParts []string
	if description := s.description(); description != "" {
		bodyParts = append(bodyParts, s.styles.Description.Render(description))
	}
	switch s.step {
	case stepProjectIdentity:
		bodyParts = append(bodyParts, s.identityView())
	case stepLayout:
		bodyParts = append(bodyParts, s.selectInput.View(s.styles.Components()))
		bodyParts = append(bodyParts, components.Preview(s.styles.Components(), s.layoutPreview()))
	case stepPreview:
		bodyParts = append(bodyParts, s.preview())
	case stepDeployment, stepCI, stepObservability:
		bodyParts = append(bodyParts, s.multiInput.View(s.styles.Components()))
	default:
		bodyParts = append(bodyParts, s.selectInput.View(s.styles.Components()))
	}

	errorPanel := ""
	if s.err != nil {
		errorPanel = components.ErrorPanel(s.styles.Components(), s.err)
	}
	step, total := s.progress()
	return RenderShell(s.styles, LayoutProps{
		Title:      s.title(),
		Subtitle:   s.stepTitle(),
		Step:       step,
		TotalSteps: total,
		Sidebar:    s.sidebar(),
		Body:       strings.Join(bodyParts, "\n\n"),
		Preview:    s.livePreview(),
		StackLine:  s.stackLine(),
		Help:       s.hint(),
		Error:      errorPanel,
		Width:      s.width,
		Height:     s.height,
	})
}

func (s configureScreen) UsesShellLayout() bool {
	return true
}

func (s configureScreen) updateSelect(key tea.KeyMsg, msg tea.Msg) (screens.Screen, tea.Cmd) {
	if isConfigureBackKey(key) {
		if s.step == stepWelcome {
			return s, func() tea.Msg { return Cancel() }
		}
		return s.previousScreen(), nil
	}
	if key.String() != "enter" {
		s.selectInput = s.selectInput.Update(msg)
		return s, nil
	}

	value := s.selectInput.Selected().Value
	switch s.step {
	case stepWelcome:
		if value == "cancel" {
			return s, func() tea.Msg { return Cancel() }
		}
	case stepProjectType:
		s.state.ProjectType = value
	case stepGoVersion:
		s.state.GoVersion = value
	case stepLayout:
		s.state.Layout = value
	case stepServer:
		s.state.Server = value
	case stepConfiguration:
		s.state.ConfigurationFormat = value
	case stepLogging:
		s.state.Logging = value
	case stepDatabase:
		s.state.Database = value
		s.applyDatabaseDefaults()
	case stepDatabaseFramework:
		s.state.DatabaseFramework = value
	case stepMigrations:
		s.state.Migrations = value
	case stepTaskScheduler:
		s.state.TaskScheduler = value
	case stepPreview:
		if value == "back" {
			return s.previousScreen(), nil
		}
		if value == "cancel" {
			return s, func() tea.Msg { return Cancel() }
		}
		if s.state.options.Mode == ConfigureWizardModeGeneration {
			switch value {
			case "generate", "generate_overwrite":
				s.state.Action = ConfigureWizardActionGenerate
				s.state.ForceOverwrite = value == "generate_overwrite"
				return s, tea.Quit
			case "save":
				if err := s.saveRecipe(false); err != nil {
					s.err = err
					return s, nil
				}
				s.state.Action = ConfigureWizardActionSave
				return s, tea.Quit
			}
		}
	case stepSave:
		switch value {
		case "save", "overwrite":
			if err := s.saveRecipe(value == "overwrite"); err != nil {
				s.err = err
				return s, nil
			}
			return s, tea.Quit
		case "back":
			return s.previousScreen(), nil
		case "cancel":
			return s, func() tea.Msg { return Cancel() }
		}
	}

	return s.nextScreen(), nil
}

func (s configureScreen) updateMultiSelect(key tea.KeyMsg, msg tea.Msg) (screens.Screen, tea.Cmd) {
	if isConfigureBackKey(key) {
		return s.previousScreen(), nil
	}
	if key.String() != "enter" {
		s.multiInput = s.multiInput.Update(msg)
		return s, nil
	}

	values := selectedSet(s.multiInput.Values())
	switch s.step {
	case stepObservability:
		s.state.Health = values["health"]
		s.state.Readiness = values["readiness"]
	case stepDeployment:
		s.state.Docker = values["docker"]
		s.state.Compose = values["compose"]
	case stepCI:
		s.state.GitHubActions = values["github_actions"]
		s.state.GitLabCI = values["gitlab_ci"]
		s.state.AzurePipelines = values["azure_pipelines"]
	}
	return s.nextScreen(), nil
}

func (s configureScreen) updateIdentity(key tea.KeyMsg, msg tea.Msg) (screens.Screen, tea.Cmd) {
	switch key.String() {
	case "ctrl+b", "esc":
		return s.previousScreen(), nil
	case "tab", "shift+tab", "up", "down":
		if s.focus == 0 {
			s.focus = 1
			s.nameInput.Blur()
			return s.updateTextFocus(s.moduleInput.Focus())
		}
		s.focus = 0
		s.moduleInput.Blur()
		return s.updateTextFocus(s.nameInput.Focus())
	case "enter":
		s.state.Name = strings.TrimSpace(s.nameInput.Value())
		s.state.Module = strings.TrimSpace(s.moduleInput.Value())
		if err := recipe.Validate(s.state.Recipe()); err != nil {
			s.err = err
			return s, nil
		}
		return s.nextScreen(), nil
	default:
		return s.updateText(msg)
	}
}

func (s configureScreen) updateText(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmd tea.Cmd
	if s.focus == 0 {
		s.nameInput, cmd = s.nameInput.Update(msg)
	} else {
		s.moduleInput, cmd = s.moduleInput.Update(msg)
	}
	s.state.Name = strings.TrimSpace(s.nameInput.Value())
	s.state.Module = strings.TrimSpace(s.moduleInput.Value())
	s.err = nil
	return s, cmd
}

func (s configureScreen) updateTextFocus(cmd tea.Cmd) (screens.Screen, tea.Cmd) {
	return s, cmd
}

func (s configureScreen) nextScreen() configureScreen {
	next := newConfigureScreen(s.styles, s.state, nextConfigureStep(s.step, s.state))
	next.width = s.width
	next.height = s.height
	return next
}

func (s configureScreen) previousScreen() configureScreen {
	previous := newConfigureScreen(s.styles, s.state, previousConfigureStep(s.step, s.state))
	previous.width = s.width
	previous.height = s.height
	return previous
}

func (s configureScreen) applyDatabaseDefaults() {
	switch s.state.Database {
	case recipe.DatabaseDriverPostgres:
		if !contains(compatibleFrameworks(s.state.Database), s.state.DatabaseFramework) {
			s.state.DatabaseFramework = recipe.DatabaseFrameworkPGX
		}
		if s.state.Migrations == "" {
			s.state.Migrations = recipe.DatabaseMigrationsNone
		}
	case recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite:
		if !contains(compatibleFrameworks(s.state.Database), s.state.DatabaseFramework) {
			s.state.DatabaseFramework = recipe.DatabaseFrameworkDatabaseSQL
		}
		if s.state.Migrations == "" {
			s.state.Migrations = recipe.DatabaseMigrationsNone
		}
	default:
		s.state.DatabaseFramework = recipe.DatabaseFrameworkNone
		s.state.Migrations = recipe.DatabaseMigrationsNone
	}
}

func (s configureScreen) saveRecipe(confirmedOverwrite bool) error {
	if !s.state.options.Overwrite && !confirmedOverwrite {
		if _, err := os.Stat(s.state.options.RecipePath); err == nil {
			return fmt.Errorf("recipe file %q already exists; choose overwrite to replace it", s.state.options.RecipePath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check recipe output %q: %w", s.state.options.RecipePath, err)
		}
	}

	r := s.state.Recipe()
	if err := recipe.Save(s.state.options.RecipePath, r); err != nil {
		return err
	}
	s.state.saved = true
	return nil
}

func (s configureScreen) preview() string {
	r := s.state.Recipe()
	hasGoMod := false
	var builder strings.Builder
	builder.WriteString("Project:\n")
	builder.WriteString(fmt.Sprintf("  name: %s\n", r.Project.Name))
	builder.WriteString(fmt.Sprintf("  module: %s\n", r.Project.Module))
	if s.state.options.Mode == ConfigureWizardModeGeneration {
		builder.WriteString(fmt.Sprintf("  output directory: %s\n", s.state.OutputDirectory()))
	}
	builder.WriteString(fmt.Sprintf("  type: %s\n\n", r.Project.Type))

	builder.WriteString("Selected stack:\n")
	builder.WriteString(fmt.Sprintf("  go: %s\n", r.Go.Version))
	builder.WriteString(fmt.Sprintf("  layout: %s\n", r.Layout.Style))
	if r.Project.Type == recipe.ProjectTypeWeb {
		builder.WriteString(fmt.Sprintf("  server: %s\n", r.Server.Framework))
	}
	builder.WriteString(fmt.Sprintf("  configuration: %s\n", r.Configuration.Format))
	builder.WriteString(fmt.Sprintf("  logging: %s\n", r.Logging.Framework))
	builder.WriteString(fmt.Sprintf("  database/framework/migrations: %s/%s/%s\n", s.state.Database, r.Database.Framework, r.Database.Migrations))
	builder.WriteString(fmt.Sprintf("  task_scheduler: %s\n", r.TaskScheduler))
	builder.WriteString(fmt.Sprintf("  deployment: %s\n", joinEnabled(enabledLabel{r.Deployment.Docker, "docker"}, enabledLabel{r.Deployment.Compose, "compose"})))
	builder.WriteString(fmt.Sprintf("  ci: %s\n\n", joinEnabled(enabledLabel{r.CI.GitHubActions, "github_actions"}, enabledLabel{r.CI.GitLabCI, "gitlab_ci"}, enabledLabel{r.CI.AzurePipelines, "azure_pipelines"})))

	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		builder.WriteString("Plan error:\n")
		builder.WriteString("  " + err.Error() + "\n\n")
	} else {
		builder.WriteString("Components:\n")
		for _, current := range plan.Components {
			builder.WriteString("  - " + current.ID + "\n")
		}
		files, err := generator.RenderFileTargets(r, plan)
		if err != nil {
			builder.WriteString("\nFiles:\n")
			builder.WriteString("  " + err.Error() + "\n")
		}
		for _, file := range files {
			if file.Target == "go.mod" {
				hasGoMod = true
			}
		}
		builder.WriteString("\n")
	}

	if s.state.options.Mode == ConfigureWizardModeGeneration {
		builder.WriteString("Actions:\n")
		builder.WriteString("  - write files\n")
		if !s.state.options.SkipGoModTidy && hasGoMod {
			builder.WriteString("  - run go mod tidy\n")
		}
		builder.WriteString("\n")
		return builder.String()
	}

	if _, err := recipe.MarshalYAML(r); err != nil {
		builder.WriteString("Recipe YAML error: " + err.Error() + "\n")
	}
	return builder.String()
}

func (s configureScreen) configureInputs() configureScreen {
	switch s.step {
	case stepWelcome:
		startLabel := "Start configure wizard"
		if s.state.options.Mode == ConfigureWizardModeGeneration {
			startLabel = "Start new project wizard"
		}
		s.selectInput = selectWithCurrent("", []components.SelectOption{
			{Label: startLabel, Value: "start"},
			{Label: "Cancel", Value: "cancel"},
		}, "start")
	case stepProjectIdentity:
		s.nameInput = textinput.New()
		s.nameInput.Placeholder = "example-web"
		s.nameInput.SetValue(s.state.Name)
		s.nameInput.CharLimit = 80
		s.moduleInput = textinput.New()
		s.moduleInput.Placeholder = "example.com/example-web"
		s.moduleInput.SetValue(s.state.Module)
		s.moduleInput.CharLimit = 180
	case stepProjectType:
		s.selectInput = selectWithCurrent("Project type", []components.SelectOption{
			{Label: "Web service", Value: recipe.ProjectTypeWeb, Description: "HTTP API/application scaffold"},
			{Label: "CLI application", Value: recipe.ProjectTypeCLI, Description: "Command-line app scaffold"},
		}, s.state.ProjectType)
	case stepGoVersion:
		s.selectInput = selectWithCurrent("Go version", []components.SelectOption{
			{Label: "Go 1.26", Value: "1.26", Description: "Latest toolchain target"},
			{Label: "Go 1.25", Value: "1.25", Description: "Stable modern runtime"},
			{Label: "Go 1.24", Value: "1.24", Description: "Conservative compatibility"},
		}, s.state.GoVersion)
	case stepLayout:
		s.selectInput = selectWithCurrent("Layout style", []components.SelectOption{
			{Label: "Minimal", Value: recipe.LayoutStyleMinimal, Description: "Small apps, fewer packages"},
			{Label: "Layered", Value: recipe.LayoutStyleLayered, Description: "Config, handlers, services, repositories"},
		}, s.state.Layout)
	case stepServer:
		s.selectInput = selectWithCurrent("Server provider", []components.SelectOption{
			{Label: "net/http", Value: recipe.ServerFrameworkNetHTTP, Description: "Standard library, zero external router"},
			{Label: "chi", Value: recipe.ServerFrameworkChi, Description: "Lightweight router, clean APIs"},
			{Label: "gin", Value: recipe.ServerFrameworkGin, Description: "Popular web framework"},
			{Label: "echo", Value: recipe.ServerFrameworkEcho, Description: "Minimal framework with middleware"},
			{Label: "fiber", Value: recipe.ServerFrameworkFiber, Description: "Express-like API, fasthttp-based"},
		}, s.state.Server)
	case stepConfiguration:
		s.selectInput = selectWithCurrent("Configuration format", []components.SelectOption{
			{Label: "ENV", Value: recipe.ConfigurationFormatEnv, Description: "Environment variables only"},
			{Label: "YAML", Value: recipe.ConfigurationFormatYAML, Description: "Human-friendly config files"},
			{Label: "JSON", Value: recipe.ConfigurationFormatJSON, Description: "Machine-friendly config files"},
			{Label: "TOML", Value: recipe.ConfigurationFormatTOML, Description: "Compact structured config"},
		}, s.state.ConfigurationFormat)
	case stepLogging:
		s.selectInput = selectWithCurrent("Logging provider", []components.SelectOption{
			{Label: "slog", Value: recipe.LoggingFrameworkSlog, Description: "Standard library, simple and stable"},
			{Label: "zap", Value: recipe.LoggingFrameworkZap, Description: "High-performance structured logging"},
			{Label: "zerolog", Value: recipe.LoggingFrameworkZerolog, Description: "Minimal allocations, JSON-first"},
			{Label: "logrus", Value: recipe.LoggingFrameworkLogrus, Description: "Mature, widely known"},
		}, s.state.Logging)
	case stepDatabase:
		s.selectInput = selectWithCurrent("Database driver", []components.SelectOption{
			{Label: "None", Value: recipe.DatabaseDriverNone, Description: "No database integration"},
			{Label: "PostgreSQL", Value: recipe.DatabaseDriverPostgres, Description: "SQL database, production default"},
			{Label: "MySQL", Value: recipe.DatabaseDriverMySQL, Description: "SQL database, common web stack"},
			{Label: "SQLite", Value: recipe.DatabaseDriverSQLite, Description: "Local/file database"},
			{Label: "Redis", Value: recipe.DatabaseDriverRedis, Description: "NoSQL key-value database"},
			{Label: "MongoDB", Value: recipe.DatabaseDriverMongoDB, Description: "NoSQL document database"},
		}, s.state.Database)
	case stepDatabaseFramework:
		s.selectInput = selectWithCurrent("Database framework", frameworkOptions(s.state.Database), s.state.DatabaseFramework)
	case stepMigrations:
		s.selectInput = selectWithCurrent("Migration tool", []components.SelectOption{
			{Label: "None", Value: recipe.DatabaseMigrationsNone, Description: "Manage schema outside crego"},
			{Label: "goose", Value: recipe.DatabaseMigrationsGoose, Description: "SQL migrations with Go-friendly tooling"},
			{Label: "migrate", Value: recipe.DatabaseMigrationsMigrate, Description: "Broad database migration CLI"},
		}, s.state.Migrations)
	case stepObservability:
		s.multiInput = multiWithCurrent("Observability", []components.SelectOption{
			{Label: "Health endpoint", Value: "health", Description: "/healthz for basic liveness"},
			{Label: "Readiness endpoint", Value: "readiness", Description: "/readyz for dependency readiness"},
		}, map[string]bool{"health": s.state.Health, "readiness": s.state.Readiness})
	case stepTaskScheduler:
		s.selectInput = selectWithCurrent("Task scheduler", []components.SelectOption{
			{Label: "None", Value: recipe.TaskSchedulerNone, Description: "No background scheduler"},
			{Label: "gocron", Value: recipe.TaskSchedulerGocron, Description: "In-process scheduled jobs"},
		}, s.state.TaskScheduler)
	case stepDeployment:
		s.multiInput = multiWithCurrent("Deployment", []components.SelectOption{
			{Label: "Dockerfile", Value: "docker", Description: "Container image build file"},
			{Label: "Docker Compose", Value: "compose", Description: "Local service orchestration"},
		}, map[string]bool{"docker": s.state.Docker, "compose": s.state.Compose})
	case stepCI:
		s.multiInput = multiWithCurrent("CI/CD", []components.SelectOption{
			{Label: "GitHub Actions", Value: "github_actions", Description: "Workflow for GitHub-hosted repositories"},
			{Label: "GitLab CI", Value: "gitlab_ci", Description: "Pipeline for GitLab projects"},
			{Label: "Azure Pipelines", Value: "azure_pipelines", Description: "Pipeline for Azure DevOps"},
		}, map[string]bool{"github_actions": s.state.GitHubActions, "gitlab_ci": s.state.GitLabCI, "azure_pipelines": s.state.AzurePipelines})
	case stepPreview:
		options := []components.SelectOption{
			{Label: "Continue to save", Value: "continue"},
			{Label: "Back", Value: "back"},
			{Label: "Cancel", Value: "cancel"},
		}
		if s.state.options.Mode == ConfigureWizardModeGeneration {
			generate := components.SelectOption{Label: "Generate project", Value: "generate"}
			if !s.state.options.Force && outputDirectoryHasEntries(s.state.OutputDirectory()) {
				generate = components.SelectOption{Label: "Generate project and overwrite non-empty output directory", Value: "generate_overwrite"}
			}
			options = []components.SelectOption{
				generate,
				{Label: "Save recipe only", Value: "save"},
				{Label: "Back", Value: "back"},
				{Label: "Cancel", Value: "cancel"},
			}
		}
		s.selectInput = selectWithCurrent("", options, options[0].Value)
	case stepSave:
		options := []components.SelectOption{
			{Label: "Save recipe", Value: "save"},
			{Label: "Back to preview", Value: "back"},
			{Label: "Cancel", Value: "cancel"},
		}
		if !s.state.options.Overwrite && fileExists(s.state.options.RecipePath) {
			options = []components.SelectOption{
				{Label: "Back to preview", Value: "back"},
				{Label: "Overwrite existing recipe", Value: "overwrite"},
				{Label: "Cancel", Value: "cancel"},
			}
		}
		s.selectInput = selectWithCurrent("", options, options[0].Value)
	}
	return s
}

func (s configureScreen) title() string {
	if s.state.options.Mode == ConfigureWizardModeGeneration {
		return "crego new"
	}
	return "crego configure"
}

func (s configureScreen) stepTitle() string {
	switch s.step {
	case stepWelcome:
		return "Welcome"
	case stepProjectIdentity:
		return "Project identity"
	case stepGoVersion:
		return "Go version"
	case stepProjectType:
		return "Project type"
	case stepLayout:
		return "Layout style"
	case stepServer:
		return "Server provider"
	case stepConfiguration:
		return "Configuration"
	case stepLogging:
		return "Logging"
	case stepDatabase:
		return "Database"
	case stepDatabaseFramework:
		return "SQL framework"
	case stepMigrations:
		return "Migrations"
	case stepObservability:
		return "Observability"
	case stepTaskScheduler:
		return "Scheduler"
	case stepDeployment:
		return "Deployment"
	case stepCI:
		return "CI"
	case stepPreview:
		return "Preview"
	case stepSave:
		return "Save recipe"
	default:
		return ""
	}
}

func (s configureScreen) description() string {
	switch s.step {
	case stepWelcome:
		if s.state.options.Mode == ConfigureWizardModeGeneration {
			return "Create a Go project from guided choices."
		}
		return "Build a reusable crego.yaml from practical project choices."
	case stepProjectIdentity:
		return "Enter a safe project name and Go module path. Tab switches fields."
	case stepDatabaseFramework:
		return "Only frameworks compatible with the selected SQL database are shown."
	case stepPreview:
		if s.state.options.Mode == ConfigureWizardModeGeneration {
			return "Review the resolved generation plan before writing files."
		}
		return "Review normalized recipe YAML and the resolved generation plan before writing files."
	case stepSave:
		return fmt.Sprintf("Recipe path: %s", s.state.options.RecipePath)
	default:
		return ""
	}
}

func (s configureScreen) hint() string {
	switch s.step {
	case stepProjectIdentity:
		return HelpLine(HelpInput)
	case stepObservability, stepDeployment, stepCI:
		return HelpLine(HelpMulti)
	case stepPreview, stepSave:
		return HelpLine(HelpPreview)
	default:
		return HelpLine(HelpSelect)
	}
}

func (s configureScreen) identityView() string {
	namePrefix := "  "
	modulePrefix := "  "
	if s.focus == 0 {
		namePrefix = "-"
	} else {
		modulePrefix = "-"
	}
	return strings.Join([]string{
		s.styles.Description.Render("Project name"),
		namePrefix + s.nameInput.View(),
		s.styles.Description.Render("Go module"),
		modulePrefix + s.moduleInput.View(),
	}, "\n")
}

func (s configureScreen) progress() (int, int) {
	steps := s.visibleProgressSteps()
	total := len(steps)
	for index, step := range steps {
		if step == s.step {
			return index + 1, total
		}
	}
	return 0, total
}

func (s configureScreen) visibleProgressSteps() []configureStep {
	last := stepSave
	if s.state.options.Mode == ConfigureWizardModeGeneration {
		last = stepPreview
	}
	steps := make([]configureStep, 0, int(last))
	for step := stepProjectIdentity; step <= last; step++ {
		if shouldShowConfigureStep(step, s.state) {
			steps = append(steps, step)
		}
	}
	return steps
}

func (s configureScreen) sidebar() string {
	lines := make([]string, 0, len(s.visibleProgressSteps())+3)
	currentGroup := ""
	for _, step := range s.visibleProgressSteps() {
		group := stepGroup(step)
		if group != currentGroup {
			currentGroup = group
			lines = append(lines, s.styles.Description.Render(group))
		}
		marker := "○"
		style := s.styles.Description
		if step == s.step {
			marker = "●"
			style = s.styles.Selected
		} else if step < s.step {
			marker = "✓"
			style = s.styles.Success
		}
		lines = append(lines, style.Render(fmt.Sprintf("  %s %s", marker, progressStepLabel(step))))
	}
	return strings.Join(lines, "\n")
}

func stepGroup(step configureStep) string {
	switch step {
	case stepProjectIdentity, stepGoVersion, stepProjectType, stepLayout:
		return "Project"
	case stepServer, stepConfiguration, stepLogging:
		return "Stack"
	case stepDatabase, stepDatabaseFramework, stepMigrations:
		return "Data"
	case stepObservability, stepTaskScheduler:
		return "Runtime"
	case stepDeployment, stepCI:
		return "Delivery"
	default:
		return "Review"
	}
}

func progressStepLabel(step configureStep) string {
	switch step {
	case stepProjectIdentity:
		return "Identity"
	case stepGoVersion:
		return "Go version"
	case stepProjectType:
		return "Project type"
	case stepLayout:
		return "Layout"
	case stepServer:
		return "Server"
	case stepConfiguration:
		return "Config"
	case stepLogging:
		return "Logging"
	case stepDatabase:
		return "Database"
	case stepDatabaseFramework:
		return "SQL framework"
	case stepMigrations:
		return "Migrations"
	case stepObservability:
		return "Observability"
	case stepTaskScheduler:
		return "Scheduler"
	case stepDeployment:
		return "Deployment"
	case stepCI:
		return "CI"
	case stepPreview:
		return "Preview"
	case stepSave:
		return "Save"
	default:
		return "Start"
	}
}

func (s configureScreen) stackLine() string {
	values := selectedStackValues(s.state)
	if len(values) == 0 {
		return ""
	}
	return s.styles.Footer.Render("Stack: " + strings.Join(values, " · "))
}

func selectedStackValues(state *ConfigureWizardState) []string {
	if state == nil {
		return nil
	}
	values := make([]string, 0, 12)
	appendIf := func(value string) {
		value = strings.TrimSpace(value)
		if value != "" && value != recipe.DatabaseDriverNone && value != recipe.DatabaseFrameworkNone && value != recipe.DatabaseMigrationsNone && value != recipe.TaskSchedulerNone {
			values = append(values, value)
		}
	}
	appendIf(state.ProjectType)
	appendIf(state.Layout)
	if state.GoVersion != "" {
		appendIf("Go " + state.GoVersion)
	}
	if state.ProjectType == recipe.ProjectTypeWeb {
		appendIf(state.Server)
	}
	appendIf(state.ConfigurationFormat)
	appendIf(state.Logging)
	if isSQLDatabase(state.Database) && state.DatabaseFramework != "" && state.DatabaseFramework != recipe.DatabaseFrameworkNone {
		appendIf(state.Database + "/" + state.DatabaseFramework)
	} else {
		appendIf(state.Database)
	}
	appendIf(state.Migrations)
	appendIf(state.TaskScheduler)
	if state.Docker {
		appendIf("docker")
	}
	if state.Compose {
		appendIf("compose")
	}
	if state.GitHubActions {
		appendIf("github actions")
	}
	if state.GitLabCI {
		appendIf("gitlab ci")
	}
	if state.AzurePipelines {
		appendIf("azure pipelines")
	}
	return values
}

func (s configureScreen) livePreview() string {
	return s.filesPreview()
}

func (s configureScreen) filesPreview() string {
	r := s.state.Recipe()
	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		return "Plan error:\n  " + err.Error()
	}
	files, err := generator.RenderFileTargets(r, plan)
	if err != nil {
		return "Plan error:\n  " + err.Error()
	}

	lines := []string{"Files tree:\n"}
	for _, file := range files {
		lines = append(lines, "  "+file.Target)
	}
	return strings.Join(lines, "\n")
}

func (s configureScreen) layoutPreview() string {
	return layoutPreview(s.selectInput.Selected().Value)
}

func layoutPreview(style string) string {
	switch style {
	case recipe.LayoutStyleLayered:
		return `Preview
┌─────────────────────────────┐
│ cmd/app/main.go             │
│ internal/app/app.go         │
│ internal/config/config.go   │
│ internal/handler/http.go    │
│ internal/service/service.go │
└─────────────────────────────┘`
	default:
		return `Preview
┌─────────────────────────────┐
│ cmd/app/main.go             │
│ internal/app/app.go         │
│ internal/app/config.go      │
│ internal/app/http.go        │
│ internal/app/service.go     │
└─────────────────────────────┘`
	}
}

func nextConfigureStep(current configureStep, state *ConfigureWizardState) configureStep {
	for step := current + 1; step <= stepSave; step++ {
		if shouldShowConfigureStep(step, state) {
			return step
		}
	}
	return stepSave
}

func previousConfigureStep(current configureStep, state *ConfigureWizardState) configureStep {
	for step := current - 1; step >= stepWelcome; step-- {
		if shouldShowConfigureStep(step, state) {
			return step
		}
	}
	return stepWelcome
}

func shouldShowConfigureStep(step configureStep, state *ConfigureWizardState) bool {
	switch step {
	case stepServer, stepObservability:
		return state.ProjectType == recipe.ProjectTypeWeb
	case stepDatabaseFramework, stepMigrations:
		return isSQLDatabase(state.Database)
	default:
		return true
	}
}

func isConfigureBackKey(key tea.KeyMsg) bool {
	return key.String() == "ctrl+b" || key.String() == "esc"
}

func frameworkOptions(driver string) []components.SelectOption {
	options := compatibleFrameworks(driver)
	result := make([]components.SelectOption, 0, len(options))
	for _, option := range options {
		label := option
		description := "No database access layer"
		if option == recipe.DatabaseFrameworkDatabaseSQL {
			label = "database/sql"
			description = "Standard library SQL API"
		}
		if option == recipe.DatabaseFrameworkPGX {
			description = "PostgreSQL-native driver and toolkit"
		}
		if option == recipe.DatabaseFrameworkGORM {
			description = "Full-featured ORM"
		}
		result = append(result, components.SelectOption{Label: label, Value: option, Description: description})
	}
	return result
}

func compatibleFrameworks(driver string) []string {
	switch driver {
	case recipe.DatabaseDriverPostgres:
		return []string{recipe.DatabaseFrameworkPGX, recipe.DatabaseFrameworkDatabaseSQL, recipe.DatabaseFrameworkGORM}
	case recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite:
		return []string{recipe.DatabaseFrameworkDatabaseSQL, recipe.DatabaseFrameworkGORM}
	default:
		return []string{recipe.DatabaseFrameworkNone}
	}
}

func selectWithCurrent(title string, options []components.SelectOption, current string) components.Select {
	input := components.NewSelect(title, options)
	for index, option := range options {
		if option.Value == current {
			input.Cursor = index
			break
		}
	}
	return input
}

func multiWithCurrent(title string, options []components.SelectOption, selected map[string]bool) components.MultiSelect {
	input := components.NewMultiSelect(title, options)
	for key, value := range selected {
		input.Selected[key] = value
	}
	return input
}

func selectedSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func isSQLDatabase(driver string) bool {
	switch driver {
	case recipe.DatabaseDriverPostgres, recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite:
		return true
	default:
		return false
	}
}

func isNoSQLDatabase(driver string) bool {
	switch driver {
	case recipe.DatabaseDriverRedis, recipe.DatabaseDriverMongoDB:
		return true
	default:
		return false
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

type enabledLabel struct {
	enabled bool
	label   string
}

func joinEnabled(labels ...enabledLabel) string {
	values := make([]string, 0, len(labels))
	for _, item := range labels {
		if item.enabled {
			values = append(values, item.label)
		}
	}
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

func outputDirectoryHasEntries(path string) bool {
	entries, err := os.ReadDir(path)
	return err == nil && len(entries) > 0
}
