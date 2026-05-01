package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/recipe"
	"gopkg.in/yaml.v3"
)

const recipeValidationExitCode = 3

type (
	recipeInitOptions struct {
		outPath   string
		preset    string
		module    string
		name      string
		overwrite bool
	}

	recipeValidateOptions struct {
		strict bool
		json   bool
	}

	recipeValidateResult struct {
		Valid    bool     `json:"valid"`
		Errors   []string `json:"errors"`
		Warnings []string `json:"warnings"`
	}

	recipePrintOptions struct {
		json       bool
		noComments bool
	}
)

func newRecipeCommand(out io.Writer, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recipe",
		Short: "Work with crego recipes",
		Long: `Work with crego recipes.

Recipes will describe project metadata, selected components, and generation
settings so projects can be reproduced deterministically.`,
		Example: `  crego recipe init
  crego recipe init --preset web-postgres --module github.com/acme/orders
  crego recipe validate crego.yaml
  crego recipe print crego.yaml --json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			return fmt.Errorf("recipe requires a subcommand: init, validate, or print")
		},
	}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.AddCommand(
		newRecipeInitCommand(out),
		newRecipeValidateCommand(out),
		newRecipePrintCommand(out),
	)
	return cmd
}

func newRecipeInitCommand(out io.Writer) *cobra.Command {
	opts := &recipeInitOptions{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a starter recipe",
		Long: `Create a starter crego recipe file.

The generated recipe is normalized and can be validated, reviewed, or passed to
future generation commands.`,
		Example: `  crego recipe init
  crego recipe init --preset web-postgres --module github.com/acme/orders
  crego recipe init --preset cli-basic --name worker-tools --out recipe.yaml
  crego recipe init --out crego.yaml --overwrite`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRecipeInit(out, opts)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVarP(&opts.outPath, "out", "o", "crego.yaml", "Path to write the recipe file")
	cmd.Flags().StringVar(&opts.preset, "preset", recipe.PresetWebBasic, "Starter preset: web-basic, web-postgres, web-mysql, web-sqlite, web-redis, web-mongodb, cli-basic")
	cmd.Flags().StringVar(&opts.module, "module", "", "Go module path to set in project.module")
	cmd.Flags().StringVar(&opts.name, "name", "", "Project name to set in project.name")
	cmd.Flags().BoolVar(&opts.overwrite, "overwrite", false, "Overwrite an existing recipe file")
	return cmd
}

func runRecipeInit(out io.Writer, opts *recipeInitOptions) error {
	r, err := recipe.NewPreset(opts.preset)
	if err != nil {
		return err
	}

	if opts.module != "" {
		r.Project.Module = opts.module
		if opts.name == "" {
			r.Project.Name = moduleBaseName(opts.module)
		}
	}
	if opts.name != "" {
		r.Project.Name = opts.name
	}

	if !opts.overwrite {
		if _, err := os.Stat(opts.outPath); err == nil {
			return fmt.Errorf("recipe file %q already exists; pass --overwrite to replace it", opts.outPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check recipe output %q: %w", opts.outPath, err)
		}
	}

	if err := recipe.Save(opts.outPath, r); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "created recipe: %s\n", opts.outPath)
	return err
}

func newRecipeValidateCommand(out io.Writer) *cobra.Command {
	opts := &recipeValidateOptions{}
	cmd := &cobra.Command{
		Use:   "validate [recipe]",
		Short: "Validate a recipe file",
		Long: `Validate a crego recipe file.

Validation loads the recipe, applies defaults, normalizes enum-like fields, and
checks schema and compatibility rules before generation.`,
		Example: `  crego recipe validate
  crego recipe validate crego.yaml
  crego recipe validate crego.yaml --json
  crego recipe validate recipe.yaml --strict`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRecipeValidate(out, opts, recipePathArg(args))
		},
	}
	cmd.SetOut(out)
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Treat warnings as validation errors")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print machine-readable JSON validation output")
	return cmd
}

func runRecipeValidate(out io.Writer, opts *recipeValidateOptions, recipePath string) error {
	warnings := []string{}
	_, err := recipe.Load(recipePath)
	result := recipeValidateResult{
		Valid:    err == nil,
		Errors:   recipeErrorMessages(err),
		Warnings: warnings,
	}
	if opts.strict && len(warnings) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "recipe warnings rejected by --strict")
	}

	if opts.json {
		if encodeErr := json.NewEncoder(out).Encode(result); encodeErr != nil {
			return encodeErr
		}
	} else if result.Valid {
		if _, writeErr := fmt.Fprintf(out, "recipe valid: %s\n", recipePath); writeErr != nil {
			return writeErr
		}
	} else {
		if _, writeErr := fmt.Fprintf(out, "recipe invalid: %s\n%s\n", recipePath, err); writeErr != nil {
			return writeErr
		}
	}

	if !result.Valid {
		if err == nil {
			err = fmt.Errorf("recipe validation failed")
		}
		return commandErrorWithExitCode(err, recipeValidationExitCode)
	}

	return nil
}

func newRecipePrintCommand(out io.Writer) *cobra.Command {
	opts := &recipePrintOptions{}
	cmd := &cobra.Command{
		Use:   "print [recipe]",
		Short: "Print the normalized recipe",
		Long: `Print the normalized crego recipe.

The printed recipe has defaults applied and enum-like values normalized, making
it suitable for review and deterministic generation.`,
		Example: `  crego recipe print
  crego recipe print crego.yaml
  crego recipe print crego.yaml --json
  crego recipe print recipe.yaml --no-comments`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRecipePrint(out, opts, recipePathArg(args))
		},
	}
	cmd.SetOut(out)
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print normalized recipe as JSON")
	cmd.Flags().BoolVar(&opts.noComments, "no-comments", false, "Omit comments from YAML output")
	return cmd
}

func runRecipePrint(out io.Writer, opts *recipePrintOptions, recipePath string) error {
	r, err := recipe.Load(recipePath)
	if err != nil {
		return err
	}

	if opts.json {
		mapped, err := yamlTaggedRecipe(r)
		if err != nil {
			return err
		}

		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(mapped)
	}

	data, err := recipe.MarshalYAML(r)
	if err != nil {
		return fmt.Errorf("print recipe %q: %w", recipePath, err)
	}
	_, err = out.Write(data)
	return err
}

func recipePathArg(args []string) string {
	if len(args) == 0 {
		return "crego.yaml"
	}

	return args[0]
}

func recipeErrorMessages(err error) []string {
	if err == nil {
		return []string{}
	}

	var validationErr *recipe.ValidationError
	if errors.As(err, &validationErr) {
		return append([]string(nil), validationErr.Problems...)
	}

	return []string{err.Error()}
}

func moduleBaseName(module string) string {
	module = strings.TrimSpace(strings.TrimRight(module, "/"))
	if module == "" {
		return ""
	}

	return path.Base(module)
}

func yamlTaggedRecipe(r *recipe.Recipe) (map[string]any, error) {
	data, err := recipe.MarshalYAML(r)
	if err != nil {
		return nil, err
	}

	var mapped map[string]any
	if err := yaml.Unmarshal(data, &mapped); err != nil {
		return nil, err
	}

	return mapped, nil
}
