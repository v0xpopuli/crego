package generator

import (
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
)

var (
	projectTypeComponentIDs = map[string]string{
		recipe.ProjectTypeWeb: component.IDProjectWeb,
		recipe.ProjectTypeCLI: component.IDProjectCLI,
	}

	layoutStyleComponentIDs = map[string]string{
		recipe.LayoutStyleMinimal: component.IDLayoutMinimal,
		recipe.LayoutStyleLayered: component.IDLayoutLayered,
	}

	serverFrameworkComponentIDs = map[string]string{
		recipe.ServerFrameworkNetHTTP: component.IDServerNetHTTP,
		recipe.ServerFrameworkChi:     component.IDServerChi,
		recipe.ServerFrameworkGin:     component.IDServerGin,
		recipe.ServerFrameworkEcho:    component.IDServerEcho,
		recipe.ServerFrameworkFiber:   component.IDServerFiber,
	}

	configurationFormatComponentIDs = map[string]string{
		recipe.ConfigurationFormatEnv:  component.IDConfigurationEnv,
		recipe.ConfigurationFormatYAML: component.IDConfigurationYAML,
		recipe.ConfigurationFormatJSON: component.IDConfigurationJSON,
		recipe.ConfigurationFormatTOML: component.IDConfigurationTOML,
	}

	databaseDriverComponentIDs = map[string]string{
		recipe.DatabaseDriverNone:     component.IDDatabaseNone,
		recipe.DatabaseDriverPostgres: component.IDDatabasePostgres,
		recipe.DatabaseDriverMySQL:    component.IDDatabaseMySQL,
		recipe.DatabaseDriverSQLite:   component.IDDatabaseSQLite,
		recipe.DatabaseDriverRedis:    component.IDDatabaseRedis,
		recipe.DatabaseDriverMongoDB:  component.IDDatabaseMongoDB,
	}

	databaseFrameworkComponentIDs = map[string]string{
		recipe.DatabaseFrameworkPGX:         component.IDDatabaseFrameworkPGX,
		recipe.DatabaseFrameworkDatabaseSQL: component.IDDatabaseFrameworkSQL,
		recipe.DatabaseFrameworkGORM:        component.IDDatabaseFrameworkGORM,
	}

	databaseMigrationsComponentIDs = map[string]string{
		recipe.DatabaseMigrationsNone:    component.IDMigrationsNone,
		recipe.DatabaseMigrationsGoose:   component.IDMigrationsGoose,
		recipe.DatabaseMigrationsMigrate: component.IDMigrationsMigrate,
	}

	loggingFrameworkComponentIDs = map[string]string{
		recipe.LoggingFrameworkSlog:    component.IDLoggingSlog,
		recipe.LoggingFrameworkZap:     component.IDLoggingZap,
		recipe.LoggingFrameworkZerolog: component.IDLoggingZerolog,
		recipe.LoggingFrameworkLogrus:  component.IDLoggingLogrus,
	}

	databaseBackedComponents = []string{
		component.IDDatabasePostgres,
		component.IDDatabaseMySQL,
		component.IDDatabaseSQLite,
	}
)

func Resolve(registry *component.Registry, source *recipe.Recipe) (*Plan, error) {
	if registry == nil {
		registry = component.NewRegistry()
	}
	if source == nil {
		return nil, recipe.Validate(nil)
	}

	resolved := *source
	recipe.Normalize(&resolved)
	recipe.ApplyDefaults(&resolved)
	if err := recipe.Validate(&resolved); err != nil {
		return nil, err
	}

	selectedIDs := mappedComponentIDs(&resolved)
	selected := make(map[string]struct{}, len(selectedIDs))
	visiting := make(map[string]bool)
	for _, id := range selectedIDs {
		if err := expandComponent(registry, selected, visiting, id, ""); err != nil {
			return nil, err
		}
	}

	if err := validateOneOfRequirements(selected); err != nil {
		return nil, err
	}
	if err := validateConflicts(registry, selected); err != nil {
		return nil, err
	}

	return buildPlan(registry, selected, &resolved), nil
}

