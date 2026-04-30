package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newRecipeCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recipe",
		Short: "Work with crego recipes",
		Long: `Work with crego recipes.

Recipes will describe project metadata, selected components, and generation
settings so projects can be reproduced deterministically.`,
		Example: `  crego recipe init
  crego recipe validate --config crego.yaml
  crego recipe print --config crego.yaml`,
		RunE: notImplementedRunE,
	}
	cmd.SetOut(out)
	cmd.AddCommand(
		notImplementedCommand(
			"init",
			"Create a starter recipe",
			`Create a starter crego recipe file.

This command will eventually write a minimal recipe that can be edited,
validated, and passed to crego generate.`,
			`  crego recipe init
  crego recipe init --config crego.yaml`,
			out,
		),
		notImplementedCommand(
			"validate",
			"Validate a recipe file",
			`Validate a crego recipe file.

This command will eventually check recipe syntax, schema compatibility, and
referenced components before generation.`,
			`  crego recipe validate --config crego.yaml`,
			out,
		),
		notImplementedCommand(
			"print",
			"Print the resolved recipe",
			`Print a resolved crego recipe.

This command will eventually render the recipe after defaults, configuration,
and component selections have been applied.`,
			`  crego recipe print --config crego.yaml`,
			out,
		),
	)
	return cmd
}
