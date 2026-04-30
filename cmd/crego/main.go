package main

import (
	"fmt"
	"os"

	"github.com/v0xpopuli/crego/internal/cli"
)

var (
	version = "dev"
	commit  = "unknown"
	built   = "unknown"
)

func main() {
	root := cli.NewRootCommand(cli.VersionInfo{
		Version: version,
		Commit:  commit,
		Built:   built,
	})

	if err := root.Execute(); err != nil {
		if cli.ShouldPrintError(err) {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(cli.ExitCode(err))
	}
}
