package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
)

type DeploymentTemplateTestSuite struct {
	suite.Suite
}

func TestDeploymentTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTemplateTestSuite))
}

func (s *DeploymentTemplateTestSuite) TestRendersDockerComposeAndCIFiles() {
	r := generatorTestRecipe()
	r.Server.Framework = recipe.ServerFrameworkChi
	r.Configuration.Format = recipe.ConfigurationFormatYAML
	r.Database.Driver = recipe.DatabaseDriverPostgres
	r.Database.Framework = recipe.DatabaseFrameworkPGX
	r.Logging.Framework = recipe.LoggingFrameworkZap
	r.Deployment.Docker = true
	r.Deployment.Compose = true
	r.CI.GitHubActions = true
	r.CI.GitLabCI = true
	r.CI.AzurePipelines = true

	plan, err := Resolve(component.NewRegistry(), r)
	s.Require().NoError(err)

	outDir := s.T().TempDir()
	result, err := NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})

	s.Require().NoError(err)
	s.Require().Contains(result.FilesWritten, "deployments/Dockerfile")
	s.Require().Contains(result.FilesWritten, "deployments/.dockerignore")
	s.Require().Contains(result.FilesWritten, "deployments/docker-compose.yml")
	s.Require().Contains(result.FilesWritten, ".github/workflows/test.yml")
	s.Require().Contains(result.FilesWritten, ".gitlab-ci.yml")
	s.Require().Contains(result.FilesWritten, "azure-pipelines.yml")

	dockerfile := readGeneratedFile(s, outDir, "deployments/Dockerfile")
	s.Require().Contains(dockerfile, "go mod tidy")
	s.Require().Contains(dockerfile, "go build -trimpath")
	s.Require().Contains(dockerfile, "./cmd/example")
	s.Require().Contains(dockerfile, "COPY --from=build /src/configs ./configs")
	s.Require().Contains(dockerfile, "USER app")
	s.Require().Contains(dockerfile, "EXPOSE 8080")

	compose := readGeneratedFile(s, outDir, "deployments/docker-compose.yml")
	s.Require().Contains(compose, "DATABASE_POSTGRES_HOST: postgres:5432")
	s.Require().Contains(compose, "DATABASE_POSTGRES_DATABASE: app")
	s.Require().Contains(compose, "CONFIG_PATH: configs/config.yaml")
	s.Require().Contains(compose, "postgres:")
	s.Require().NotContains(compose, "mysql:")

	githubActions := readGeneratedFile(s, outDir, ".github/workflows/test.yml")
	s.Require().Contains(githubActions, "uses: actions/setup-go@v5")
	s.Require().Contains(githubActions, "go mod tidy")
	s.Require().Contains(githubActions, "go test ./...")

	gitlabCI := readGeneratedFile(s, outDir, ".gitlab-ci.yml")
	s.Require().Contains(gitlabCI, "image: golang:1.25")
	s.Require().Contains(gitlabCI, "go mod tidy")
	s.Require().Contains(gitlabCI, "go vet ./...")
	s.Require().Contains(gitlabCI, "go test ./...")

	azurePipelines := readGeneratedFile(s, outDir, "azure-pipelines.yml")
	s.Require().Contains(azurePipelines, "version: '1.25'")
	s.Require().Contains(azurePipelines, "go mod tidy")
	s.Require().Contains(azurePipelines, "go test -v ./...")
}

func (s *DeploymentTemplateTestSuite) TestComposeOmitsDatabaseServiceWhenDatabaseIsNone() {
	r := generatorTestRecipe()
	r.Deployment.Compose = true

	plan, err := Resolve(component.NewRegistry(), r)
	s.Require().NoError(err)

	outDir := s.T().TempDir()
	_, err = NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})
	s.Require().NoError(err)

	compose := readGeneratedFile(s, outDir, "deployments/docker-compose.yml")
	s.Require().Contains(compose, "app:")
	s.Require().NotContains(compose, "postgres:")
	s.Require().NotContains(compose, "mysql:")
	s.Require().NotContains(compose, "redis:")
	s.Require().NotContains(compose, "mongodb:")
}

func readGeneratedFile(s *DeploymentTemplateTestSuite, root string, target string) string {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target)))
	s.Require().NoError(err)
	return string(content)
}