func mappedComponentIDs(r *recipe.Recipe) []string {
	ids := make([]string, 0, 12)
	add := func(id string) {
		if id != "" {
			ids = append(ids, id)
		}
	}

	add(mappedID(projectTypeComponentIDs, component.CategoryProject, r.Project.Type))
	add(mappedID(layoutStyleComponentIDs, component.CategoryLayout, r.Layout.Style))

	if r.Project.Type == recipe.ProjectTypeWeb && r.Server.Framework != "" {
		add(mappedID(serverFrameworkComponentIDs, component.CategoryServer, r.Server.Framework))
	}

	if r.Project.Type == recipe.ProjectTypeWeb && r.Configuration.Format != "" {
		add(mappedID(configurationFormatComponentIDs, component.CategoryConfiguration, r.Configuration.Format))
	}

	for _, driver := range recipe.DatabaseDrivers(r.Database) {
		add(mappedID(databaseDriverComponentIDs, component.CategoryDatabase, driver))
	}
	if r.Database.Framework != "" && r.Database.Framework != recipe.DatabaseFrameworkNone {
		add(mappedID(databaseFrameworkComponentIDs, component.CategoryDatabaseFramework, r.Database.Framework))
	}
	if r.Database.Migrations != "" && r.Database.Migrations != recipe.DatabaseMigrationsNone {
		add(mappedID(databaseMigrationsComponentIDs, component.CategoryMigrations, r.Database.Migrations))
	} else if len(recipe.DatabaseDrivers(r.Database)) == 1 && recipe.DatabaseDrivers(r.Database)[0] == recipe.DatabaseDriverNone {
		add(mappedID(databaseMigrationsComponentIDs, component.CategoryMigrations, recipe.DatabaseMigrationsNone))
	}

	if r.Logging.Framework != "" {
		add(mappedID(loggingFrameworkComponentIDs, component.CategoryLogging, r.Logging.Framework))
	}
	if r.Observability.Health {
		add(component.IDObservabilityHealth)
	}
	if r.Observability.Readiness {
		add(component.IDObservabilityReadiness)
	}
	if r.Deployment.Docker {
		add(component.IDDeploymentDocker)
	}
	if r.Deployment.Compose {
		add(component.IDDeploymentCompose)
	}
	if r.CI.GitHubActions {
		add(component.IDCIGitHubActions)
	}
	if r.CI.GitLabCI {
		add(component.IDCIGitLabCI)
	}
	if r.CI.AzurePipelines {
		add(component.IDCIAzurePipelines)
	}

	return ids
}

func mappedID(known map[string]string, category string, value string) string {
	if id, ok := known[value]; ok {
		return id
	}
	return category + "." + value
}

func expandComponent(registry *component.Registry, selected map[string]struct{}, visiting map[string]bool, id string, parentID string) error {
	if _, ok := selected[id]; ok {
		return nil
	}

	current, ok := registry.Get(id)
	if !ok {
		if parentID != "" {
			return &component.MissingDependencyError{ComponentID: parentID, DependencyID: id}
		}
		return &component.UnknownComponentError{ID: id}
	}

	if visiting[id] {
		return &component.DependencyCycleError{ComponentID: id}
	}

	visiting[id] = true
	for _, requiredID := range current.Requires {
		if err := expandComponent(registry, selected, visiting, requiredID, current.ID); err != nil {
			return err
		}
	}
	visiting[id] = false

	selected[id] = struct{}{}
	return nil
}

func validateOneOfRequirements(selected map[string]struct{}) error {
	for _, id := range []string{component.IDDatabaseFrameworkSQL, component.IDDatabaseFrameworkGORM, component.IDMigrationsGoose, component.IDMigrationsMigrate} {
		if !isSelected(selected, id) {
			continue
		}
		if hasAnySelected(selected, databaseBackedComponents) {
			continue
		}
		return &component.MissingRequirementError{ComponentID: id, Requires: databaseBackedComponents}
	}

	return nil
}

func validateConflicts(registry *component.Registry, selected map[string]struct{}) error {
	for _, current := range registry.List() {
		if !isSelected(selected, current.ID) {
			continue
		}
		for _, conflictID := range current.Conflicts {
			if isSelected(selected, conflictID) {
				return &component.ConflictError{ComponentID: current.ID, ConflictID: conflictID}
			}
		}
	}

	return nil
}

