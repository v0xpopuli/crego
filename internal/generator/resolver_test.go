package generator

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
)

type ResolverTestSuite struct {
	suite.Suite
}

func TestResolverTestSuite(t *testing.T) {
	suite.Run(t, new(ResolverTestSuite))
}

func (s *ResolverTestSuite) TestResolveWebChiPostgresGoose() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Layout.Style = recipe.LayoutStyleLayered
	r.Server.Framework = recipe.ServerFrameworkChi
	r.Database.Driver = recipe.DatabaseDriverPostgres
	r.Database.Framework = recipe.DatabaseFrameworkPGX
	r.Database.Migrations = recipe.DatabaseMigrationsGoose
	r.Observability.Health = true
	r.Observability.Readiness = true
	r.Deployment.Docker = true
	r.Deployment.Compose = true
	r.CI.GitHubActions = true
	r.CI.GitLabCI = true

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	s.Require().Equal([]string{
		component.IDProjectWeb,
		component.IDLayoutLayered,
		component.IDServerChi,
		component.IDConfigurationEnv,
		component.IDDatabasePostgres,
		component.IDDatabaseFrameworkPGX,
		component.IDMigrationsGoose,
		component.IDLoggingSlog,
		component.IDObservabilityHealth,
		component.IDObservabilityReadiness,
		component.IDDeploymentDocker,
		component.IDDeploymentCompose,
		component.IDCIGitHubActions,
		component.IDCIGitLabCI,
	}, planComponentIDs(plan))
}

func (s *ResolverTestSuite) TestResolveDeploymentAndCIFileTargets() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Deployment.Docker = true
	r.Deployment.Compose = true
	r.CI.GitHubActions = true
	r.CI.GitLabCI = true

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	targets, err := RenderFileTargets(r, plan)
	s.Require().NoError(err)
	s.Require().Contains(templateTargets(targets), "deployments/Dockerfile")
	s.Require().Contains(templateTargets(targets), "deployments/.dockerignore")
	s.Require().Contains(templateTargets(targets), "deployments/docker-compose.yml")
	s.Require().Contains(templateTargets(targets), ".github/workflows/test.yml")
	s.Require().Contains(templateTargets(targets), ".gitlab-ci.yml")
}

func (s *ResolverTestSuite) TestResolveWebGinMySQL() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkGin
	r.Database.Driver = recipe.DatabaseDriverMySQL
	r.Logging.Framework = recipe.LoggingFrameworkZerolog

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	s.Require().Equal([]string{
		component.IDProjectWeb,
		component.IDLayoutMinimal,
		component.IDServerGin,
		component.IDConfigurationEnv,
		component.IDDatabaseMySQL,
		component.IDDatabaseFrameworkSQL,
		component.IDLoggingZerolog,
	}, planComponentIDs(plan))
}

func (s *ResolverTestSuite) TestResolveWebNetHTTPSQLite() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkNetHTTP
	r.Database.Driver = recipe.DatabaseDriverSQLite

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	s.Require().Equal([]string{
		component.IDProjectWeb,
		component.IDLayoutMinimal,
		component.IDServerNetHTTP,
		component.IDConfigurationEnv,
		component.IDDatabaseSQLite,
		component.IDDatabaseFrameworkSQL,
		component.IDLoggingSlog,
	}, planComponentIDs(plan))
}

func (s *ResolverTestSuite) TestResolveWebRedisAndMongoDB() {
	for _, tc := range []struct {
		driver string
		id     string
		module string
	}{
		{driver: recipe.DatabaseDriverRedis, id: component.IDDatabaseRedis, module: "github.com/redis/go-redis/v9"},
		{driver: recipe.DatabaseDriverMongoDB, id: component.IDDatabaseMongoDB, module: "go.mongodb.org/mongo-driver/v2"},
	} {
		s.Run(tc.driver, func() {
			r := baseRecipe(recipe.ProjectTypeWeb)
			r.Database.Driver = tc.driver

			plan, err := Resolve(component.NewRegistry(), r)

			s.Require().NoError(err)
			s.Require().Contains(planComponentIDs(plan), tc.id)
			s.Require().NotContains(planComponentIDs(plan), component.IDDatabaseFrameworkSQL)
			s.Require().NotContains(planComponentIDs(plan), component.IDMigrationsMigrate)
			s.Require().Contains(planGoModulePaths(plan), tc.module)
		})
	}
}

