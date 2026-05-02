package cli

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/generator"
	"github.com/v0xpopuli/crego/internal/recipe"
	"github.com/v0xpopuli/crego/internal/tui"
)

type generateOptions struct {
	recipePath     string
	outDir         string
	dryRun         bool
	force          bool
	skipGoModTidy  bool
	skipGitInit    bool
	nonInteractive bool
}

func newGenerateCommand(out io.Writer, global *globalOptions) *cobra.Command {
	opts := &generateOptions{}
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a project from a recipe",
		Long: `Generate a Go project from an existing crego recipe.

The command loads and validates the recipe, resolves selected components, and
renders project files into the output directory.`,
		Example: `  crego generate --recipe crego.yaml
  crego generate --config crego.yaml --out ./orders-api
  crego generate --recipe ./recipes/service.yaml --out ./service
  crego generate --recipe crego.yaml --out ./service --dry-run
  crego generate --recipe crego.yaml --out ./service --force`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !cmd.Flags().Changed("recipe") && global != nil && global.Config != "" {
				opts.recipePath = global.Config
			}
			return runGenerate(out, global, opts)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVarP(&opts.recipePath, "recipe", "r", "crego.yaml", "Path to the recipe file")
	cmd.Flags().StringVarP(&opts.outDir, "out", "o", "", "Directory to write generated project files")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Print generated files without writing them")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing generated target files")
	cmd.Flags().BoolVar(&opts.skipGoModTidy, "skip-go-mod-tidy", false, "Skip go mod tidy after generation")
	cmd.Flags().BoolVar(&opts.skipGitInit, "skip-git-init", false, "Skip git repository initialization after generation")
	cmd.Flags().BoolVar(&opts.nonInteractive, "non-interactive", false, "Run without interactive prompts")
	return cmd
}

func runGenerate(out io.Writer, global *globalOptions, opts *generateOptions) error {
	recipePath := opts.recipePath
	if recipePath == "" && global != nil {
		recipePath = global.Config
	}
	if recipePath == "" {
		recipePath = "crego.yaml"
	}

	r, err := recipe.Load(recipePath)
	if err != nil {
		return err
	}

	return generateRecipe(out, r, *opts)
}

func defaultOutputDirectory(r *recipe.Recipe) string {
	if r == nil {
		return "."
	}
	if strings.TrimSpace(r.Project.Name) != "" {
		return r.Project.Name
	}
	module := strings.TrimSuffix(strings.TrimSpace(r.Project.Module), "/")
	if module == "" {
		return "."
	}
	name := path.Base(module)
	if name == "." || name == "/" {
		return "."
	}
	return name
}

func moduleBasename(module string) string {
	module = strings.TrimSuffix(strings.TrimSpace(module), "/")
	if module == "" {
		return ""
	}
	name := path.Base(module)
	if name == "." || name == "/" {
		return ""
	}
	return name
}

func writeSuccess(out io.Writer, outDir string) error {
	return writeSuccessSummary(out, generationSummary{
		GenerationSummary: tui.GenerationSummary{OutputDir: outDir},
	})
}

func writeSuccessSummary(out io.Writer, summary generationSummary) error {
	files := "unknown"
	if summary.FileCount > 0 {
		files = fmt.Sprintf("%d", summary.FileCount)
	}
	stack := summary.Stack
	if stack == "" {
		stack = "unknown"
	}
	elapsed := "unknown"
	if summary.Elapsed > 0 {
		elapsed = fmt.Sprintf("%.1fs", summary.Elapsed.Seconds())
	}

	_, err := fmt.Fprintf(out, `Project generated successfully.

Created: %s
Files:   %s
Stack:   %s
Time:    %s

Next steps:
  cd %s
  make test
  make run
`, summary.OutputDir, files, stack, elapsed, summary.OutputDir)
	return err
}

func recipeStackSummary(r *recipe.Recipe) string {
	if r == nil {
		return ""
	}
	values := make([]string, 0, 10)
	appendIf := func(value string) {
		value = strings.TrimSpace(value)
		if value != "" && value != recipe.DatabaseDriverNone && value != recipe.DatabaseFrameworkNone && value != recipe.DatabaseMigrationsNone && value != recipe.TaskSchedulerNone {
			values = append(values, value)
		}
	}
	appendIf(r.Project.Type)
	appendIf(r.Layout.Style)
	if r.Go.Version != "" {
		appendIf("Go " + r.Go.Version)
	}
	if r.Project.Type == recipe.ProjectTypeWeb {
		appendIf(r.Server.Framework)
	}
	appendIf(r.Configuration.Format)
	appendIf(r.Logging.Framework)
	drivers := recipe.DatabaseDrivers(r.Database)
	if len(drivers) > 0 {
		if r.Database.Framework != "" && r.Database.Framework != recipe.DatabaseFrameworkNone {
			appendIf(drivers[0] + "/" + r.Database.Framework)
		} else {
			appendIf(drivers[0])
		}
	}
	appendIf(r.Database.Migrations)
	appendIf(r.TaskScheduler)
	if r.Deployment.Docker {
		appendIf("docker")
	}
	if r.Deployment.Compose {
		appendIf("compose")
	}
	if r.CI.GitHubActions {
		appendIf("github actions")
	}
	if r.CI.GitLabCI {
		appendIf("gitlab ci")
	}
	if r.CI.AzurePipelines {
		appendIf("azure pipelines")
	}
	return strings.Join(values, " · ")
}

func writeGenerationPlan(out io.Writer, plan *generator.Plan, result *generator.Result) error {
	if _, err := fmt.Fprintln(out, "Generation plan"); err != nil {
		return err
	}
	if err := writeGenerationList(out, "Files", result.FilesPlanned); err != nil {
		return err
	}
	if plan == nil {
		return nil
	}
	modules := make([]string, 0, len(plan.GoModules))
	for _, module := range plan.GoModules {
		if module.Version == "" {
			modules = append(modules, module.Path)
			continue
		}
		modules = append(modules, module.Path+" "+module.Version)
	}
	if err := writeGenerationList(out, "Go modules", modules); err != nil {
		return err
	}
	hooks := make([]string, 0, len(plan.Hooks))
	for _, hook := range plan.Hooks {
		hooks = append(hooks, hook.Name)
	}
	return writeGenerationList(out, "Hooks", hooks)
}

func writeGenerationList(out io.Writer, label string, values []string) error {
	if _, err := fmt.Fprintf(out, "%s:\n", label); err != nil {
		return err
	}
	if len(values) == 0 {
		_, err := fmt.Fprintln(out, "  none")
		return err
	}
	for _, value := range values {
		if _, err := fmt.Fprintf(out, "  %s\n", value); err != nil {
			return err
		}
	}
	return nil
}
