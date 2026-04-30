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
		component.IDDatabasePostgres,
		component.IDDatabaseFrameworkPGX,
		component.IDMigrationsGoose,
		component.IDConfigurationEnv,
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

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	s.Require().Equal([]string{
		component.IDProjectWeb,
		component.IDLayoutMinimal,
		component.IDServerGin,
		component.IDDatabaseMySQL,
		component.IDDatabaseFrameworkSQL,
		component.IDMigrationsNone,
		component.IDConfigurationEnv,
		component.IDLoggingSlog,
		component.IDCIGitHubActions,
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
		component.IDDatabaseSQLite,
		component.IDDatabaseFrameworkSQL,
		component.IDMigrationsNone,
		component.IDConfigurationEnv,
		component.IDLoggingSlog,
		component.IDCIGitHubActions,
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
		component.IDConfigurationEnv,
		component.IDLoggingSlog,
		component.IDCIGitHubActions,
	}, planComponentIDs(plan))
}

func (s *ResolverTestSuite) TestResolveConfigurationFormats() {
	for _, tc := range []struct {
		format      string
		componentID string
	}{
		{format: recipe.ConfigurationFormatYAML, componentID: component.IDConfigurationYAML},
		{format: recipe.ConfigurationFormatJSON, componentID: component.IDConfigurationJSON},
		{format: recipe.ConfigurationFormatTOML, componentID: component.IDConfigurationTOML},
	} {
		s.Run(tc.format, func() {
			r := baseRecipe(recipe.ProjectTypeWeb)
			r.Configuration.Format = tc.format

			plan, err := Resolve(component.NewRegistry(), r)

			s.Require().NoError(err)
			s.Require().Contains(planComponentIDs(plan), tc.componentID)
		})
	}
}

func (s *ResolverTestSuite) TestResolveLoggingProviders() {
	for _, tc := range []struct {
		provider    string
		componentID string
	}{
		{provider: recipe.LoggingProviderZap, componentID: component.IDLoggingZap},
		{provider: recipe.LoggingProviderZerolog, componentID: component.IDLoggingZerolog},
		{provider: recipe.LoggingProviderLogrus, componentID: component.IDLoggingLogrus},
	} {
		s.Run(tc.provider, func() {
			r := baseRecipe(recipe.ProjectTypeWeb)
			r.Logging.Provider = tc.provider

			plan, err := Resolve(component.NewRegistry(), r)

			s.Require().NoError(err)
			s.Require().Contains(planComponentIDs(plan), tc.componentID)
		})
	}
}

func (s *ResolverTestSuite) TestResolveGitLabCI() {
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.CI.GitLabCI = true

	plan, err := Resolve(component.NewRegistry(), r)

	s.Require().NoError(err)
	s.Require().Contains(planComponentIDs(plan), component.IDCIGitHubActions)
	s.Require().Contains(planComponentIDs(plan), component.IDCIGitLabCI)
}

func (s *ResolverTestSuite) TestRejectsServerFrameworkConflict() {
	registry := component.NewRegistryFromComponents([]component.Component{
		{ID: component.IDProjectWeb},
		{ID: component.IDLayoutMinimal},
		{ID: component.IDServerNetHTTP, Requires: []string{component.IDProjectWeb}, Conflicts: []string{component.IDServerChi}},
		{ID: component.IDServerChi, Requires: []string{component.IDProjectWeb, component.IDServerNetHTTP}, Conflicts: []string{component.IDServerNetHTTP}},
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDLoggingSlog},
		{ID: component.IDCIGitHubActions},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkChi

	_, err := Resolve(registry, r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), component.IDServerNetHTTP+" conflicts with "+component.IDServerChi)
}

func (s *ResolverTestSuite) TestRejectsConfigurationConflict() {
	registry := component.NewRegistryFromComponents([]component.Component{
		{ID: component.IDProjectWeb},
		{ID: component.IDLayoutMinimal},
		{ID: component.IDServerNetHTTP, Requires: []string{component.IDProjectWeb}},
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDConfigurationEnv, Conflicts: []string{component.IDConfigurationYAML}},
		{ID: component.IDConfigurationYAML, Requires: []string{component.IDConfigurationEnv}, Conflicts: []string{component.IDConfigurationEnv}},
		{ID: component.IDLoggingSlog},
		{ID: component.IDCIGitHubActions},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Configuration.Format = recipe.ConfigurationFormatYAML

	_, err := Resolve(registry, r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), component.IDConfigurationEnv+" conflicts with "+component.IDConfigurationYAML)
}

func (s *ResolverTestSuite) TestRejectsLoggingConflict() {
	registry := component.NewRegistryFromComponents([]component.Component{
		{ID: component.IDProjectWeb},
		{ID: component.IDLayoutMinimal},
		{ID: component.IDServerNetHTTP, Requires: []string{component.IDProjectWeb}},
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDLoggingSlog, Conflicts: []string{component.IDLoggingZap}},
		{ID: component.IDLoggingZap, Requires: []string{component.IDLoggingSlog}, Conflicts: []string{component.IDLoggingSlog}},
		{ID: component.IDCIGitHubActions},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Logging.Provider = recipe.LoggingProviderZap

	_, err := Resolve(registry, r)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), component.IDLoggingSlog+" conflicts with "+component.IDLoggingZap)
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
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDLoggingSlog},
		{ID: component.IDCIGitHubActions},
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
		{ID: component.IDDatabaseNone},
		{ID: component.IDMigrationsNone},
		{ID: component.IDConfigurationEnv},
		{ID: component.IDLoggingSlog},
		{ID: component.IDCIGitHubActions},
	})
	r := baseRecipe(recipe.ProjectTypeWeb)
	r.Server.Framework = recipe.ServerFrameworkChi

	plan, err := Resolve(registry, r)

	s.Require().NoError(err)
	s.Require().Equal([]string{component.IDProjectWeb, component.IDLayoutMinimal, component.IDServerChi, component.IDDatabaseNone, component.IDMigrationsNone, component.IDConfigurationEnv, component.IDLoggingSlog, component.IDCIGitHubActions}, planComponentIDs(plan))
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
