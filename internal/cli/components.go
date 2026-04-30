package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v0xpopuli/crego/internal/component"
)

type (
	componentsListOptions struct {
		category string
		json     bool
	}

	componentsShowOptions struct {
		json bool
	}
)

func newComponentsCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "components",
		Short: "Explore available project components",
		Long: `Explore available project components.

Components describe selectable project capabilities such as HTTP servers,
databases, telemetry, containerization, and CI setup.`,
		Example: `  crego components list
  crego components list --category server
  crego components show server.chi`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}
			return fmt.Errorf("components requires a subcommand: list or show")
		},
	}
	cmd.SetOut(out)
	cmd.AddCommand(
		newComponentsListCommand(out),
		newComponentsShowCommand(out),
	)
	return cmd
}

func newComponentsListCommand(out io.Writer) *cobra.Command {
	opts := &componentsListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available components",
		Long: `List available crego components.

Components are grouped by category. Database framework components are displayed
under the database category because they are selected as part of database setup.`,
		Example: `  crego components list
  crego components list --category server
  crego components list --category database
  crego components list --json`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runComponentsList(out, opts)
		},
	}
	cmd.SetOut(out)
	cmd.Flags().StringVar(&opts.category, "category", "", "Filter by category: "+formatPublicComponentCategories())
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print machine-readable JSON output")
	return cmd
}

func runComponentsList(out io.Writer, opts *componentsListOptions) error {
	if opts.category != "" && !isKnownPublicComponentCategory(opts.category) {
		return fmt.Errorf("unknown component category %q; allowed categories: %s", opts.category, formatPublicComponentCategories())
	}

	result := componentsListResult(component.NewRegistry().List(), opts.category)
	if opts.json {
		return encodeJSON(out, result)
	}

	return writeComponentsList(out, result)
}

func componentsListResult(components []component.Component, category string) componentsListOutput {
	grouped := make(map[string][]componentSummaryOutput, len(componentCategoryOrder))
	for _, c := range components {
		publicCategory := publicComponentCategory(c.Category)
		if category != "" && publicCategory != category {
			continue
		}
		grouped[publicCategory] = append(grouped[publicCategory], componentSummary(c))
	}

	result := componentsListOutput{
		Categories: make([]componentCategoryOutput, 0, len(componentCategoryOrder)),
	}
	for _, currentCategory := range componentCategoryOrder {
		if category != "" && currentCategory != category {
			continue
		}
		summaries := grouped[currentCategory]
		if summaries == nil {
			summaries = []componentSummaryOutput{}
		}
		result.Categories = append(result.Categories, componentCategoryOutput{
			Category:   currentCategory,
			Components: summaries,
		})
	}
	return result
}

func writeComponentsList(out io.Writer, result componentsListOutput) error {
	for categoryIndex, category := range result.Categories {
		if categoryIndex > 0 {
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(out, "%s:\n", category.Category); err != nil {
			return err
		}
		if len(category.Components) == 0 {
			if _, err := fmt.Fprintln(out, "  none"); err != nil {
				return err
			}
			continue
		}
		currentGroup := ""
		for _, c := range category.Components {
			group, id := componentListDisplayParts(category.Category, c.ID)
			if group == "" {
				currentGroup = ""
			}
			if group != "" && group != currentGroup {
				if _, err := fmt.Fprintf(out, "  %s:\n", group); err != nil {
					return err
				}
				currentGroup = group
			}

			indent := "  "
			if group != "" {
				indent = "    "
			}
			if _, err := fmt.Fprintf(out, "%s%s - %s\n", indent, id, c.Description); err != nil {
				return err
			}
		}
	}
	return nil
}

func componentListDisplayParts(category string, id string) (string, string) {
	prefix := category + "."
	if !strings.HasPrefix(id, prefix) {
		return "", id
	}

	trimmed := strings.TrimPrefix(id, prefix)
	group, name, ok := strings.Cut(trimmed, ".")
	if !ok {
		return "", trimmed
	}
	return group, name
}

func newComponentsShowCommand(out io.Writer) *cobra.Command {
	opts := &componentsShowOptions{}
	cmd := &cobra.Command{
		Use:   "show <component-id>",
		Short: "Show component details",
		Long: `Show details for a crego component.

Details include compatibility metadata, planned files, Go modules, hooks, and
whether generation artifacts are implemented yet.`,
		Example: `  crego components show server.chi
  crego components show server.gin
  crego components show database.postgres
  crego components show database.framework.gorm --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runComponentsShow(out, opts, args[0])
		},
	}
	cmd.SetOut(out)
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print machine-readable JSON output")
	return cmd
}

func runComponentsShow(out io.Writer, opts *componentsShowOptions, id string) error {
	c, ok := component.NewRegistry().Get(id)
	if !ok {
		return &component.UnknownComponentError{ID: id}
	}

	result := componentDetail(c)
	if opts.json {
		return encodeJSON(out, result)
	}

	return writeComponentDetail(out, result)
}

func writeComponentDetail(out io.Writer, c componentDetailOutput) error {
	if _, err := fmt.Fprintf(out, "id: %s\n", c.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "category: %s\n", c.Category); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "name: %s\n", c.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "description: %s\n", c.Description); err != nil {
		return err
	}
	if err := writeStringList(out, "requires", c.Requires); err != nil {
		return err
	}
	if err := writeStringList(out, "conflicts", c.Conflicts); err != nil {
		return err
	}
	if err := writeTemplateFiles(out, c.Files); err != nil {
		return err
	}
	if err := writeGoModules(out, c.GoModules); err != nil {
		return err
	}
	if err := writeHooks(out, c.Hooks); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "support_status: %s\n", c.SupportStatus); err != nil {
		return err
	}
	if c.SupportNote != "" {
		if _, err := fmt.Fprintf(out, "support_note: %s\n", c.SupportNote); err != nil {
			return err
		}
	}
	return nil
}

func writeTemplateFiles(out io.Writer, files []templateFileOutput) error {
	if len(files) == 0 {
		_, err := fmt.Fprintln(out, "files: none")
		return err
	}

	if _, err := fmt.Fprintln(out, "files:"); err != nil {
		return err
	}
	for _, file := range files {
		if _, err := fmt.Fprintf(out, "  %s -> %s\n", file.Source, file.Target); err != nil {
			return err
		}
	}
	return nil
}

func writeGoModules(out io.Writer, modules []goModuleOutput) error {
	if len(modules) == 0 {
		_, err := fmt.Fprintln(out, "go_modules: none")
		return err
	}

	if _, err := fmt.Fprintln(out, "go_modules:"); err != nil {
		return err
	}
	for _, module := range modules {
		if _, err := fmt.Fprintf(out, "  %s %s\n", module.Path, module.Version); err != nil {
			return err
		}
	}
	return nil
}

func writeHooks(out io.Writer, hooks []hookOutput) error {
	if len(hooks) == 0 {
		_, err := fmt.Fprintln(out, "hooks: none")
		return err
	}

	if _, err := fmt.Fprintln(out, "hooks:"); err != nil {
		return err
	}
	for _, hook := range hooks {
		if _, err := fmt.Fprintf(out, "  %s\n", hook.Name); err != nil {
			return err
		}
	}
	return nil
}
