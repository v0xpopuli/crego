package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func newDoctorCommand(out io.Writer) *cobra.Command {
	return notImplementedCommand(
		"doctor",
		"Check the local environment",
		`Check whether the local environment is ready for crego.

This command will eventually inspect Go tooling, module settings, network
access, templates, and component registry availability.`,
		`  crego doctor
  crego doctor --verbose`,
		out,
	)
}