func (s *ResolverTestSuite) TestResolveWebPostgresMySQLRedisMongoDB() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Database.Drivers = []string{recipe.DatabaseDriverPostgres, recipe.DatabaseDriverMySQL, recipe.DatabaseDriverRedis, recipe.DatabaseDriverMongoDB}
	r.Database.Framework = recipe.DatabaseFrameworkDatabaseSQL
	r.Database.Migrations = recipe.DatabaseMigrationsMigrate

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	ids := planComponentIDs(plan)
	s.Require().Contains(ids, component.IDDatabasePostgres)
	s.Require().Contains(ids, component.IDDatabaseMySQL)
	s.Require().Contains(ids, component.IDDatabaseRedis)
	s.Require().Contains(ids, component.IDDatabaseMongoDB)
	s.Require().Contains(ids, component.IDDatabaseFrameworkSQL)
	s.Require().Contains(ids, component.IDMigrationsMigrate)
	paths := planGoModulePaths(plan)
	s.Require().Contains(paths, "github.com/jackc/pgx/v5")
	s.Require().Contains(paths, "github.com/go-sql-driver/mysql")
	s.Require().Contains(paths, "github.com/redis/go-redis/v9")
	s.Require().Contains(paths, "go.mongodb.org/mongo-driver/v2")
	s.Require().Contains(paths, "github.com/golang-migrate/migrate/v4")
}

func (s *ResolverTestSuite) TestResolveCLIBasic() {
	r := baseRecipe(recipe.ProjectTypeCLI)

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	s.Require().Equal([]string{
		component.IDProjectCLI,
		component.IDLayoutMinimal,
		component.IDDatabaseNone,
		component.IDMigrationsNone,
		component.IDLoggingSlog,
	}, planComponentIDs(plan))
}

func (s *ResolverTestSuite) TestRejectsServerFrameworkConflict() {
	registry := component.NewRegistryFromComponents([]component.Component{
		{ID: component.IDProjectWeb},
		{ID: component.IDLayoutMinimal},
		{ID: component.IDServerNetHTTP, Requires: []string{component.IDProjectWeb}, Conflicts: []string{component.IDServerChi}},
		{ID: component.IDServerChi, Requires: []string{component.IDProjectWeb, component.IDServerNetHTTP}, Conflicts: []string{component.IDServerNetHTTP}},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDLoggingSlog},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkChi

	_, err := Resolve(registry, r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), component.IDServerNetHTTP+" conflicts with "+component.IDServerChi)
}

func (s *ResolverTestSuite) TestRejectsDatabaseFrameworkCompatibility() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Database.Driver = recipe.DatabaseDriverMySQL
	r.Database.Framework = recipe.DatabaseFrameworkPGX

	_, err := Resolve(component.NewRegistry(), r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "database.framework=pgx is only supported with database.driver=postgres")
}

func (s *ResolverTestSuite) TestRejectsMigrationsWithoutDatabase() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Database.Driver = recipe.DatabaseDriverNone
	r.Database.Migrations = recipe.DatabaseMigrationsGoose

	_, err := Resolve(component.NewRegistry(), r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "database.migrations=goose requires database.driver to be postgres, mysql, or sqlite")
}

func (s *ResolverTestSuite) TestRejectsNoSQLMigrationsAndFrameworks() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Database.Driver = recipe.DatabaseDriverMongoDB
	r.Database.Framework = recipe.DatabaseFrameworkGORM
	r.Database.Migrations = recipe.DatabaseMigrationsGoose

	_, err := Resolve(component.NewRegistry(), r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "database.framework=gorm is only supported with SQL database drivers")
	s.Require().Contains(err.Error(), "database.migrations=goose is only supported with SQL database drivers")
}

