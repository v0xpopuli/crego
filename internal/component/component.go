package component

const (
	CategoryProject           = "project"
	CategoryLayout            = "layout"
	CategoryServer            = "server"
	CategoryConfiguration     = "configuration"
	CategoryDatabase          = "database"
	CategoryDatabaseFramework = "database.framework"
	CategorySQLDatabase       = "sql_database"
	CategoryORMFramework      = "orm_framework"
	CategoryNoSQLDatabase     = "nosql_database"
	CategoryMigrations        = "migrations"
	CategoryTaskScheduler     = "task_scheduler"
	CategoryLogging           = "logging"
	CategoryObservability     = "observability"
	CategoryDeployment        = "deployment"
	CategoryCI                = "ci"

	IDProjectWeb = "project.web"
	IDProjectCLI = "project.cli"

	IDLayoutMinimal = "layout.minimal"
	IDLayoutLayered = "layout.layered"

	IDServerNetHTTP = "server.nethttp"
	IDServerChi     = "server.chi"
	IDServerGin     = "server.gin"
	IDServerEcho    = "server.echo"
	IDServerFiber   = "server.fiber"

	IDConfigurationEnv  = "configuration.env"
	IDConfigurationYAML = "configuration.yaml"
	IDConfigurationJSON = "configuration.json"
	IDConfigurationTOML = "configuration.toml"

	IDDatabaseNone     = "database.none"
	IDDatabasePostgres = "database.postgres"
	IDDatabaseMySQL    = "database.mysql"
	IDDatabaseSQLite   = "database.sqlite"
	IDDatabaseRedis    = "database.redis"
	IDDatabaseMongoDB  = "database.mongodb"

	IDDatabaseFrameworkPGX  = "database.framework.pgx"
	IDDatabaseFrameworkSQL  = "database.framework.sql"
	IDDatabaseFrameworkGORM = "database.framework.gorm"

	IDMigrationsNone    = "migrations.none"
	IDMigrationsGoose   = "migrations.goose"
	IDMigrationsMigrate = "migrations.migrate"

	IDTaskSchedulerNone   = "task_scheduler.none"
	IDTaskSchedulerGocron = "task_scheduler.gocron"

	IDLoggingSlog    = "logging.slog"
	IDLoggingZap     = "logging.zap"
	IDLoggingZerolog = "logging.zerolog"
	IDLoggingLogrus  = "logging.logrus"

	IDObservabilityHealth    = "observability.health"
	IDObservabilityReadiness = "observability.readiness"

	IDDeploymentDocker  = "deployment.docker"
	IDDeploymentCompose = "deployment.compose"

	IDCIGitHubActions = "ci.github_actions"
	IDCIGitLabCI      = "ci.gitlab_ci"
)

type (
	Component struct {
		ID          string
		Category    string
		Name        string
		Description string
		Requires    []string
		Conflicts   []string
		Files       []TemplateFile
		GoModules   []GoModule
		Hooks       []Hook
	}

	TemplateFile struct {
		Source string
		Target string
	}

	GoModule struct {
		Path    string
		Version string
	}

	Hook struct {
		Name string
	}
)
