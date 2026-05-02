package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/recipe"
	"github.com/v0xpopuli/crego/internal/tui"
)

type configureOptions struct {
	recipePath string
	preset     string
	minimal    bool
	overwrite  bool
}

func newConfigureCommand(out io.Writer, globalOpts *globalOptions) *cobra.Command {
	opts := &configureOptions{}
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Create a crego recipe with an interactive wizard",
		Long: `Create a crego recipe with an interactive terminal wizard.

The wizard asks practical project questions, previews the normalized recipe and
resolved generation plan, and writes a reusable crego.yaml file.`,
		Example: `  crego configure
  crego configure --recipe web-postgres.yaml
  crego configure --preset web-postgres
  crego configure --recipe web-postgres.yaml --preset web-postgres`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigure(out, globalOpts, opts)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVarP(&opts.recipePath, "recipe", "r", "crego.yaml", "Path to write the recipe file")
	cmd.Flags().StringVar(&opts.preset, "preset", "", "Starter preset: web-basic, web-postgres, web-mysql, web-sqlite, web-redis, web-mongodb, cli-basic")
	cmd.Flags().BoolVar(&opts.minimal, "minimal", false, "Start from a minimal recipe selection")
	cmd.Flags().BoolVar(&opts.overwrite, "overwrite", false, "Overwrite an existing recipe file without prompting")
	return cmd
}

func runConfigure(out io.Writer, globalOpts *globalOptions, opts *configureOptions) error {
	if globalOpts == nil {
		globalOpts = &globalOptions{}
	}
	if opts == nil {
		opts = &configureOptions{recipePath: "crego.yaml"}
	}
	if opts.recipePath == "" {
		opts.recipePath = "crego.yaml"
	}

	source, err := configureBaseRecipe(opts)
	if err != nil {
		return err
	}

	state := tui.NewConfigureWizardState(source, tui.ConfigureWizardOptions{
		RecipePath: opts.recipePath,
		Minimal:    opts.minimal,
		Overwrite:  opts.overwrite,
	})

	app := tui.NewConfigureApp(state, tui.AppOptions{
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
	if state.Saved() {
		_, err := fmt.Fprintf(out, "saved recipe: %s\n", state.RecipePath())
		return err
	}
	return nil
}

func configureBaseRecipe(opts *configureOptions) (*recipe.Recipe, error) {
	preset := recipe.PresetWebBasic
	if opts != nil && opts.preset != "" {
		preset = opts.preset
	} else if opts != nil && opts.minimal {
		preset = recipe.PresetCLIBasic
	}

	r, err := recipe.NewPreset(preset)
	if err != nil {
		return nil, err
	}
	if opts != nil && opts.minimal {
		r.Layout.Style = recipe.LayoutStyleMinimal
		r.Database.Driver = recipe.DatabaseDriverNone
		r.Database.Framework = recipe.DatabaseFrameworkNone
		r.Database.Migrations = recipe.DatabaseMigrationsNone
		r.Observability.Health = false
		r.Observability.Readiness = false
		r.Deployment.Docker = false
		r.Deployment.Compose = false
		r.CI.GitHubActions = false
		r.CI.GitLabCI = false
		r.CI.AzurePipelines = false
	}
	return r, nil
}
