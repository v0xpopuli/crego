package generator

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
)

const updateGoldenEnv = "UPDATE_GOLDEN"

type (
	GoldenProjectMatrixTestSuite struct {
		suite.Suite
	}

	goldenCase struct {
		Name string
	}

	goldenFixtureCoverage struct {
		ServerFrameworks     map[string]struct{}
		ConfigurationFormats map[string]struct{}
		LoggingFrameworks    map[string]struct{}
		SQLDrivers           map[string]struct{}
		NoSQLDrivers         map[string]struct{}
		TaskSchedulers       map[string]struct{}
		GitHubActions        bool
		GitLabCI             bool
		DockerCompose        bool
	}
)

func TestGoldenProjectMatrixTestSuite(t *testing.T) {
	suite.Run(t, new(GoldenProjectMatrixTestSuite))
}

func (s *GoldenProjectMatrixTestSuite) TestGeneratedProjectMatrixMatchesGoldenFiles() {
	// Refresh with:
	// UPDATE_GOLDEN=1 go test ./internal/generator -run TestGoldenProjectMatrixTestSuite
	for _, tc := range goldenCases() {
		s.Run(tc.Name, func() {
			r, err := recipe.Load(goldenRecipePath(tc.Name))
			s.Require().NoError(err)

			plan, err := Resolve(component.NewRegistry(), r)
			s.Require().NoError(err)

			outDir := s.T().TempDir()
			_, err = NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})
			s.Require().NoError(err)

			goldenDir := goldenDirectory(tc.Name)
			if shouldUpdateGolden() {
				s.Require().NoError(replaceGoldenTree(goldenDir, outDir))
				return
			}

			s.requireGeneratedTreeMatchesGolden(goldenDir, outDir)
		})
	}
}

func (s *GoldenProjectMatrixTestSuite) TestFixtureMatrixCoversRequiredChoices() {
	coverage := newGoldenFixtureCoverage()
	for _, tc := range goldenCases() {
		r, err := recipe.Load(goldenRecipePath(tc.Name))
		s.Require().NoError(err)
		coverage.Add(r)
	}

	s.requireSetContainsAll("server frameworks", coverage.ServerFrameworks, []string{
		recipe.ServerFrameworkNetHTTP,
		recipe.ServerFrameworkChi,
		recipe.ServerFrameworkGin,
		recipe.ServerFrameworkEcho,
		recipe.ServerFrameworkFiber,
	})
	s.requireSetContainsAll("configuration formats", coverage.ConfigurationFormats, []string{
		recipe.ConfigurationFormatEnv,
		recipe.ConfigurationFormatYAML,
		recipe.ConfigurationFormatJSON,
		recipe.ConfigurationFormatTOML,
	})
	s.requireSetContainsAll("logging frameworks", coverage.LoggingFrameworks, []string{
		recipe.LoggingFrameworkSlog,
		recipe.LoggingFrameworkZap,
		recipe.LoggingFrameworkZerolog,
		recipe.LoggingFrameworkLogrus,
	})
	s.requireSetContainsAll("SQL drivers", coverage.SQLDrivers, []string{
		recipe.DatabaseDriverPostgres,
		recipe.DatabaseDriverMySQL,
		recipe.DatabaseDriverSQLite,
	})
	s.requireSetContainsAll("NoSQL drivers", coverage.NoSQLDrivers, []string{
		recipe.DatabaseDriverRedis,
		recipe.DatabaseDriverMongoDB,
	})
	s.requireSetContainsAll("task schedulers", coverage.TaskSchedulers, []string{
		recipe.TaskSchedulerGocron,
		recipe.TaskSchedulerNone,
	})
	s.Require().True(coverage.GitHubActions, "fixture matrix must cover GitHub Actions output")
	s.Require().True(coverage.GitLabCI, "fixture matrix must cover GitLab CI output")
	s.Require().True(coverage.DockerCompose, "fixture matrix must cover Docker Compose output")
}

func (s *GoldenProjectMatrixTestSuite) requireGeneratedTreeMatchesGolden(goldenDir string, generatedDir string) {
	s.T().Helper()
	s.Require().DirExists(goldenDir, "%s=1 refreshes golden files", updateGoldenEnv)

	expectedFiles, err := listRegularFiles(goldenDir)
	s.Require().NoError(err)
	actualFiles, err := listRegularFiles(generatedDir)
	s.Require().NoError(err)

	missingFiles := missingFrom(expectedFiles, actualFiles)
	unexpectedFiles := missingFrom(actualFiles, expectedFiles)
	s.Require().Empty(missingFiles, "missing generated files")
	s.Require().Empty(unexpectedFiles, "unexpected generated files")

	for _, target := range expectedFiles {
		expected, err := readNormalizedFile(filepath.Join(goldenDir, filepath.FromSlash(target)))
		s.Require().NoError(err)
		actual, err := readNormalizedFile(filepath.Join(generatedDir, filepath.FromSlash(target)))
		s.Require().NoError(err)
		if expected == actual {
			continue
		}
		s.Failf("golden file mismatch", "%s\n%s", target, unifiedFileDiff(filepath.Base(goldenDir), target, expected, actual))
	}
}

