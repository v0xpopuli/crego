package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var errNotImplemented = errors.New("command not implemented yet")

type (
	// VersionInfo describes the build metadata printed by the version command.
	VersionInfo struct {
		Version string
		Commit  string
		Built   string
	}

	globalOptions struct {
		NoColor bool
		Verbose bool
		Debug   bool
		Config  string
	}

	exitCoder interface {
		ExitCode() int
	}

	commandError struct {
		err      error
		exitCode int
		handled  bool
	}
)

func (e commandError) Error() string {
	return e.err.Error()
}

func (e commandError) Unwrap() error {
	return e.err
}

func (e commandError) ExitCode() int {
	return e.exitCode
}

func commandErrorWithExitCode(err error, exitCode int) error {
	return &commandError{
		err:      err,
		exitCode: exitCode,
		handled:  true,
	}
}

// ExitCode returns the process exit code represented by err.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var coder exitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}

	return 1
}

// ShouldPrintError reports whether the caller should render err after command execution.
func ShouldPrintError(err error) bool {
	if err == nil {
		return false
	}

	if commandErr, ok := errors.AsType[*commandError](err); ok && commandErr.handled {
		return false
	}

	return true
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
  crego recipe validate crego.yaml
  crego components list
  crego version`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       versionInfo.Version,
	}
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable colorized output")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&opts.Debug, "debug", false, "Enable debug output")
	cmd.PersistentFlags().StringVar(&opts.Config, "config", "", "Path to a crego recipe or configuration file")

	cmd.AddCommand(
		newNewCommand(out, opts),
		newConfigureCommand(out, opts),
		newGenerateCommand(out, opts),
		newRecipeCommand(out, errOut, opts),
		newComponentsCommand(out),
		newExplainCommand(out),
		newVersionCommand(versionInfo, out),
	)

	return cmd
}

func notImplementedRunE(_ *cobra.Command, _ []string) error {
	return errNotImplemented
}

func versionLine(info VersionInfo) string {
	info = normalizeVersionInfo(info)
	return fmt.Sprintf("version: %s\ncommit: %s\nbuilt: %s\n", info.Version, info.Commit, info.Built)
}
