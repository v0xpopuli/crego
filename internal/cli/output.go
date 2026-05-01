package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/generator"
)

var componentCategoryOrder = []string{
	component.CategoryProject,
	component.CategoryLayout,
	component.CategoryConfiguration,
	component.CategoryServer,
	component.CategoryConfiguration,
	component.CategorySQLDatabase,
	component.CategoryORMFramework,
	component.CategoryNoSQLDatabase,
	component.CategoryMigrations,
	component.CategoryTaskScheduler,
	component.CategoryLogging,
	component.CategoryObservability,
	component.CategoryDeployment,
	component.CategoryCI,
}

type (
	componentSummaryOutput struct {
		ID          string `json:"id"`
		Category    string `json:"category"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	componentDetailOutput struct {
		ID          string               `json:"id"`
		Category    string               `json:"category"`
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Requires    []string             `json:"requires"`
		Conflicts   []string             `json:"conflicts"`
		Files       []templateFileOutput `json:"files"`
		GoModules   []goModuleOutput     `json:"go_modules"`
		Hooks       []hookOutput         `json:"hooks"`
	}

	componentCategoryOutput struct {
		Category   string                   `json:"category"`
		Components []componentSummaryOutput `json:"components"`
	}

	componentsListOutput struct {
		Categories []componentCategoryOutput `json:"categories"`
	}

	templateFileOutput struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	goModuleOutput struct {
		Path    string `json:"path"`
		Version string `json:"version"`
	}

	hookOutput struct {
		Name string `json:"name"`
	}

	explainOutput struct {
		Recipe     string                  `json:"recipe"`
		Components []componentDetailOutput `json:"components"`
		Files      []templateFileOutput    `json:"files"`
		GoModules  []goModuleOutput        `json:"go_modules"`
		Hooks      []hookOutput            `json:"hooks"`
	}
)

func encodeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func publicComponentCategory(category string) string {
	if category == component.CategoryDatabaseFramework {
		return component.CategoryORMFramework
	}
	return category
}

func isKnownPublicComponentCategory(category string) bool {
	for _, known := range componentCategoryOrder {
		if category == known {
			return true
		}
	}
	return false
}

func formatPublicComponentCategories() string {
	return strings.Join(componentCategoryOrder, ", ")
}

func componentSummary(c component.Component) componentSummaryOutput {
	category := publicComponentCategory(c.Category)
	switch c.ID {
	case component.IDDatabaseNone, component.IDDatabasePostgres, component.IDDatabaseMySQL, component.IDDatabaseSQLite:
		category = component.CategorySQLDatabase
	case component.IDDatabaseRedis, component.IDDatabaseMongoDB:
		category = component.CategoryNoSQLDatabase
	}
	summary := componentSummaryOutput{
		ID:          c.ID,
		Category:    category,
		Name:        c.Name,
		Description: c.Description,
	}
	if c.ID == component.IDDatabaseNone {
		summary.Description = "Project without a sql database integration."
	}
	return summary
}

func componentDetail(c component.Component) componentDetailOutput {
	summary := componentSummary(c)
	return componentDetailOutput{
		ID:          c.ID,
		Category:    summary.Category,
		Name:        c.Name,
		Description: c.Description,
		Requires:    stringSlice(c.Requires),
		Conflicts:   stringSlice(c.Conflicts),
		Files:       templateFilesOutput(c.Files),
		GoModules:   goModulesOutput(c.GoModules),
		Hooks:       hooksOutput(c.Hooks),
	}
}

func templateFilesOutput(files []component.TemplateFile) []templateFileOutput {
	result := make([]templateFileOutput, 0, len(files))
	for _, file := range files {
		result = append(result, templateFileOutput{
			Source: file.Source,
			Target: file.Target,
		})
	}
	return result
}

func goModulesOutput(modules []component.GoModule) []goModuleOutput {
	result := make([]goModuleOutput, 0, len(modules))
	for _, module := range modules {
		result = append(result, goModuleOutput{
			Path:    module.Path,
			Version: module.Version,
		})
	}
	return result
}

func hooksOutput(hooks []component.Hook) []hookOutput {
	result := make([]hookOutput, 0, len(hooks))
	for _, hook := range hooks {
		result = append(result, hookOutput{
			Name: hook.Name,
		})
	}
	return result
}

func stringSlice(values []string) []string {
	return append([]string{}, values...)
}

func explainResult(recipePath string, plan *generator.Plan) explainOutput {
	result := explainOutput{
		Recipe:     recipePath,
		Components: make([]componentDetailOutput, 0),
		Files:      make([]templateFileOutput, 0),
		GoModules:  make([]goModuleOutput, 0),
		Hooks:      make([]hookOutput, 0),
	}
	if plan == nil {
		return result
	}

	result.Components = make([]componentDetailOutput, 0, len(plan.Components))
	for _, c := range plan.Components {
		result.Components = append(result.Components, componentDetail(c))
	}
	result.Files = templateFilesOutput(plan.Files)
	result.GoModules = goModulesOutput(plan.GoModules)
	result.Hooks = hooksOutput(plan.Hooks)
	return result
}

func writeStringList(out io.Writer, label string, values []string) error {
	if len(values) == 0 {
		_, err := fmt.Fprintf(out, "%s: none\n", label)
		return err
	}

	_, err := fmt.Fprintf(out, "%s: %s\n", label, strings.Join(values, ", "))
	return err
}
