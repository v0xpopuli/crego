package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newNewCommand(out io.Writer) *cobra.Command {
	return notImplementedCommand(
		"new",
		"Start an interactive project setup",
		`Start crego's interactive project setup.

This command will eventually launch the TUI flow for selecting project metadata,
modules, components, and output settings.`,
		`  crego new
  crego new --config crego.yaml`,
		out,
	)
}
