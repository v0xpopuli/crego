package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/generator"
	"github.com/v0xpopuli/crego/internal/recipe"
)

type explainOptions struct {
	recipePath string
	json       bool
}

func newExplainCommand(out io.Writer) *cobra.Command {
	opts := &explainOptions{}
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Explain what crego would generate",
		Long: `Explain what crego would generate for a recipe.

The command loads and validates the recipe, resolves selected components, and
prints a generation plan without rendering templates or writing project files.`,
		Example: `  crego explain
  crego explain --recipe crego.yaml
  crego explain --recipe ./recipes/api.yaml --json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runExplain(out, opts)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVarP(&opts.recipePath, "recipe", "r", "crego.yaml", "Path to the recipe file")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print machine-readable JSON output")
	return cmd
}

func runExplain(out io.Writer, opts *explainOptions) error {
	r, err := recipe.Load(opts.recipePath)
	if err != nil {
		return err
	}

	plan, err := generator.Resolve(component.NewRegistry(), r)
	if err != nil {
		return err
	}

	result := explainResult(opts.recipePath, plan)
	if opts.json {
		return encodeJSON(out, result)
	}

	return writeExplain(out, result)
}

func writeExplain(out io.Writer, result explainOutput) error {
	if _, err := fmt.Fprintf(out, "recipe: %s\n", result.Recipe); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "\nselected components:"); err != nil {
		return err
	}
	if len(result.Components) == 0 {
		if _, err := fmt.Fprintln(out, "  none"); err != nil {
			return err
		}
	}
	for _, c := range result.Components {
		if _, err := fmt.Fprintf(out, "  %s (%s) - %s\n", c.ID, c.Category, c.Description); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(out, "\ngenerated files:"); err != nil {
		return err
	}
	if err := writeExplainFiles(out, result.Files); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "\ngo modules:"); err != nil {
		return err
	}
	if err := writeExplainGoModules(out, result.GoModules); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "\nhooks:"); err != nil {
		return err
	}
	return writeExplainHooks(out, result.Hooks)
}

func writeExplainFiles(out io.Writer, files []templateFileOutput) error {
	if len(files) == 0 {
		_, err := fmt.Fprintln(out, "  none")
		return err
	}
	for _, file := range files {
		if _, err := fmt.Fprintf(out, "  %s <- %s\n", file.Target, file.Source); err != nil {
			return err
		}
	}
	return nil
}

func writeExplainGoModules(out io.Writer, modules []goModuleOutput) error {
	if len(modules) == 0 {
		_, err := fmt.Fprintln(out, "  none")
		return err
	}
	for _, module := range modules {
		if _, err := fmt.Fprintf(out, "  %s %s\n", module.Path, module.Version); err != nil {
			return err
		}
	}
	return nil
}

func writeExplainHooks(out io.Writer, hooks []hookOutput) error {
	if len(hooks) == 0 {
		_, err := fmt.Fprintln(out, "  none")
		return err
	}
	for _, hook := range hooks {
		if _, err := fmt.Fprintf(out, "  %s\n", hook.Name); err != nil {
			return err
		}
	}
	return nil
}
