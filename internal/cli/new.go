package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/generator"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
	"github.com/v0xpopuli/crego/internal/tui"
)

type newOptions struct {
	recipePath     string
	preset         string
	projectType    string
	layout         string
	server         string
	configuration  string
	logging        string
	database       string
	framework      string
	migrations     string
	docker         bool
	compose        bool
	githubActions  bool
	gitlabCI       bool
	azurePipelines bool
	health         bool
	readiness      bool
	outDir         string
	dryRun         bool
	force          bool
	skipGoModTidy  bool
	skipGitInit    bool
	nonInteractive bool
	overwrite      bool
}

func newNewCommand(out io.Writer, globalOpts *globalOptions) *cobra.Command {
	opts := &newOptions{}
	cmd := &cobra.Command{
		Use:   "new [module]",
		Short: "Create a new Go project",
		Long: `Create a new Go project from an interactive wizard, a recipe, or
non-interactive flags.

By default, crego new opens the TUI wizard, previews the resolved generation
plan, and can generate the project immediately or save the recipe only.`,
		Example: `  crego new github.com/example/orders-web --type web --server chi --non-interactive
  crego new github.com/example/orders-web --preset web-postgres
  crego new github.com/example/orders-web --type web --server chi --configuration yaml --logging zap --database postgres --framework pgx --migrations goose --non-interactive
  crego new --recipe crego.yaml --out ./orders-web`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runNew(out, globalOpts, opts, args)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVarP(&opts.recipePath, "recipe", "r", "", "Path to an existing recipe file")
	cmd.Flags().StringVar(&opts.preset, "preset", "", "Starter preset: web-basic, web-postgres, web-mysql, web-sqlite, web-redis, web-mongodb, cli-basic")
	cmd.Flags().StringVar(&opts.projectType, "type", "", "Project type: web or cli")
	cmd.Flags().StringVar(&opts.layout, "layout", "", "Project layout: minimal or layered")
	cmd.Flags().StringVar(&opts.server, "server", "", "Web server: nethttp, chi, gin, echo, or fiber")
	cmd.Flags().StringVar(&opts.configuration, "configuration", "", "Configuration format: env, yaml, json, or toml")
	cmd.Flags().StringVar(&opts.logging, "logging", "", "Logging framework: slog, zap, zerolog, or logrus")
	cmd.Flags().StringVar(&opts.database, "database", "", "Database driver: none, postgres, mysql, sqlite, redis, or mongodb")
	cmd.Flags().StringVar(&opts.framework, "framework", "", "Database framework: none, pgx, sql, or gorm")
	cmd.Flags().StringVar(&opts.migrations, "migrations", "", "Migration tool: none, goose, or migrate")
	cmd.Flags().BoolVar(&opts.docker, "docker", false, "Include Dockerfile")
	cmd.Flags().BoolVar(&opts.compose, "compose", false, "Include Docker Compose")
	cmd.Flags().BoolVar(&opts.githubActions, "github-actions", false, "Include GitHub Actions workflow")
	cmd.Flags().BoolVar(&opts.gitlabCI, "gitlab-ci", false, "Include GitLab CI pipeline")
	cmd.Flags().BoolVar(&opts.azurePipelines, "azure-pipelines", false, "Include Azure Pipelines workflow")
	cmd.Flags().BoolVar(&opts.health, "health", false, "Include health endpoint")
	cmd.Flags().BoolVar(&opts.readiness, "readiness", false, "Include readiness endpoint")
	cmd.Flags().StringVarP(&opts.outDir, "out", "o", "", "Directory to write generated project files")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Print generated files without writing them")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing generated target files")
	cmd.Flags().BoolVar(&opts.skipGoModTidy, "skip-go-mod-tidy", false, "Skip go mod tidy after generation")
	cmd.Flags().BoolVar(&opts.skipGitInit, "skip-git-init", false, "Skip git repository initialization after generation")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run without interactive prompts")
	cmd.Flags().BoolVar(&opts.overwrite, "overwrite", false, "Overwrite an existing saved recipe without prompting")
	return cmd
}

func runNew(out io.Writer, globalOpts *globalOptions, opts *newOptions, args []string) error {
	if globalOpts == nil {
		globalOpts = &globalOptions{}
	}
	if opts == nil {
		opts = &newOptions{}
	}
	if opts.recipePath != "" && opts.nonInteractive {
		r, err := recipe.Load(opts.recipePath)
		if err != nil {
			return err
		}
		return generateRecipe(out, r, generationOptionsFromNew(opts))
	}
	if opts.nonInteractive {
		if len(args) == 0 || args[0] == "" {
			return errors.New("module path is required for non-interactive new")
		}

		r := recipeFromNewOptions(args[0], opts)
		return generateRecipe(out, r, generationOptionsFromNew(opts))
	}

	return runInteractiveNew(out, globalOpts, opts, args)
}

