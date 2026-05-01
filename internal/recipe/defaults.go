package recipe

const (
	defaultGoVersion        = "1.25"
	defaultLayoutStyle      = LayoutStyleMinimal
	defaultServerFramework  = ServerFrameworkNetHTTP
	defaultServerPort       = 8080
	defaultConfigFormat     = ConfigurationFormatEnv
	defaultLoggingFramework = LoggingFrameworkSlog
	defaultLoggingFormat    = LoggingFormatText
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

	if r.Configuration.Format == "" {
		r.Configuration.Format = defaultConfigFormat
	}

	if r.Database.Driver == "" {
		if len(r.Database.Drivers) == 0 {
			r.Database.Driver = DatabaseDriverNone
		} else {
			r.Database.Driver = primaryDatabaseDriver(r.Database.Drivers)
		}
	}
	if len(r.Database.Drivers) == 0 {
		r.Database.Drivers = []string{r.Database.Driver}
	}
	r.Database.Driver = primaryDatabaseDriver(r.Database.Drivers)
	if r.Database.Framework == "" {
		r.Database.Framework = defaultDatabaseFramework(r.Database.Drivers)
	}
	if r.Database.Migrations == "" {
		r.Database.Migrations = DatabaseMigrationsNone
	}
	r.SQLDatabase = primarySQLDatabaseDriver(r.Database.Drivers)
	r.ORMFramework = r.Database.Framework
	r.NoSQLDatabase = noSQLDatabaseDrivers(r.Database.Drivers)
	r.Migrations = r.Database.Migrations
	if r.Database.SQL == "" {
		r.Database.SQL = r.SQLDatabase
	}
	if r.Database.ORMFramework == "" {
		r.Database.ORMFramework = r.Database.Framework
	}
	if len(r.Database.NoSQL) == 0 {
		r.Database.NoSQL = append([]string(nil), r.NoSQLDatabase...)
	}

	if r.Logging.Framework == "" {
		r.Logging.Framework = defaultLoggingFramework
	}
	if r.Logging.Format == "" {
		r.Logging.Format = defaultLoggingFormat
	}

	return r
}

func defaultDatabaseFramework(drivers []string) string {
	sqlDrivers := sqlDatabaseDrivers(drivers)
	if len(sqlDrivers) > 1 {
		return DatabaseFrameworkDatabaseSQL
	}
	if len(sqlDrivers) == 0 {
		return DatabaseFrameworkNone
	}

	switch sqlDrivers[0] {
	case DatabaseDriverPostgres:
		return DatabaseFrameworkPGX
	case DatabaseDriverMySQL, DatabaseDriverSQLite:
		return DatabaseFrameworkDatabaseSQL
	default:
		return DatabaseFrameworkNone
	}
}
