package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newGenerateCommand(out io.Writer) *cobra.Command {
	return notImplementedCommand(
		"generate",
		"Generate a project from a recipe",
		`Generate a Go project from an existing crego recipe.

This command is intended for deterministic and CI-friendly project generation
once recipe execution is implemented.`,
		`  crego generate --config crego.yaml
  crego generate --config ./recipes/service.yaml --verbose`,
		out,
	)
}
