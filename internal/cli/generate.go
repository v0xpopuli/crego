package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/generator"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
)

type generateOptions struct {
	recipePath string
	outDir     string
	dryRun     bool
	force      bool
}

func newGenerateCommand(out io.Writer, global *globalOptions) *cobra.Command {
	opts := &generateOptions{}
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a project from a recipe",
		Long: `Generate a Go project from an existing crego recipe.

The command loads and validates the recipe, resolves selected components, and
renders project files into the output directory.`,
		Example: `  crego generate --config crego.yaml --out ./orders-api
  crego generate --recipe ./recipes/service.yaml --out ./service
  crego generate --recipe crego.yaml --out ./service --dry-run
  crego generate --recipe crego.yaml --out ./service --force`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runGenerate(out, global, opts)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVarP(&opts.recipePath, "recipe", "r", "", "Path to the recipe file")
	cmd.Flags().StringVarP(&opts.outDir, "out", "o", ".", "Directory to write generated project files")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Print generated files without writing them")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing generated target files")
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

	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		return err
	}

	result, err := generator.NewGenerator(templatefs.FS).Generate(nil, r, plan, generator.Options{
		OutDir: opts.outDir,
		DryRun: opts.dryRun,
		Force:  opts.force,
	})
	if err != nil {
		return err
	}

	if opts.dryRun {
		return writeGeneratedFiles(out, "planned files", result.FilesPlanned)
	}
	return writeGeneratedFiles(out, "generated files", result.FilesWritten)
}

func writeGeneratedFiles(out io.Writer, label string, files []string) error {
	if _, err := fmt.Fprintf(out, "%s:\n", label); err != nil {
		return err
	}
	if len(files) == 0 {
		_, err := fmt.Fprintln(out, "  none")
		return err
	}
	for _, file := range files {
		if _, err := fmt.Fprintf(out, "  %s\n", file); err != nil {
			return err
		}
	}
	return nil
}
