package component

type Registry struct {
	components []Component
	byID       map[string]Component
}

func NewRegistry() *Registry {
	return NewRegistryFromComponents(defaultComponents())
}

func NewRegistryFromComponents(components []Component) *Registry {
	registry := &Registry{
		components: make([]Component, 0, len(components)),
		byID:       make(map[string]Component, len(components)),
	}

	for _, component := range components {
		cloned := cloneComponent(component)
		registry.components = append(registry.components, cloned)
		registry.byID[cloned.ID] = cloned
	}

	return registry
}

func (r *Registry) List() []Component {
	if r == nil {
		return nil
	}

	components := make([]Component, 0, len(r.components))
	for _, component := range r.components {
		components = append(components, cloneComponent(component))
	}
	return components
}

func (r *Registry) Get(id string) (Component, bool) {
	if r == nil {
		return Component{}, false
	}

	component, ok := r.byID[id]
	if !ok {
		return Component{}, false
	}
	return cloneComponent(component), true
}

func defaultComponents() []Component {
	serverConflicts := []string{IDServerNetHTTP, IDServerChi, IDServerGin, IDServerEcho, IDServerFiber}
	configurationConflicts := []string{IDConfigurationEnv, IDConfigurationYAML, IDConfigurationJSON, IDConfigurationTOML}
	databaseComponents := []string{IDDatabaseNone, IDDatabasePostgres, IDDatabaseMySQL, IDDatabaseSQLite, IDDatabaseRedis, IDDatabaseMongoDB}
	loggingConflicts := []string{IDLoggingSlog, IDLoggingZap, IDLoggingZerolog, IDLoggingLogrus}

	return []Component{
		{
			ID:          IDProjectWeb,
			Category:    CategoryProject,
			Name:        "Web project",
			Description: "HTTP service project scaffold.",
			Files: []TemplateFile{
				{Source: "web/go.mod.tmpl", Target: "go.mod"},
				{Source: "web/README.md.tmpl", Target: "README.md"},
				{Source: "web/Makefile.tmpl", Target: "Makefile"},
				{Source: "web/main.go.tmpl", Target: "cmd/{{ .ProjectName }}/main.go"},
				{Source: "web/app.go.tmpl", Target: "internal/app/app.go"},
				{Source: "web/config.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/config.go{{ else }}internal/config/config.go{{ end }}"},
				{Source: "web/logger.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/logger.go{{ else }}internal/logging/logger.go{{ end }}"},
				{Source: "web/server.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/server.go{{ else }}internal/server/server.go{{ end }}"},
				{Source: "web/routes.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/routes.go{{ else }}internal/server/routes.go{{ end }}"},
				{Source: "web/readiness.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/readiness.go{{ else }}internal/server/readiness.go{{ end }}"},
				{Source: "web/request_id.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/request_id.go{{ else }}internal/server/middleware/request_id.go{{ end }}"},
				{Source: "web/logging_middleware.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/logging.go{{ else }}internal/server/middleware/logging.go{{ end }}"},
				{Source: "web/recover.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/recover.go{{ else }}internal/server/middleware/recover.go{{ end }}"},
			},
		},
		{
			ID:          IDProjectCLI,
			Category:    CategoryProject,
			Name:        "CLI project",
			Description: "Command-line application project scaffold.",
			Files:       []TemplateFile{{Source: "project/README.md.tmpl", Target: "README.md"}},
		},
		{
			ID:          IDLayoutMinimal,
			Category:    CategoryLayout,
			Name:        "Minimal layout",
			Description: "Small project layout with minimal package structure.",
		},
		{
			ID:          IDLayoutLayered,
			Category:    CategoryLayout,
			Name:        "Layered layout",
			Description: "Layered project layout for separated application concerns.",
		},
		serverComponent(IDServerNetHTTP, "net/http server", "HTTP server built with the Go standard library.", without(serverConflicts, IDServerNetHTTP), nil),
		serverComponent(IDServerChi, "Chi server", "HTTP server built with chi.", without(serverConflicts, IDServerChi), []GoModule{{Path: "github.com/go-chi/chi/v5", Version: "v5.2.5"}}),
		serverComponent(IDServerGin, "Gin server", "HTTP server built with Gin.", without(serverConflicts, IDServerGin), []GoModule{{Path: "github.com/gin-gonic/gin", Version: "v1.12.0"}}),
		serverComponent(IDServerEcho, "Echo server", "HTTP server built with Echo.", without(serverConflicts, IDServerEcho), []GoModule{{Path: "github.com/labstack/echo/v4", Version: "v4.15.1"}}),
		serverComponent(IDServerFiber, "Fiber server", "HTTP server built with Fiber.", without(serverConflicts, IDServerFiber), []GoModule{{Path: "github.com/gofiber/fiber/v2", Version: "v2.52.12"}}),
		configurationComponent(IDConfigurationEnv, "Environment configuration", "Configuration loaded from environment variables.", without(configurationConflicts, IDConfigurationEnv), nil, nil),
		configurationComponent(IDConfigurationYAML, "YAML configuration", "Configuration loaded from YAML files.", without(configurationConflicts, IDConfigurationYAML), []TemplateFile{{Source: "web/config.yaml.tmpl", Target: "configs/config.yaml"}}, []GoModule{{Path: "gopkg.in/yaml.v3", Version: "v3.0.1"}}),
		configurationComponent(IDConfigurationJSON, "JSON configuration", "Configuration loaded from JSON files.", without(configurationConflicts, IDConfigurationJSON), []TemplateFile{{Source: "web/config.json.tmpl", Target: "configs/config.json"}}, nil),
		configurationComponent(IDConfigurationTOML, "TOML configuration", "Configuration loaded from TOML files.", without(configurationConflicts, IDConfigurationTOML), []TemplateFile{{Source: "web/config.toml.tmpl", Target: "configs/config.toml"}}, []GoModule{{Path: "github.com/pelletier/go-toml/v2", Version: "v2.3.0"}}),
		databaseComponent(IDDatabaseNone, "No database", "Project without a database integration.", without(databaseComponents, IDDatabaseNone), nil),
		databaseComponent(IDDatabasePostgres, "Postgres database", "PostgreSQL database integration.", []string{IDDatabaseNone}, []TemplateFile{
			{Source: "web/postgres.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/postgres.go{{ else }}internal/database/postgres.go{{ end }}"},
		}),
		databaseComponent(IDDatabaseMySQL, "MySQL database", "MySQL database integration.", []string{IDDatabaseNone}, []TemplateFile{
			{Source: "web/mysql.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/mysql.go{{ else }}internal/database/mysql.go{{ end }}"},
		}),
		databaseComponent(IDDatabaseSQLite, "SQLite database", "SQLite database integration.", []string{IDDatabaseNone}, []TemplateFile{
			{Source: "web/sqlite.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/sqlite.go{{ else }}internal/database/sqlite.go{{ end }}"},
		}),
		databaseComponent(IDDatabaseRedis, "Redis database", "Redis database integration.", []string{IDDatabaseNone}, []TemplateFile{
			{Source: "web/redis.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/redis.go{{ else }}internal/database/redis.go{{ end }}"},
		}),
		databaseComponent(IDDatabaseMongoDB, "MongoDB database", "MongoDB database integration.", []string{IDDatabaseNone}, []TemplateFile{
			{Source: "web/mongodb.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/mongodb.go{{ else }}internal/database/mongodb.go{{ end }}"},
		}),
		{
			ID:          IDDatabaseFrameworkPGX,
			Category:    CategoryDatabaseFramework,
			Name:        "pgx",
			Description: "PostgreSQL access through pgx.",
			Requires:    []string{IDDatabasePostgres},
		},
		{
			ID:          IDDatabaseFrameworkSQL,
			Category:    CategoryDatabaseFramework,
			Name:        "database/sql",
			Description: "Database access through the Go database/sql package.",
		},
		{
			ID:          IDDatabaseFrameworkGORM,
			Category:    CategoryDatabaseFramework,
			Name:        "GORM",
			Description: "Database access through GORM.",
		},
		{
			ID:          IDMigrationsNone,
			Category:    CategoryMigrations,
			Name:        "No migrations",
			Description: "Project without a database migration tool.",
			Conflicts:   []string{IDMigrationsGoose, IDMigrationsMigrate},
		},
		{
			ID:          IDMigrationsGoose,
			Category:    CategoryMigrations,
			Name:        "Goose migrations",
			Description: "Database migrations through goose.",
			Conflicts:   []string{IDMigrationsNone, IDMigrationsMigrate},
			Files: []TemplateFile{
				{Source: "web/migrations.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/migrations.go{{ else }}internal/database/migrations.go{{ end }}"},
				{Source: "web/migration_goose.sql.tmpl", Target: "scripts/migrations/000001_init.sql"},
			},
		},
		{
			ID:          IDMigrationsMigrate,
			Category:    CategoryMigrations,
			Name:        "golang-migrate migrations",
			Description: "Database migrations through golang-migrate.",
			Conflicts:   []string{IDMigrationsNone, IDMigrationsGoose},
			Files: []TemplateFile{
				{Source: "web/migrations.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/migrations.go{{ else }}internal/database/migrations.go{{ end }}"},
				{Source: "web/migration_migrate_up.sql.tmpl", Target: "scripts/migrations/000001_init.up.sql"},
				{Source: "web/migration_migrate_down.sql.tmpl", Target: "scripts/migrations/000001_init.down.sql"},
			},
		},
		{
			ID:          IDLoggingSlog,
			Category:    CategoryLogging,
			Name:        "slog logging",
			Description: "Structured logging through the Go standard library slog package.",
			Conflicts:   without(loggingConflicts, IDLoggingSlog),
		},
		{
			ID:          IDLoggingZap,
			Category:    CategoryLogging,
			Name:        "zap logging",
			Description: "Structured logging through zap.",
			Conflicts:   without(loggingConflicts, IDLoggingZap),
			GoModules:   []GoModule{{Path: "go.uber.org/zap", Version: "v1.27.1"}},
		},
		{
			ID:          IDLoggingZerolog,
			Category:    CategoryLogging,
			Name:        "zerolog logging",
			Description: "Structured logging through zerolog.",
			Conflicts:   without(loggingConflicts, IDLoggingZerolog),
			GoModules:   []GoModule{{Path: "github.com/rs/zerolog", Version: "v1.35.0"}},
		},
		{
			ID:          IDLoggingLogrus,
			Category:    CategoryLogging,
			Name:        "logrus logging",
			Description: "Structured logging through logrus.",
			Conflicts:   without(loggingConflicts, IDLoggingLogrus),
			GoModules:   []GoModule{{Path: "github.com/sirupsen/logrus", Version: "v1.9.4"}},
		},
		{
			ID:          IDObservabilityHealth,
			Category:    CategoryObservability,
			Name:        "Health endpoint",
			Description: "Basic health endpoint.",
			Files:       []TemplateFile{{Source: "web/health.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/health.go{{ else }}internal/server/handler/health.go{{ end }}"}},
		},
		{
			ID:          IDObservabilityReadiness,
			Category:    CategoryObservability,
			Name:        "Readiness endpoint",
			Description: "Readiness endpoint for dependency checks.",
			Files:       []TemplateFile{{Source: "web/ready.go.tmpl", Target: "{{ if eq .Recipe.Layout.Style \"minimal\" }}internal/app/ready.go{{ else }}internal/server/ready.go{{ end }}"}},
		},
		{
			ID:          IDDeploymentDocker,
			Category:    CategoryDeployment,
			Name:        "Docker",
			Description: "Dockerfile for containerized builds.",
			Files: []TemplateFile{
				{Source: "web/Dockerfile.tmpl", Target: "deployments/Dockerfile"},
				{Source: "web/dockerignore.tmpl", Target: "deployments/.dockerignore"},
			},
		},
		{
			ID:          IDDeploymentCompose,
			Category:    CategoryDeployment,
			Name:        "Docker Compose",
			Description: "Docker Compose app service.",
			Requires:    []string{IDDeploymentDocker},
			Files:       []TemplateFile{{Source: "web/docker-compose.yml.tmpl", Target: "deployments/docker-compose.yml"}},
		},
		{
			ID:          IDCIGitHubActions,
			Category:    CategoryCI,
			Name:        "GitHub Actions",
			Description: "GitHub Actions workflow.",
			Files:       []TemplateFile{{Source: "web/github-actions-test.yml.tmpl", Target: ".github/workflows/test.yml"}},
		},
		{
			ID:          IDCIGitLabCI,
			Category:    CategoryCI,
			Name:        "GitLab CI",
			Description: "GitLab CI pipeline.",
			Files:       []TemplateFile{{Source: "web/gitlab-ci.yml.tmpl", Target: ".gitlab-ci.yml"}},
		},
	}
}

