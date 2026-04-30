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
	}, planComponentIDs(plan))
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
		component.IDMigrationsNone,
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
		component.IDMigrationsNone,
		component.IDLoggingSlog,
	}, planComponentIDs(plan))
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