func (s *GoldenProjectMatrixTestSuite) requireSetContainsAll(name string, values map[string]struct{}, required []string) {
	s.T().Helper()
	missing := []string{}
	for _, value := range required {
		if _, ok := values[value]; !ok {
			missing = append(missing, value)
		}
	}
	s.Require().Empty(missing, "%s coverage is incomplete", name)
}

func goldenCases() []goldenCase {
	return []goldenCase{
		{Name: "web-nethttp-env-slog"},
		{Name: "web-chi-yaml-zap-postgres-pgx-goose-github-gitlab-gocron"},
		{Name: "web-gin-json-zerolog-mysql-sql-migrate"},
		{Name: "web-echo-toml-logrus-sqlite-gorm-goose"},
		{Name: "web-fiber-env-slog-docker-compose"},
		{Name: "web-redis-gocron"},
		{Name: "web-mongodb"},
		{Name: "cli-basic"},
	}
}

func goldenRecipePath(name string) string {
	return filepath.Join(goldenTestdataDirectory(), "recipes", name+".yaml")
}

func goldenDirectory(name string) string {
	return filepath.Join(goldenTestdataDirectory(), "golden", name)
}

func goldenTestdataDirectory() string {
	return filepath.Join("..", "..", "testdata")
}

func shouldUpdateGolden() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(updateGoldenEnv)))
	return value == "1" || value == "true"
}

func replaceGoldenTree(goldenDir string, generatedDir string) error {
	if err := os.RemoveAll(goldenDir); err != nil {
		return err
	}
	targets, err := listRegularFiles(generatedDir)
	if err != nil {
		return err
	}
	for _, target := range targets {
		content, err := readNormalizedFile(filepath.Join(generatedDir, filepath.FromSlash(target)))
		if err != nil {
			return err
		}
		destination := filepath.Join(goldenDir, filepath.FromSlash(target))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, []byte(content), regularFileMode); err != nil {
			return err
		}
	}
	return nil
}

func newGoldenFixtureCoverage() goldenFixtureCoverage {
	return goldenFixtureCoverage{
		ServerFrameworks:     make(map[string]struct{}),
		ConfigurationFormats: make(map[string]struct{}),
		LoggingFrameworks:    make(map[string]struct{}),
		SQLDrivers:           make(map[string]struct{}),
		NoSQLDrivers:         make(map[string]struct{}),
		TaskSchedulers:       make(map[string]struct{}),
	}
}

func (c *goldenFixtureCoverage) Add(r *recipe.Recipe) {
	if r == nil {
		return
	}
	if r.Project.Type == recipe.ProjectTypeWeb {
		c.ServerFrameworks[r.Server.Framework] = struct{}{}
	}
	c.ConfigurationFormats[r.Configuration.Format] = struct{}{}
	c.LoggingFrameworks[r.Logging.Framework] = struct{}{}
	for _, driver := range recipe.DatabaseDrivers(r.Database) {
		switch driver {
		case recipe.DatabaseDriverPostgres, recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite:
			c.SQLDrivers[driver] = struct{}{}
		case recipe.DatabaseDriverRedis, recipe.DatabaseDriverMongoDB:
			c.NoSQLDrivers[driver] = struct{}{}
		}
	}
	c.TaskSchedulers[r.TaskScheduler] = struct{}{}
	c.GitHubActions = c.GitHubActions || r.CI.GitHubActions
	c.GitLabCI = c.GitLabCI || r.CI.GitLabCI
	c.DockerCompose = c.DockerCompose || r.Deployment.Compose
}

func listRegularFiles(root string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func readNormalizedFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizeLineEndings(string(data)), nil
}

func normalizeLineEndings(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.ReplaceAll(value, "\r", "\n")
}

func missingFrom(expected []string, actual []string) []string {
	actualSet := make(map[string]struct{}, len(actual))
	for _, value := range actual {
		actualSet[value] = struct{}{}
	}
	missing := []string{}
	for _, value := range expected {
		if _, ok := actualSet[value]; !ok {
			missing = append(missing, value)
		}
	}
	return missing
}

func unifiedFileDiff(caseName string, target string, expected string, actual string) string {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: filepath.ToSlash(filepath.Join("testdata", "golden", caseName, target)),
		ToFile:   filepath.ToSlash(filepath.Join("generated", caseName, target)),
		Context:  3,
	})
	if err != nil {
		return err.Error()
	}
	return diff
}