func serverComponent(id string, name string, description string, conflicts []string, modules []GoModule) Component {
	return Component{
		ID:          id,
		Category:    CategoryServer,
		Name:        name,
		Description: description,
		Requires:    []string{IDProjectWeb},
		Conflicts:   conflicts,
		GoModules:   modules,
	}
}

func configurationComponent(id string, name string, description string, conflicts []string, files []TemplateFile, modules []GoModule) Component {
	return Component{
		ID:          id,
		Category:    CategoryConfiguration,
		Name:        name,
		Description: description,
		Conflicts:   conflicts,
		Files:       files,
		GoModules:   modules,
	}
}

func databaseComponent(id string, name string, description string, conflicts []string, files []TemplateFile) Component {
	return Component{
		ID:          id,
		Category:    CategoryDatabase,
		Name:        name,
		Description: description,
		Conflicts:   conflicts,
		Files:       files,
	}
}

func without(values []string, excluded string) []string {
	result := make([]string, 0, len(values)-1)
	for _, value := range values {
		if value != excluded {
			result = append(result, value)
		}
	}
	return result
}

func cloneComponent(component Component) Component {
	component.Requires = append([]string(nil), component.Requires...)
	component.Conflicts = append([]string(nil), component.Conflicts...)
	component.Files = append([]TemplateFile(nil), component.Files...)
	component.GoModules = append([]GoModule(nil), component.GoModules...)
	component.Hooks = append([]Hook(nil), component.Hooks...)
	return component
}
