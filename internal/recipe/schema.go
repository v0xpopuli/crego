package recipe

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	VersionV1 = "v1"

	ProjectTypeWeb     = "web"
	ProjectTypeCLI     = "cli"
	ProjectTypeWorker  = "worker"
	ProjectTypeLibrary = "library"

	LayoutStyleMinimal = "minimal"
	LayoutStyleLayered = "layered"
	LayoutStyleDomain  = "domain"

	ServerFrameworkNetHTTP = "nethttp"
	ServerFrameworkChi     = "chi"
	ServerFrameworkGin     = "gin"
	ServerFrameworkEcho    = "echo"
	ServerFrameworkFiber   = "fiber"

	ConfigurationFormatEnv  = "env"
	ConfigurationFormatYAML = "yaml"
	ConfigurationFormatJSON = "json"
	ConfigurationFormatTOML = "toml"

	DatabaseDriverNone     = "none"
	DatabaseDriverPostgres = "postgres"
	DatabaseDriverMySQL    = "mysql"
	DatabaseDriverSQLite   = "sqlite"
	DatabaseDriverRedis    = "redis"
	DatabaseDriverMongoDB  = "mongodb"

	DatabaseFrameworkNone        = "none"
	DatabaseFrameworkPGX         = "pgx"
	DatabaseFrameworkDatabaseSQL = "sql"
	DatabaseFrameworkGORM        = "gorm"

	DatabaseMigrationsNone    = "none"
	DatabaseMigrationsGoose   = "goose"
	DatabaseMigrationsMigrate = "migrate"

	LoggingFrameworkSlog    = "slog"
	LoggingFrameworkZap     = "zap"
	LoggingFrameworkZerolog = "zerolog"
	LoggingFrameworkLogrus  = "logrus"

	LoggingFormatText = "text"
	LoggingFormatJSON = "json"

	PresetWebBasic    = "web-basic"
	PresetWebPostgres = "web-postgres"
	PresetWebMySQL    = "web-mysql"
	PresetWebSQLite   = "web-sqlite"
	PresetWebRedis    = "web-redis"
	PresetWebMongoDB  = "web-mongodb"
	PresetCLIBasic    = "cli-basic"
)

type (
	Recipe struct {
		Version       string              `yaml:"version"`
		Project       ProjectConfig       `yaml:"project"`
		Go            GoConfig            `yaml:"go"`
		Layout        LayoutConfig        `yaml:"layout"`
		Server        ServerConfig        `yaml:"server,omitempty"`
		Configuration ConfigurationConfig `yaml:"configuration"`
		SQLDatabase   string              `yaml:"sql_database,omitempty"`
		ORMFramework  string              `yaml:"orm_framework,omitempty"`
		NoSQLDatabase NoSQLDrivers        `yaml:"nosql_database,omitempty"`
		Migrations    string              `yaml:"migrations,omitempty"`
		Database      DatabaseConfig      `yaml:"database,omitempty"`
		Logging       LoggingConfig       `yaml:"logging"`
		Observability ObservabilityConfig `yaml:"observability"`
		Deployment    DeploymentConfig    `yaml:"deployment"`
		CI            CIConfig            `yaml:"ci"`
	}

	ProjectConfig struct {
		Name   string `yaml:"name"`
		Module string `yaml:"module"`
		Type   string `yaml:"type"`
	}

	GoConfig struct {
		Version string `yaml:"version"`
	}

	LayoutConfig struct {
		Style string `yaml:"style"`
	}

	ServerConfig struct {
		Framework        string `yaml:"framework"`
		Port             int    `yaml:"port"`
		GracefulShutdown bool   `yaml:"graceful_shutdown"`

		gracefulShutdownSet bool
	}

	ConfigurationConfig struct {
		Format string `yaml:"format"`
	}

	DatabaseConfig struct {
		Driver     string   `yaml:"driver"`
		Drivers    []string `yaml:"drivers,omitempty"`
		Framework  string   `yaml:"framework,omitempty"`
		Migrations string   `yaml:"migrations,omitempty"`
	}

	LoggingConfig struct {
		Framework      string `yaml:"framework"`
		Format         string `yaml:"format"`
		RequestLogging bool   `yaml:"request_logging"`
	}

	ObservabilityConfig struct {
		Health    bool `yaml:"health"`
		Readiness bool `yaml:"readiness"`
		Metrics   bool `yaml:"metrics"`
		Tracing   bool `yaml:"tracing"`
	}

	DeploymentConfig struct {
		Docker  bool `yaml:"docker"`
		Compose bool `yaml:"compose"`
	}

	CIConfig struct {
		GitHubActions bool `yaml:"github_actions"`
		GitLabCI      bool `yaml:"gitlab_ci"`
	}
)

