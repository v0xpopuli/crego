package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newComponentsCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "components",
		Short: "Explore available project components",
		Long: `Explore available project components.

Components will represent optional project capabilities such as HTTP servers,
databases, telemetry, containerization, and CI setup.`,
		Example: `  crego components list
  crego components show http-server`,
		RunE: notImplementedRunE,
	}
	cmd.SetOut(out)
	cmd.AddCommand(
		notImplementedCommand(
			"list",
			"List available components",
			`List available crego components.

This command will eventually read the component registry and display component
IDs, descriptions, and compatibility metadata.`,
			`  crego components list
  crego components list --verbose`,
			out,
		),
		notImplementedCommand(
			"show <component>",
			"Show component details",
			`Show details for a crego component.

This command will eventually display inputs, generated files, dependencies, and
compatibility notes for a selected component.`,
			`  crego components show http-server
  crego components show postgres --config crego.yaml`,
			out,
		),
	)
	return cmd
}
