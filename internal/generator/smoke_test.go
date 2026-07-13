package generator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
	"github.com/v0xpopuli/crego/internal/testutil"
)

const (
	smokeCommandTimeout = 3 * time.Minute
	smokeStartTimeout   = 45 * time.Second
	smokeStopTimeout    = 10 * time.Second
)

type (
	smokeCase struct {
		Name string
	}

	smokeFixtureCoverage struct {
		ServerFrameworks     map[string]struct{}
		ConfigurationFormats map[string]struct{}
		LoggingFrameworks    map[string]struct{}
		SQLDrivers           map[string]struct{}
		NoSQLDrivers         map[string]struct{}
		TaskSchedulers       map[string]struct{}
		DockerCompose        bool
	}
)

var smokeCases = []smokeCase{
	{Name: "web-nethttp-env-slog"},
	{Name: "web-chi-yaml-zap-postgres-pgx-goose-gocron"},
	{Name: "web-gin-json-zerolog-mysql-sql-migrate"},
	{Name: "web-echo-toml-logrus-sqlite-gorm-goose"},
	{Name: "web-fiber-env-slog-docker-compose"},
	{Name: "web-redis-gocron"},
	{Name: "web-mongodb"},
	{Name: "cli-basic"},
}

func TestGeneratedProjectsCompileAndBasicToolingWorks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated-project smoke test in short mode")
	}

	for _, tc := range smokeCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			r, outDir := generateSmokeProject(t, tc.Name)

			runGoCommand(t, outDir, "mod", "tidy")
			runGoCommand(t, outDir, "test", "./...")
			requireTaskSchedulerOutput(t, r, outDir)
			runDockerComposeConfigIfAvailable(t, r, outDir)
			runWebRuntimeProbeIfPractical(t, r, outDir)
		})
	}
}

func TestSmokeFixtureMatrixCoversRequiredChoices(t *testing.T) {
	coverage := newSmokeFixtureCoverage()
	for _, tc := range smokeCases {
		r, err := recipe.Load(smokeRecipePath(tc.Name))
		require.NoError(t, err)
		coverage.Add(r)
	}

	requireSetContainsAll(t, "server frameworks", coverage.ServerFrameworks, []string{
		recipe.ServerFrameworkNetHTTP,
		recipe.ServerFrameworkChi,
		recipe.ServerFrameworkGin,
		recipe.ServerFrameworkEcho,
		recipe.ServerFrameworkFiber,
	})
	requireSetContainsAll(t, "configuration formats", coverage.ConfigurationFormats, []string{
		recipe.ConfigurationFormatEnv,
		recipe.ConfigurationFormatYAML,
		recipe.ConfigurationFormatJSON,
		recipe.ConfigurationFormatTOML,
	})
	requireSetContainsAll(t, "logging frameworks", coverage.LoggingFrameworks, []string{
		recipe.LoggingFrameworkSlog,
		recipe.LoggingFrameworkZap,
		recipe.LoggingFrameworkZerolog,
		recipe.LoggingFrameworkLogrus,
	})
	requireSetContainsAll(t, "SQL drivers", coverage.SQLDrivers, []string{
		recipe.DatabaseDriverPostgres,
		recipe.DatabaseDriverMySQL,
		recipe.DatabaseDriverSQLite,
	})
	requireSetContainsAll(t, "NoSQL drivers", coverage.NoSQLDrivers, []string{
		recipe.DatabaseDriverRedis,
		recipe.DatabaseDriverMongoDB,
	})
	requireSetContainsAll(t, "task schedulers", coverage.TaskSchedulers, []string{
		recipe.TaskSchedulerGocron,
		recipe.TaskSchedulerNone,
	})
	require.True(t, coverage.DockerCompose, "smoke fixture matrix must cover Docker Compose output")
}

func generateSmokeProject(t *testing.T, name string) (*recipe.Recipe, string) {
	t.Helper()

	r, err := recipe.Load(smokeRecipePath(name))
	require.NoError(t, err)

	plan, err := Resolve(component.NewRegistry(), r)
	require.NoError(t, err)

	outDir := t.TempDir()
	_, err = NewGenerator(templatefs.FS).Generate(context.Background(), r, plan, Options{OutDir: outDir})
	require.NoError(t, err)
	return r, outDir
}

func runGoCommand(t *testing.T, outDir string, args ...string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), smokeCommandTimeout)
	defer cancel()
	testutil.RunCommand(t, ctx, outDir, nil, "go", args...)
}

func requireTaskSchedulerOutput(t *testing.T, r *recipe.Recipe, outDir string) {
	t.Helper()

	if r.TaskScheduler == recipe.TaskSchedulerNone {
		requireGeneratedPathAbsent(t, outDir, filepath.Join("internal", "scheduler"))
		return
	}
	if r.TaskScheduler != recipe.TaskSchedulerGocron {
		return
	}

	requireGeneratedFileContains(t, outDir, filepath.Join("internal", "scheduler", "scheduler.go"), "github.com/go-co-op/gocron/v2")
	requireGeneratedFileContains(t, outDir, filepath.Join("internal", "scheduler", "scheduler.go"), "task scheduler started")
	requireGeneratedFileContains(t, outDir, filepath.Join("internal", "scheduler", "tasks", "example_cleanup.go"), "ExampleCleanupTask struct")
}

