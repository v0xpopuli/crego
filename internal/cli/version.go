package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newVersionCommand(versionInfo VersionInfo, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print crego build information",
		Long: `Print crego build information.

The output includes the semantic version, commit identifier, and build time.`,
		Example: `  crego version`,
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := fmt.Fprint(out, versionLine(versionInfo))
			return err
		},
	}
}