func (r Recipe) MarshalYAML() (any, error) {
	resolved := r
	Normalize(&resolved)
	ApplyDefaults(&resolved)

	drivers := DatabaseDrivers(resolved.Database)
	noSQLDrivers := noSQLDatabaseDrivers(drivers)
	sqlDatabase := primarySQLDatabaseDriver(drivers)
	ormFramework := resolved.Database.Framework
	if sqlDatabase == DatabaseDriverNone || ormFramework == DatabaseFrameworkNone {
		ormFramework = ""
	}
	migrations := resolved.Database.Migrations
	if sqlDatabase == DatabaseDriverNone {
		migrations = ""
	}

	type recipeYAML struct {
		Version       string              `yaml:"version"`
		Project       ProjectConfig       `yaml:"project"`
		Go            GoConfig            `yaml:"go"`
		Layout        LayoutConfig        `yaml:"layout"`
		Server        ServerConfig        `yaml:"server,omitempty"`
		Configuration ConfigurationConfig `yaml:"configuration"`
		SQLDatabase   string              `yaml:"sql_database"`
		ORMFramework  string              `yaml:"orm_framework,omitempty"`
		NoSQLDatabase any                 `yaml:"nosql_database"`
		Migrations    string              `yaml:"migrations,omitempty"`
		Logging       LoggingConfig       `yaml:"logging"`
		Observability ObservabilityConfig `yaml:"observability"`
		Deployment    DeploymentConfig    `yaml:"deployment"`
		CI            CIConfig            `yaml:"ci"`
	}

	var noSQL any = DatabaseDriverNone
	if len(noSQLDrivers) == 1 {
		noSQL = noSQLDrivers[0]
	} else if len(noSQLDrivers) > 1 {
		noSQL = noSQLDrivers
	}

	return recipeYAML{
		Version:       resolved.Version,
		Project:       resolved.Project,
		Go:            resolved.Go,
		Layout:        resolved.Layout,
		Server:        resolved.Server,
		Configuration: resolved.Configuration,
		SQLDatabase:   sqlDatabase,
		ORMFramework:  ormFramework,
		NoSQLDatabase: noSQL,
		Migrations:    migrations,
		Logging:       resolved.Logging,
		Observability: resolved.Observability,
		Deployment:    resolved.Deployment,
		CI:            resolved.CI,
	}, nil
}

func (c DatabaseConfig) MarshalYAML() (any, error) {
	drivers := DatabaseDrivers(c)
	if len(drivers) > 1 {
		return struct {
			Drivers    []string `yaml:"drivers"`
			Framework  string   `yaml:"framework,omitempty"`
			Migrations string   `yaml:"migrations,omitempty"`
		}{
			Drivers:    drivers,
			Framework:  c.Framework,
			Migrations: c.Migrations,
		}, nil
	}
	if len(drivers) == 0 || drivers[0] == DatabaseDriverNone || drivers[0] == DatabaseDriverRedis || drivers[0] == DatabaseDriverMongoDB {
		return struct {
			Driver string `yaml:"driver"`
		}{
			Driver: firstDatabaseDriver(drivers),
		}, nil
	}

	return struct {
		Driver     string `yaml:"driver"`
		Framework  string `yaml:"framework,omitempty"`
		Migrations string `yaml:"migrations,omitempty"`
	}{
		Driver:     drivers[0],
		Framework:  c.Framework,
		Migrations: c.Migrations,
	}, nil
}

func (c *ServerConfig) UnmarshalYAML(value *yaml.Node) error {
	for _, key := range yamlMappingKeys(value) {
		switch key {
		case "framework", "port", "graceful_shutdown":
		default:
			return fmt.Errorf("unknown server field %q", key)
		}
	}

	type serverConfig ServerConfig
	var decoded serverConfig
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = ServerConfig(decoded)
	c.gracefulShutdownSet = yamlMappingContains(value, "graceful_shutdown")
	return nil
}

func (c *LoggingConfig) UnmarshalYAML(value *yaml.Node) error {
	for _, key := range yamlMappingKeys(value) {
		switch key {
		case "framework", "format", "request_logging":
		default:
			return fmt.Errorf("unknown logging field %q", key)
		}
	}

	type loggingConfig LoggingConfig
	var decoded loggingConfig
	if err := value.Decode(&decoded); err != nil {
		return err
	}

	*c = LoggingConfig(decoded)
	return nil
}
