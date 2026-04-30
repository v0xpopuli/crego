package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var errNotImplemented = errors.New("command not implemented yet")

// VersionInfo describes the build metadata printed by the version command.
type VersionInfo struct {
	Version string
	Commit  string
	Built   string
}

type globalOptions struct {
	NoColor bool
	Verbose bool
	Debug   bool
	Config  string
}

func normalizeVersionInfo(info VersionInfo) VersionInfo {
	if info.Version == "" {
		info.Version = "dev"
	}
	if info.Commit == "" {
		info.Commit = "unknown"
	}
	if info.Built == "" {
		info.Built = "unknown"
	}
	return info
}

// NewRootCommand creates the crego CLI with production stdout/stderr writers.
func NewRootCommand(versionInfo VersionInfo) *cobra.Command {
	return NewRootCommandWithWriters(versionInfo, os.Stdout, os.Stderr)
}

// NewRootCommandWithWriters creates the crego CLI with injected writers for tests.
func NewRootCommandWithWriters(versionInfo VersionInfo, out io.Writer, errOut io.Writer) *cobra.Command {
	versionInfo = normalizeVersionInfo(versionInfo)
	opts := &globalOptions{}

	cmd := &cobra.Command{
		Use:   "crego",
		Short: "Create Go projects from interactive prompts or deterministic recipes",
		Long: `crego is a TUI-first Go project generator.

It is interactive by default, deterministic by recipe, and scriptable for CI.
Generation, recipe schema, and component registry behavior will be added in later milestones.`,
		Example: `  crego new
  crego configure
  crego generate --config crego.yaml
  crego recipe validate --config crego.yaml
  crego components list
  crego version`,
		SilenceUsage: true,
		Version:      versionInfo.Version,
	}
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable colorized output")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&opts.Debug, "debug", false, "Enable debug output")
	cmd.PersistentFlags().StringVar(&opts.Config, "config", "", "Path to a crego recipe or configuration file")

	cmd.AddCommand(
		newNewCommand(out),
		newConfigureCommand(out),
		newGenerateCommand(out),
		newRecipeCommand(out),
		newComponentsCommand(out),
		newExplainCommand(out),
		newDoctorCommand(out),
		newVersionCommand(versionInfo, out),
		newCompletionCommand(cmd, out),
	)

	return cmd
}

func notImplementedRunE(_ *cobra.Command, _ []string) error {
	return errNotImplemented
}

func notImplementedCommand(use string, short string, long string, example string, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Example: example,
		RunE:    notImplementedRunE,
	}
	cmd.SetOut(out)
	return cmd
}

func versionLine(info VersionInfo) string {
	info = normalizeVersionInfo(info)
	return fmt.Sprintf("version: %s\ncommit: %s\nbuilt: %s\n", info.Version, info.Commit, info.Built)
}
