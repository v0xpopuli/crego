package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newExplainCommand(out io.Writer) *cobra.Command {
	return notImplementedCommand(
		"explain",
		"Explain what crego would generate",
		`Explain what crego would generate for the selected recipe or configuration.

This command will eventually provide a dry-run style summary of planned files,
dependencies, commands, and component decisions.`,
		`  crego explain --config crego.yaml
  crego explain --config ./recipes/api.yaml --debug`,
		out,
	)
}
