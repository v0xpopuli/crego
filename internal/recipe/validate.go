package recipe

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	moduleFirstElementPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*[.][a-zA-Z0-9.-]+$`)
	modulePathSegmentPattern  = regexp.MustCompile(`^[A-Za-z0-9._~!$&'()*+,;=:@/-]+$`)
)

func Validate(r *Recipe) error {
	var problems []string
	if r == nil {
		return validationFailed("recipe is required")
	}

	resolved := *r
	Normalize(&resolved)
	ApplyDefaults(&resolved)
	r = &resolved

	if r.Version != VersionV1 {
		problems = append(problems, "version must be v1")
	}
	if r.Project.Name == "" {
		problems = append(problems, "project.name is required")
	}
	if r.Project.Module == "" {
		problems = append(problems, "project.module is required")
	} else if !looksLikeGoModulePath(r.Project.Module) {
		problems = append(problems, "project.module must look like a Go module path")
	}
	if r.Project.Type == "" {
		problems = append(problems, "project.type is required")
	} else {
		problems = appendEnumProblem(problems, "project.type", r.Project.Type, projectTypeValues())
	}

	validDatabaseDriver := containsString(databaseDriverValues(), r.Database.Driver)
	validDatabaseFramework := containsString(databaseFrameworkValues(), r.Database.Framework)

	problems = appendEnumProblem(problems, "layout.style", r.Layout.Style, layoutStyleValues())
	if r.Server.Framework != "" {
		problems = appendEnumProblem(problems, "server.framework", r.Server.Framework, serverFrameworkValues())
	}
	problems = appendEnumProblem(problems, "database.driver", r.Database.Driver, databaseDriverValues())
	problems = appendEnumProblem(problems, "database.framework", r.Database.Framework, databaseFrameworkValues())
	problems = appendEnumProblem(problems, "database.migrations", r.Database.Migrations, databaseMigrationValues())
	problems = appendEnumProblem(problems, "logging.provider", r.Logging.Provider, loggingProviderValues())
	problems = appendEnumProblem(problems, "logging.format", r.Logging.Format, loggingFormatValues())

	if validDatabaseDriver && validDatabaseFramework {
		problems = appendDatabaseCompatibilityProblems(problems, r.Database)
	}

	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}

	return nil
}

func appendDatabaseCompatibilityProblems(problems []string, database DatabaseConfig) []string {
	if database.Driver == DatabaseDriverNone {
		if database.Framework != "" && database.Framework != DatabaseFrameworkNone {
			problems = append(problems, "database.framework must be none when database.driver=none")
		}
		if database.Migrations == DatabaseMigrationsGoose || database.Migrations == DatabaseMigrationsMigrate {
			problems = append(problems, "database.migrations="+database.Migrations+" requires database.driver to be postgres, mysql, or sqlite")
		}
		return problems
	}

	if containsString(compatibleDatabaseFrameworks(database.Driver), database.Framework) {
		return problems
	}

	switch database.Framework {
	case DatabaseFrameworkPGX:
		problems = append(problems, "database.framework=pgx is only supported with database.driver=postgres")
	case DatabaseFrameworkNone:
		problems = append(problems, "database.framework=none is only supported with database.driver=none")
	default:
		problems = append(problems, "database.framework="+database.Framework+" is not supported with database.driver="+database.Driver)
	}

	return problems
}

func compatibleDatabaseFrameworks(driver string) []string {
	switch driver {
	case DatabaseDriverPostgres:
		return []string{DatabaseFrameworkPGX, DatabaseFrameworkDatabaseSQL, DatabaseFrameworkGORM}
	case DatabaseDriverMySQL, DatabaseDriverSQLite:
		return []string{DatabaseFrameworkDatabaseSQL, DatabaseFrameworkGORM}
	default:
		return []string{DatabaseFrameworkNone}
	}
}

func projectTypeValues() []string {
	return []string{ProjectTypeWeb, ProjectTypeCLI, ProjectTypeWorker, ProjectTypeLibrary}
}

func layoutStyleValues() []string {
	return []string{LayoutStyleMinimal, LayoutStyleLayered, LayoutStyleDomain}
}

func serverFrameworkValues() []string {
	return []string{ServerFrameworkNetHTTP, ServerFrameworkChi, ServerFrameworkGin, ServerFrameworkEcho, ServerFrameworkFiber}
}

func databaseDriverValues() []string {
	return []string{DatabaseDriverNone, DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite}
}

func databaseFrameworkValues() []string {
	return []string{DatabaseFrameworkNone, DatabaseFrameworkPGX, DatabaseFrameworkDatabaseSQL, DatabaseFrameworkGORM}
}

func databaseMigrationValues() []string {
	return []string{DatabaseMigrationsNone, DatabaseMigrationsGoose, DatabaseMigrationsMigrate}
}

func loggingProviderValues() []string {
	return []string{LoggingProviderSlog, LoggingProviderZap, LoggingProviderZerolog}
}

func loggingFormatValues() []string {
	return []string{LoggingFormatText, LoggingFormatJSON}
}

func appendEnumProblem(problems []string, field string, value string, allowed []string) []string {
	if containsString(allowed, value) {
		return problems
	}

	return append(problems, field+"="+value+" is invalid; allowed values: "+strings.Join(allowed, ", "))
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func looksLikeGoModulePath(module string) bool {
	if strings.ContainsAny(module, " \t\r\n") {
		return false
	}
	if parsed, err := url.Parse(module); err == nil && parsed.Scheme != "" {
		return false
	}
	parts := strings.Split(module, "/")
	if len(parts) < 2 {
		return false
	}
	if !moduleFirstElementPattern.MatchString(parts[0]) {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || !modulePathSegmentPattern.MatchString(part) {
			return false
		}
	}
	return true
}
