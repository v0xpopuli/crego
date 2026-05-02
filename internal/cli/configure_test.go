package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/recipe"
)

type ConfigureCommandTestSuite struct {
	suite.Suite
}

func TestConfigureCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigureCommandTestSuite))
}

func (s *ConfigureCommandTestSuite) TestConfigureCommandFlags() {
	var out bytes.Buffer
	cmd := newConfigureCommand(&out, &globalOptions{})

	for _, flag := range []string{"recipe", "preset", "minimal", "overwrite"} {
		s.Require().NotNil(cmd.Flags().Lookup(flag))
	}
	s.Require().NotNil(cmd.Flags().ShorthandLookup("r"))
}

func (s *ConfigureCommandTestSuite) TestConfigureBaseRecipe() {
	s.Run("uses requested preset", func() {
		r, err := configureBaseRecipe(&configureOptions{preset: recipe.PresetWebPostgres})

		s.Require().NoError(err)
		s.Require().Equal(recipe.ProjectTypeWeb, r.Project.Type)
		s.Require().Equal(recipe.DatabaseDriverPostgres, r.Database.Driver)
	})

	s.Run("minimal starts from a CLI recipe without optional selections", func() {
		r, err := configureBaseRecipe(&configureOptions{minimal: true})

		s.Require().NoError(err)
		s.Require().Equal(recipe.ProjectTypeCLI, r.Project.Type)
		s.Require().Equal(recipe.DatabaseDriverNone, r.Database.Driver)
		s.Require().False(r.CI.GitHubActions)
		s.Require().False(r.Deployment.Docker)
	})
}