func runDockerComposeConfigIfAvailable(t *testing.T, r *recipe.Recipe, outDir string) {
	t.Helper()

	if !r.Deployment.Compose {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), smokeCommandTimeout)
	defer cancel()

	versionResult, err := testutil.RunCommandResult(ctx, outDir, nil, "docker", "compose", "version")
	if err != nil {
		t.Logf("skipping docker compose config: %s", versionResult.FailureMessage(err))
		return
	}

	testutil.RunCommand(t, ctx, outDir, nil, "docker", "compose", "-f", filepath.Join("deployments", "docker-compose.yml"), "config")
}

func runWebRuntimeProbeIfPractical(t *testing.T, r *recipe.Recipe, outDir string) {
	t.Helper()

	if r.Project.Type != recipe.ProjectTypeWeb {
		return
	}
	if !hasNoDatabase(r) {
		if r.TaskScheduler == recipe.TaskSchedulerGocron {
			t.Log("skipping generated scheduler runtime probe because the generated app requires an external database")
		}
		return
	}

	port := freeTCPPort(t)
	binaryPath := filepath.Join(t.TempDir(), r.Project.Name)

	ctx, cancel := context.WithTimeout(context.Background(), smokeCommandTimeout)
	defer cancel()
	testutil.RunCommand(t, ctx, outDir, nil, "go", "build", "-o", binaryPath, "./cmd/"+r.Project.Name)

	process, err := testutil.StartCommand(
		ctx,
		outDir,
		[]string{fmt.Sprintf("SERVER_PORT=%d", port)},
		binaryPath,
	)
	require.NoError(t, err)
	defer func() {
		err := process.Stop(smokeStopTimeout)
		require.NoError(t, err, "generated web service did not stop cleanly\n%s", process.Output())
	}()

	requireHTTPStatus(t, process, fmt.Sprintf("http://127.0.0.1:%d/health", port), http.StatusOK)
	if r.Observability.Readiness {
		requireHTTPStatus(t, process, fmt.Sprintf("http://127.0.0.1:%d/ready", port), http.StatusOK)
	}
}

func requireHTTPStatus(t *testing.T, process *testutil.StartedCommand, url string, expectedStatus int) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(smokeStartTimeout)
	var lastErr error
	var lastStatus int

	for time.Now().Before(deadline) {
		select {
		case err := <-process.Done():
			t.Fatalf("generated web service exited before %s became healthy: %v\n%s", url, err, process.Output())
		default:
		}

		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(200 * time.Millisecond)
			continue
		}

		lastStatus = resp.StatusCode
		_ = resp.Body.Close()
		if resp.StatusCode == expectedStatus {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s to return %d; last status=%d last error=%v\n%s", url, expectedStatus, lastStatus, lastErr, process.Output())
}

func requireGeneratedFileContains(t *testing.T, outDir string, target string, expected string) {
	t.Helper()

	path := filepath.Join(outDir, target)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), expected)
}

func requireGeneratedPathAbsent(t *testing.T, outDir string, target string) {
	t.Helper()

	path := filepath.Join(outDir, target)
	_, err := os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func requireSetContainsAll(t *testing.T, name string, values map[string]struct{}, required []string) {
	t.Helper()

	missing := []string{}
	for _, value := range required {
		if _, ok := values[value]; !ok {
			missing = append(missing, value)
		}
	}
	require.Empty(t, missing, "%s coverage is incomplete", name)
}

func smokeRecipePath(name string) string {
	return filepath.Join(goldenTestdataDirectory(), "recipes", name+".yaml")
}

func hasNoDatabase(r *recipe.Recipe) bool {
	drivers := recipe.DatabaseDrivers(r.Database)
	return len(drivers) == 1 && drivers[0] == recipe.DatabaseDriverNone
}

func freeTCPPort(t testing.TB) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free TCP port: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("release free TCP port: %v", err)
		}
	}()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unexpected TCP listener address: %s", listener.Addr())
	}
	return addr.Port
}

func newSmokeFixtureCoverage() smokeFixtureCoverage {
	return smokeFixtureCoverage{
		ServerFrameworks:     make(map[string]struct{}),
		ConfigurationFormats: make(map[string]struct{}),
		LoggingFrameworks:    make(map[string]struct{}),
		SQLDrivers:           make(map[string]struct{}),
		NoSQLDrivers:         make(map[string]struct{}),
		TaskSchedulers:       make(map[string]struct{}),
	}
}

func (c *smokeFixtureCoverage) Add(r *recipe.Recipe) {
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
	c.DockerCompose = c.DockerCompose || r.Deployment.Compose
}
