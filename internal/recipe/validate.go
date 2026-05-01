package recipe

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	moduleFirstElementPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*[.][a-zA-Z0-9.-]+$`)
	modulePathSegmentPattern  = regexp.MustCompile(`^[A-Za-z0-9._~!$&'()*+,;=:@/-]+$`)
	projectNamePattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
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
	} else if !isSafeProjectName(r.Project.Name) {
		problems = append(problems, "project.name must be a safe single path segment")
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

	drivers := DatabaseDrivers(r.Database)
	validDatabaseDriver := true
	for _, driver := range drivers {
		if !containsString(databaseDriverValues(), driver) {
			validDatabaseDriver = false
			break
		}
	}
	validDatabaseFramework := containsString(databaseFrameworkValues(), r.Database.Framework)

	problems = appendEnumProblem(problems, "layout.style", r.Layout.Style, layoutStyleValues())
	if r.Server.Framework != "" {
		problems = appendEnumProblem(problems, "server.framework", r.Server.Framework, serverFrameworkValues())
	}
	problems = appendEnumProblem(problems, "configuration.format", r.Configuration.Format, configurationFormatValues())
	for _, driver := range drivers {
		problems = appendEnumProblem(problems, "database.driver", driver, databaseDriverValues())
	}
	if r.Database.SQL != "" {
		problems = appendEnumProblem(problems, "database.sql", r.Database.SQL, databaseSQLDriverValues())
	}
	for _, driver := range r.Database.NoSQL {
		problems = appendEnumProblem(problems, "database.nosql", driver, databaseNoSQLDriverValues())
	}
	problems = appendEnumProblem(problems, "database.framework", r.Database.Framework, databaseFrameworkValues())
	problems = appendEnumProblem(problems, "database.migrations", r.Database.Migrations, databaseMigrationValues())
	problems = appendEnumProblem(problems, "logging.framework", r.Logging.Framework, loggingFrameworkValues())
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
	drivers := DatabaseDrivers(database)
	if len(drivers) == 0 {
		drivers = []string{DatabaseDriverNone}
	}
	if len(drivers) > 1 && containsString(drivers, DatabaseDriverNone) {
		problems = append(problems, "database.driver=none cannot be combined with other database drivers")
	}
	sqlDrivers := sqlDatabaseDrivers(drivers)
	if len(sqlDrivers) > 1 {
		problems = append(problems, "database.sql supports only one SQL database driver")
	}
	if len(sqlDrivers) == 0 && (database.Framework != "" && database.Framework != DatabaseFrameworkNone) {
		problems = append(problems, "database.framework="+database.Framework+" is only supported with SQL database drivers")
	}
	if len(sqlDrivers) == 0 && (database.Migrations == DatabaseMigrationsGoose || database.Migrations == DatabaseMigrationsMigrate) {
		problems = append(problems, "database.migrations="+database.Migrations+" is only supported with SQL database drivers")
	}

	if len(drivers) == 1 && drivers[0] == DatabaseDriverNone {
		if database.Framework != "" && database.Framework != DatabaseFrameworkNone {
			problems = append(problems, "database.framework must be none when database.driver=none")
		}
		if database.Migrations == DatabaseMigrationsGoose || database.Migrations == DatabaseMigrationsMigrate {
			problems = append(problems, "database.migrations="+database.Migrations+" requires database.driver to be postgres, mysql, or sqlite")
		}
		return problems
	}

	if len(sqlDrivers) == 0 {
		return problems
	}

	for _, sqlDriver := range sqlDrivers {
		if !containsString(compatibleDatabaseMigrations(sqlDriver), database.Migrations) {
			problems = append(problems, "database.migrations="+database.Migrations+" is not supported with database.driver="+sqlDriver)
		}

		if containsString(compatibleDatabaseFrameworks(sqlDriver), database.Framework) {
			continue
		}

		switch database.Framework {
		case DatabaseFrameworkPGX:
			problems = append(problems, "database.framework=pgx is only supported with database.driver=postgres")
		case DatabaseFrameworkNone:
			problems = append(problems, "database.framework=none is only supported with database.driver=none")
		default:
			problems = append(problems, "database.framework="+database.Framework+" is not supported with database.driver="+sqlDriver)
		}
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

func compatibleDatabaseMigrations(driver string) []string {
	switch driver {
	case DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite:
		return []string{DatabaseMigrationsNone, DatabaseMigrationsGoose, DatabaseMigrationsMigrate}
	default:
		return []string{DatabaseMigrationsNone}
	}
}

func isSQLDatabaseDriver(driver string) bool {
	switch driver {
	case DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite:
		return true
	default:
		return false
	}
}

func isNoSQLDatabaseDriver(driver string) bool {
	switch driver {
	case DatabaseDriverRedis, DatabaseDriverMongoDB:
		return true
	default:
		return false
	}
}

func sqlDatabaseDrivers(drivers []string) []string {
	result := make([]string, 0, len(drivers))
	for _, driver := range drivers {
		if isSQLDatabaseDriver(driver) {
			result = append(result, driver)
		}
	}
	return result
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

func configurationFormatValues() []string {
	return []string{ConfigurationFormatEnv, ConfigurationFormatYAML, ConfigurationFormatJSON, ConfigurationFormatTOML}
}

func databaseDriverValues() []string {
	return []string{DatabaseDriverNone, DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite, DatabaseDriverRedis, DatabaseDriverMongoDB}
}

func databaseSQLDriverValues() []string {
	return []string{DatabaseDriverNone, DatabaseDriverPostgres, DatabaseDriverMySQL, DatabaseDriverSQLite}
}

func databaseNoSQLDriverValues() []string {
	return []string{DatabaseDriverRedis, DatabaseDriverMongoDB}
}

func databaseFrameworkValues() []string {
	return []string{DatabaseFrameworkNone, DatabaseFrameworkPGX, DatabaseFrameworkDatabaseSQL, DatabaseFrameworkGORM}
}

func databaseMigrationValues() []string {
	return []string{DatabaseMigrationsNone, DatabaseMigrationsGoose, DatabaseMigrationsMigrate}
}

func loggingFrameworkValues() []string {
	return []string{LoggingFrameworkSlog, LoggingFrameworkZap, LoggingFrameworkZerolog, LoggingFrameworkLogrus}
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

func isSafeProjectName(name string) bool {
	if !projectNamePattern.MatchString(name) {
		return false
	}
	return name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
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