func runInteractiveNew(out io.Writer, globalOpts *globalOptions, opts *newOptions, args []string) error {
	source, err := newBaseRecipe(opts)
	if err != nil {
		return err
	}
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		source.Project.Module = strings.TrimSpace(args[0])
		source.Project.Name = moduleBasename(source.Project.Module)
	}

	state := tui.NewConfigureWizardState(source, tui.ConfigureWizardOptions{
		RecipePath:    interactiveNewRecipePath(opts),
		OutputDir:     opts.outDir,
		Overwrite:     opts.overwrite,
		Force:         opts.force,
		SkipGoModTidy: opts.skipGoModTidy,
		Mode:          tui.ConfigureWizardModeGeneration,
	})
	app := tui.NewConfigureApp(state, tui.AppOptions{
		In:      os.Stdin,
		Out:     out,
		NoColor: globalOpts.NoColor,
	})
	if err := app.Run(); err != nil {
		if errors.Is(err, tui.ErrCanceled) {
			return nil
		}
		return err
	}

	switch state.Action {
	case tui.ConfigureWizardActionGenerate:
		generationOpts := generationOptionsFromNew(opts)
		if generationOpts.outDir == "" {
			generationOpts.outDir = state.OutputDirectory()
		}
		generationOpts.force = generationOpts.force || state.ForceOverwrite
		if err := generateRecipe(out, state.Recipe(), generationOpts); err != nil {
			return fmt.Errorf("generate project in %q: %w", generationOpts.outDir, err)
		}
		return nil
	case tui.ConfigureWizardActionSave:
		if state.Saved() {
			_, err := fmt.Fprintf(out, "saved recipe: %s\n", state.RecipePath())
			return err
		}
	}
	return nil
}

func newBaseRecipe(opts *newOptions) (*recipe.Recipe, error) {
	if opts != nil && opts.recipePath != "" {
		return recipe.Load(opts.recipePath)
	}
	preset := recipe.PresetWebBasic
	if opts != nil && opts.preset != "" {
		preset = opts.preset
	}
	return recipe.NewPreset(preset)
}

func interactiveNewRecipePath(opts *newOptions) string {
	if opts != nil && opts.recipePath != "" {
		return opts.recipePath
	}
	return "crego.yaml"
}

func recipeFromNewOptions(module string, opts *newOptions) *recipe.Recipe {
	r := &recipe.Recipe{
		Version: recipe.VersionV1,
		Project: recipe.ProjectConfig{
			Name:   moduleBasename(module),
			Module: module,
			Type:   opts.projectType,
		},
		Layout: recipe.LayoutConfig{
			Style: opts.layout,
		},
		Server: recipe.ServerConfig{
			Framework: opts.server,
		},
		Configuration: recipe.ConfigurationConfig{
			Format: opts.configuration,
		},
		Database: recipe.DatabaseConfig{
			Driver:     opts.database,
			Framework:  opts.framework,
			Migrations: opts.migrations,
		},
		Logging: recipe.LoggingConfig{
			Framework: opts.logging,
		},
		Observability: recipe.ObservabilityConfig{
			Health:    opts.health,
			Readiness: opts.readiness,
		},
		Deployment: recipe.DeploymentConfig{
			Docker:  opts.docker,
			Compose: opts.compose,
		},
		CI: recipe.CIConfig{
			GitHubActions:  opts.githubActions,
			GitLabCI:       opts.gitlabCI,
			AzurePipelines: opts.azurePipelines,
		},
	}
	if r.Project.Type == "" {
		r.Project.Type = recipe.ProjectTypeWeb
	}
	return r
}

func generationOptionsFromNew(opts *newOptions) generateOptions {
	return generateOptions{
		outDir:         opts.outDir,
		dryRun:         opts.dryRun,
		force:          opts.force,
		skipGoModTidy:  opts.skipGoModTidy,
		skipGitInit:    opts.skipGitInit,
		nonInteractive: opts.nonInteractive,
	}
}

func generateRecipe(out io.Writer, r *recipe.Recipe, opts generateOptions) error {
	recipe.Normalize(r)
	recipe.ApplyDefaults(r)
	if err := recipe.Validate(r); err != nil {
		return err
	}
	outDir := opts.outDir
	if outDir == "" {
		outDir = defaultOutputDirectory(r)
	}

	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		return err
	}
	result, err := generator.NewGenerator(templatefs.FS).Generate(nil, r, plan, generator.Options{
		OutDir: outDir,
		DryRun: opts.dryRun,
		Force:  opts.force,
	})
	if err != nil {
		return err
	}
	if opts.dryRun {
		return writeGenerationPlan(out, plan, result)
	}
	if err := runPostGenerationHooks(outDir, opts); err != nil {
		return err
	}
	return writeSuccess(out, outDir)
}

func runPostGenerationHooks(outDir string, opts generateOptions) error {
	if !opts.skipGoModTidy {
		goModPath := filepath.Join(outDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			if err := runQuietCommand(outDir, "go", "mod", "tidy"); err != nil {
				return err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect go.mod: %w", err)
		}
	}
	if opts.skipGitInit {
		return nil
	}
	gitDir := filepath.Join(outDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect git repository: %w", err)
	}
	return runQuietCommand(outDir, "git", "init")
}

func runQuietCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, message)
}
