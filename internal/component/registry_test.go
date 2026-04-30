package component

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RegistryTestSuite struct {
	suite.Suite
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

func (s *RegistryTestSuite) TestListIncludesMVPComponents() {
	registry := NewRegistry()

	components := registry.List()

	s.Require().Equal([]string{
		IDProjectWeb,
		IDProjectCLI,
		IDLayoutMinimal,
		IDLayoutLayered,
		IDServerNetHTTP,
		IDServerChi,
		IDServerGin,
		IDServerEcho,
		IDServerFiber,
		IDConfigurationEnv,
		IDConfigurationYAML,
		IDConfigurationJSON,
		IDConfigurationTOML,
		IDDatabaseNone,
		IDDatabasePostgres,
		IDDatabaseMySQL,
		IDDatabaseSQLite,
		IDDatabaseFrameworkPGX,
		IDDatabaseFrameworkSQL,
		IDDatabaseFrameworkGORM,
		IDMigrationsNone,
		IDMigrationsGoose,
		IDMigrationsMigrate,
		IDConfigurationEnv,
		IDConfigurationYAML,
		IDConfigurationJSON,
		IDConfigurationTOML,
		IDLoggingSlog,
		IDLoggingZap,
		IDLoggingZerolog,
		IDLoggingLogrus,
		IDObservabilityHealth,
		IDObservabilityReadiness,
		IDDeploymentDocker,
		IDDeploymentCompose,
		IDCIGitHubActions,
		IDCIGitLabCI,
	}, componentIDs(components))
}

func (s *RegistryTestSuite) TestGetReturnsKnownComponent() {
	registry := NewRegistry()

	component, ok := registry.Get(IDServerGin)

	s.Require().True(ok)
	s.Require().Equal(IDServerGin, component.ID)
	s.Require().Equal(CategoryServer, component.Category)
	s.Require().Contains(component.Requires, IDProjectWeb)
}

func (s *RegistryTestSuite) TestGetRejectsUnknownComponent() {
	registry := NewRegistry()

	_, ok := registry.Get("server.martini")

	s.Require().False(ok)
}

func componentIDs(components []Component) []string {
	ids := make([]string, 0, len(components))
	for _, component := range components {
		ids = append(ids, component.ID)
	}
	return ids
}
