package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newConfigureCommand(out io.Writer) *cobra.Command {
	return notImplementedCommand(
		"configure",
		"Configure crego defaults",
		`Configure crego defaults for future project creation.

This command will eventually manage local defaults such as author metadata,
preferred modules, output paths, and registry settings.`,
		`  crego configure
  crego configure --config ~/.config/crego/config.yaml`,
		out,
	)
}