func buildPlan(registry *component.Registry, selected map[string]struct{}, r *recipe.Recipe) *Plan {
	plan := &Plan{}
	seenFiles := make(map[string]struct{})
	seenGoModules := make(map[string]struct{})
	seenHooks := make(map[string]struct{})

	for _, current := range registry.List() {
		if !isSelected(selected, current.ID) {
			continue
		}

		plan.Components = append(plan.Components, current)
		for _, file := range current.Files {
			key := file.Source + "\x00" + file.Target
			if _, ok := seenFiles[key]; ok {
				continue
			}
			seenFiles[key] = struct{}{}
			plan.Files = append(plan.Files, file)
		}
		for _, module := range current.GoModules {
			key := module.Path + "\x00" + module.Version
			if _, ok := seenGoModules[key]; ok {
				continue
			}
			seenGoModules[key] = struct{}{}
			plan.GoModules = append(plan.GoModules, module)
		}
		for _, hook := range current.Hooks {
			key := hook.Name
			if _, ok := seenHooks[key]; ok {
				continue
			}
			seenHooks[key] = struct{}{}
			plan.Hooks = append(plan.Hooks, hook)
		}
	}
	for _, module := range databaseGoModules(r) {
		key := module.Path + "\x00" + module.Version
		if _, ok := seenGoModules[key]; ok {
			continue
		}
		seenGoModules[key] = struct{}{}
		plan.GoModules = append(plan.GoModules, module)
	}

	return plan
}

func databaseGoModules(r *recipe.Recipe) []component.GoModule {
	if r == nil {
		return nil
	}

	modules := make([]component.GoModule, 0, 6)
	add := func(path string, version string) {
		modules = append(modules, component.GoModule{Path: path, Version: version})
	}

	for _, driver := range recipe.DatabaseDrivers(r.Database) {
		switch driver {
		case recipe.DatabaseDriverPostgres:
			switch r.Database.Framework {
			case recipe.DatabaseFrameworkPGX, recipe.DatabaseFrameworkDatabaseSQL:
				add("github.com/jackc/pgx/v5", "v5.9.2")
			case recipe.DatabaseFrameworkGORM:
				add("gorm.io/gorm", "v1.31.1")
				add("gorm.io/driver/postgres", "v1.6.0")
			}
		case recipe.DatabaseDriverMySQL:
			switch r.Database.Framework {
			case recipe.DatabaseFrameworkDatabaseSQL:
				add("github.com/go-sql-driver/mysql", "v1.9.3")
			case recipe.DatabaseFrameworkGORM:
				add("gorm.io/gorm", "v1.31.1")
				add("gorm.io/driver/mysql", "v1.6.0")
			}
		case recipe.DatabaseDriverSQLite:
			switch r.Database.Framework {
			case recipe.DatabaseFrameworkDatabaseSQL:
				add("modernc.org/sqlite", "v1.48.2")
			case recipe.DatabaseFrameworkGORM:
				add("gorm.io/gorm", "v1.31.1")
				add("gorm.io/driver/sqlite", "v1.6.0")
			}
		case recipe.DatabaseDriverRedis:
			add("github.com/redis/go-redis/v9", "v9.17.2")
		case recipe.DatabaseDriverMongoDB:
			add("go.mongodb.org/mongo-driver/v2", "v2.5.1")
		}
	}

	switch r.Database.Migrations {
	case recipe.DatabaseMigrationsGoose:
		add("github.com/pressly/goose/v3", "v3.27.1")
	case recipe.DatabaseMigrationsMigrate:
		add("github.com/golang-migrate/migrate/v4", "v4.19.1")
	}

	return modules
}

func isSelected(selected map[string]struct{}, id string) bool {
	_, ok := selected[id]
	return ok
}

func hasAnySelected(selected map[string]struct{}, ids []string) bool {
	for _, id := range ids {
		if isSelected(selected, id) {
			return true
		}
	}
	return false
}
