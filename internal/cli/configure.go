package cli

import (
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/tui"
)

func newConfigureCommand(out io.Writer, globalOpts *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure crego defaults",
		Long: `Configure crego defaults for future project creation.

This command currently launches the shared TUI foundation placeholder. The full
configure wizard will manage local defaults such as author metadata, preferred
modules, output paths, and registry settings.`,
		Example: `  crego configure
  crego configure --config ~/.config/crego/config.yaml`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigure(out, globalOpts)
		},
	}
	cmd.SetOut(out)
	return cmd
}

func runConfigure(out io.Writer, globalOpts *globalOptions) error {
	if globalOpts == nil {
		globalOpts = &globalOptions{}
	}

	app := tui.NewDemoApp(tui.AppOptions{
		In:      os.Stdin,
		Out:     out,
		NoColor: globalOpts.NoColor,
	})
	if err := app.Run(); err != nil {
		if errors.Is(err, tui.ErrCanceled) {
			return nil
		}
		return err
	}
	return nil
}
