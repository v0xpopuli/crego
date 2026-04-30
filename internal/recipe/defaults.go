package recipe

const (
	defaultGoVersion       = "1.24"
	defaultLayoutStyle     = LayoutStyleMinimal
	defaultServerFramework = ServerFrameworkNetHTTP
	defaultServerPort      = 8080
	defaultLoggingProvider = LoggingProviderSlog
	defaultLoggingFormat   = LoggingFormatText
)

func ApplyDefaults(r *Recipe) *Recipe {
	if r == nil {
		return nil
	}

	if r.Version == "" {
		r.Version = VersionV1
	}
	if r.Go.Version == "" {
		r.Go.Version = defaultGoVersion
	}
	if r.Layout.Style == "" {
		r.Layout.Style = defaultLayoutStyle
	}

	if r.Project.Type == ProjectTypeWeb {
		if r.Server.Framework == "" {
			r.Server.Framework = defaultServerFramework
		}
		if r.Server.Port == 0 {
			r.Server.Port = defaultServerPort
		}
		if !r.Server.gracefulShutdownSet {
			r.Server.GracefulShutdown = true
		}
	}

	if r.Database.Driver == "" {
		r.Database.Driver = DatabaseDriverNone
	}
	if r.Database.Framework == "" {
		r.Database.Framework = defaultDatabaseFramework(r.Database.Driver)
	}
	if r.Database.Migrations == "" {
		r.Database.Migrations = DatabaseMigrationsNone
	}

	if r.Logging.Provider == "" {
		r.Logging.Provider = defaultLoggingProvider
	}
	if r.Logging.Format == "" {
		r.Logging.Format = defaultLoggingFormat
	}

	return r
}

func defaultDatabaseFramework(driver string) string {
	switch driver {
	case DatabaseDriverPostgres:
		return DatabaseFrameworkPGX
	case DatabaseDriverMySQL, DatabaseDriverSQLite:
		return DatabaseFrameworkDatabaseSQL
	default:
		return DatabaseFrameworkNone
	}
}
