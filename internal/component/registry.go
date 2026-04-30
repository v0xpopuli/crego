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
	databaseConflicts := []string{IDDatabaseNone, IDDatabasePostgres, IDDatabaseMySQL, IDDatabaseSQLite}

	return []Component{
		{
			ID:          IDProjectWeb,
			Category:    CategoryProject,
			Name:        "Web project",
			Description: "HTTP service project scaffold.",
			Files:       []TemplateFile{{Source: "project/README.md.tmpl", Target: "README.md"}},
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
		serverComponent(IDServerNetHTTP, "net/http server", "HTTP server built with the Go standard library.", without(serverConflicts, IDServerNetHTTP)),
		serverComponent(IDServerChi, "Chi server", "HTTP server built with chi.", without(serverConflicts, IDServerChi)),
		serverComponent(IDServerGin, "Gin server", "HTTP server built with Gin.", without(serverConflicts, IDServerGin)),
		serverComponent(IDServerEcho, "Echo server", "HTTP server built with Echo.", without(serverConflicts, IDServerEcho)),
		serverComponent(IDServerFiber, "Fiber server", "HTTP server built with Fiber.", without(serverConflicts, IDServerFiber)),
		databaseComponent(IDDatabaseNone, "No database", "Project without a database integration.", without(databaseConflicts, IDDatabaseNone)),
		databaseComponent(IDDatabasePostgres, "Postgres database", "PostgreSQL database integration.", without(databaseConflicts, IDDatabasePostgres)),
		databaseComponent(IDDatabaseMySQL, "MySQL database", "MySQL database integration.", without(databaseConflicts, IDDatabaseMySQL)),
		databaseComponent(IDDatabaseSQLite, "SQLite database", "SQLite database integration.", without(databaseConflicts, IDDatabaseSQLite)),
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
		},
		{
			ID:          IDMigrationsMigrate,
			Category:    CategoryMigrations,
			Name:        "golang-migrate migrations",
			Description: "Database migrations through golang-migrate.",
			Conflicts:   []string{IDMigrationsNone, IDMigrationsGoose},
		},
		{
			ID:          IDLoggingSlog,
			Category:    CategoryLogging,
			Name:        "slog logging",
			Description: "Structured logging through the Go standard library slog package.",
		},
		{
			ID:          IDObservabilityHealth,
			Category:    CategoryObservability,
			Name:        "Health endpoint",
			Description: "Basic health endpoint.",
		},
		{
			ID:          IDObservabilityReadiness,
			Category:    CategoryObservability,
			Name:        "Readiness endpoint",
			Description: "Readiness endpoint for dependency checks.",
		},
		{
			ID:          IDDeploymentDocker,
			Category:    CategoryDeployment,
			Name:        "Docker",
			Description: "Dockerfile for containerized builds.",
		},
		{
			ID:          IDDeploymentCompose,
			Category:    CategoryDeployment,
			Name:        "Docker Compose",
			Description: "Docker Compose app service.",
			Requires:    []string{IDDeploymentDocker},
		},
		{
			ID:          IDCIGitHubActions,
			Category:    CategoryCI,
			Name:        "GitHub Actions",
			Description: "GitHub Actions workflow.",
		},
	}
}

func serverComponent(id string, name string, description string, conflicts []string) Component {
	return Component{
		ID:          id,
		Category:    CategoryServer,
		Name:        name,
		Description: description,
		Requires:    []string{IDProjectWeb},
		Conflicts:   conflicts,
	}
}

func databaseComponent(id string, name string, description string, conflicts []string) Component {
	return Component{
		ID:          id,
		Category:    CategoryDatabase,
		Name:        name,
		Description: description,
		Conflicts:   conflicts,
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
