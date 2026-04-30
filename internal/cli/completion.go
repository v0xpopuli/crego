package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newCompletionCommand(root *cobra.Command, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion scripts",
		Long:      "Generate shell completion scripts for crego.",
		Example:   "  crego completion zsh > ~/.zsh/completions/_crego",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(_ *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(out)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			case "powershell":
				return root.GenPowerShellCompletion(out)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
	cmd.SetOut(out)
	return cmd
}
