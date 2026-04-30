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

	DatabaseDriverNone     = "none"
	DatabaseDriverPostgres = "postgres"
	DatabaseDriverMySQL    = "mysql"
	DatabaseDriverSQLite   = "sqlite"

	DatabaseFrameworkNone        = "none"
	DatabaseFrameworkPGX         = "pgx"
	DatabaseFrameworkDatabaseSQL = "database/sql"
	DatabaseFrameworkGORM        = "gorm"

	DatabaseMigrationsNone    = "none"
	DatabaseMigrationsGoose   = "goose"
	DatabaseMigrationsMigrate = "migrate"

	LoggingProviderSlog    = "slog"
	LoggingProviderZap     = "zap"
	LoggingProviderZerolog = "zerolog"

	LoggingFormatText = "text"
	LoggingFormatJSON = "json"

	PresetWebBasic    = "web-basic"
	PresetWebPostgres = "web-postgres"
	PresetWebMySQL    = "web-mysql"
	PresetWebSQLite   = "web-sqlite"
	PresetCLIBasic    = "cli-basic"
)

type (
	Recipe struct {
		Version       string              `yaml:"version"`
		Project       ProjectConfig       `yaml:"project"`
		Go            GoConfig            `yaml:"go"`
		Layout        LayoutConfig        `yaml:"layout"`
		Server        ServerConfig        `yaml:"server,omitempty"`
		Database      DatabaseConfig      `yaml:"database"`
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

	DatabaseConfig struct {
		Driver     string `yaml:"driver"`
		Framework  string `yaml:"framework"`
		Migrations string `yaml:"migrations"`
	}

	LoggingConfig struct {
		Provider       string `yaml:"provider"`
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
	}
)

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