func (s *ResolverTestSuite) TestAddsExactDatabaseGoModules() {
	cases := []struct {
		name      string
		driver    string
		framework string
		migration string
		expected  []string
		absent    []string
	}{
		{
			name:      "postgres pgx goose",
			driver:    recipe.DatabaseDriverPostgres,
			framework: recipe.DatabaseFrameworkPGX,
			migration: recipe.DatabaseMigrationsGoose,
			expected:  []string{"github.com/jackc/pgx/v5", "github.com/pressly/goose/v3"},
			absent:    []string{"gorm.io/gorm", "github.com/golang-migrate/migrate/v4"},
		},
		{
			name:      "mysql gorm migrate",
			driver:    recipe.DatabaseDriverMySQL,
			framework: recipe.DatabaseFrameworkGORM,
			migration: recipe.DatabaseMigrationsMigrate,
			expected:  []string{"gorm.io/gorm", "gorm.io/driver/mysql", "github.com/golang-migrate/migrate/v4"},
			absent:    []string{"github.com/go-sql-driver/mysql", "github.com/pressly/goose/v3"},
		},
		{
			name:      "sqlite sql goose",
			driver:    recipe.DatabaseDriverSQLite,
			framework: recipe.DatabaseFrameworkDatabaseSQL,
			migration: recipe.DatabaseMigrationsGoose,
			expected:  []string{"modernc.org/sqlite", "github.com/pressly/goose/v3"},
			absent:    []string{"gorm.io/gorm", "github.com/golang-migrate/migrate/v4"},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			r := baseRecipe(recipe.ProjectTypeWeb)
			r.Database.Driver = tc.driver
			r.Database.Framework = tc.framework
			r.Database.Migrations = tc.migration

			plan, err := Resolve(component.NewRegistry(), r)

			s.Require().NoError(err)
			paths := planGoModulePaths(plan)
			for _, expected := range tc.expected {
				s.Require().Contains(paths, expected)
			}
			for _, absent := range tc.absent {
				s.Require().NotContains(paths, absent)
			}
		})
	}
}

func (s *ResolverTestSuite) TestRejectsMissingDependency() {
	registry := component.NewRegistryFromComponents([]component.Component{
		{ID: component.IDProjectWeb},
		{ID: component.IDLayoutMinimal},
		{ID: component.IDServerChi, Requires: []string{component.IDProjectWeb, "server.router"}},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDLoggingSlog},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkChi

	_, err := Resolve(registry, r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), component.IDServerChi+" requires missing component server.router")
}

func (s *ResolverTestSuite) TestEliminatesDuplicates() {
	registry := component.NewRegistryFromComponents([]component.Component{
		{
			ID:        component.IDProjectWeb,
			Files:     []component.TemplateFile{{Source: "app.tmpl", Target: "app.go"}},
			GoModules: []component.GoModule{{Path: "example.com/module", Version: "v1.0.0"}},
			Hooks:     []component.Hook{{Name: "go-fmt"}},
		},
		{ID: component.IDLayoutMinimal},
		{
			ID:       component.IDServerChi,
			Requires: []string{component.IDProjectWeb},
			Files: []component.TemplateFile{
				{Source: "app.tmpl", Target: "app.go"},
				{Source: "server.tmpl", Target: "server.go"},
			},
			GoModules: []component.GoModule{
				{Path: "example.com/module", Version: "v1.0.0"},
				{Path: "example.com/server", Version: "v1.0.0"},
			},
			Hooks: []component.Hook{
				{Name: "go-fmt"},
				{Name: "go-mod-tidy"},
			},
		},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDLoggingSlog},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkChi

	plan, err := Resolve(registry, r)

	s.Require().NoError(err)
	s.Require().Equal([]string{component.IDProjectWeb, component.IDLayoutMinimal, component.IDServerChi, component.IDConfigurationEnv, component.IDDatabaseNone, component.IDMigrationsNone, component.IDLoggingSlog}, planComponentIDs(plan))
	s.Require().Len(plan.Files, 2)
	s.Require().Len(plan.GoModules, 2)
	s.Require().Len(plan.Hooks, 2)
}

func (s *ResolverTestSuite) TestRejectsUnknownComponent() {
	r := baseRecipe(recipe.ProjectTypeWorker)

	_, err := Resolve(component.NewRegistry(), r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unknown component project.worker")
}

func baseRecipe(projectType string) *recipe.Recipe {
	return &recipe.Recipe{
		Version: recipe.VersionV1,
		Project: recipe.ProjectConfig{
			Name:   "example",
			Module: "example.com/example",
			Type:   projectType,
		},
	}
}

func planComponentIDs(plan *Plan) []string {
	ids := make([]string, 0, len(plan.Components))
	for _, component := range plan.Components {
		ids = append(ids, component.ID)
	}
	return ids
}

func planGoModulePaths(plan *Plan) []string {
	paths := make([]string, 0, len(plan.GoModules))
	for _, module := range plan.GoModules {
		paths = append(paths, module.Path)
	}
	return paths
}

func templateTargets(files []component.TemplateFile) []string {
	targets := make([]string, 0, len(files))
	for _, file := range files {
		targets = append(targets, file.Target)
	}
	return targets
}
