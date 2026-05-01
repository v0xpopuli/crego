package generator

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"io/fs"
	"strings"
	"text/template"
	"unicode"

	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
)

type (
	renderContext struct {
		Recipe       *recipe.Recipe
		ProjectName  string
		ModulePath   string
		Components   []component.Component
		ComponentIDs []string
		GoModules    []component.GoModule
	}

	renderedFile struct {
		Source  string
		Target  string
		Content []byte
	}
)

func renderFiles(ctx context.Context, templates fs.FS, source *recipe.Recipe, plan *Plan) ([]renderedFile, error) {
	if plan == nil || len(plan.Files) == 0 {
		return []renderedFile{}, nil
	}

	result := make([]renderedFile, 0, len(plan.Files))
	data := newRenderContext(source, plan)
	funcs := templateFuncs(data.ComponentIDs, data.Recipe)
	for _, file := range plan.Files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		content, err := renderTemplate(templates, file.Source, data, funcs)
		if err != nil {
			return nil, err
		}
		target, err := renderTemplateText(file.Source, file.Target, data, funcs)
		if err != nil {
			return nil, err
		}
		content, err = formatRenderedContent(file.Source, target, content)
		if err != nil {
			return nil, err
		}
		result = append(result, renderedFile{
			Source:  file.Source,
			Target:  target,
			Content: content,
		})
	}
	return result, nil
}

func RenderFileTargets(source *recipe.Recipe, plan *Plan) ([]component.TemplateFile, error) {
	if plan == nil || len(plan.Files) == 0 {
		return []component.TemplateFile{}, nil
	}

	data := newRenderContext(source, plan)
	funcs := templateFuncs(data.ComponentIDs, data.Recipe)
	files := make([]component.TemplateFile, 0, len(plan.Files))
	for _, file := range plan.Files {
		target, err := renderTemplateText(file.Source, file.Target, data, funcs)
		if err != nil {
			return nil, err
		}
		files = append(files, component.TemplateFile{
			Source: file.Source,
			Target: target,
		})
	}
	return files, nil
}

func formatRenderedContent(source string, target string, content []byte) ([]byte, error) {
	if !strings.HasSuffix(target, ".go") {
		return content, nil
	}

	formatted, err := format.Source(content)
	if err != nil {
		return nil, &TemplateRenderError{Source: source, Err: fmt.Errorf("format target %q: %w", target, err)}
	}
	return formatted, nil
}

func renderTemplate(templates fs.FS, source string, data renderContext, funcs template.FuncMap) ([]byte, error) {
	raw, err := fs.ReadFile(templates, source)
	if err != nil {
		return nil, &TemplateRenderError{Source: source, Err: err}
	}

	tmpl, err := template.New(source).Option("missingkey=error").Funcs(funcs).Parse(string(raw))
	if err != nil {
		return nil, &TemplateRenderError{Source: source, Err: err}
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return nil, &TemplateRenderError{Source: source, Err: err}
	}
	return buffer.Bytes(), nil
}

func renderTemplateText(source string, raw string, data renderContext, funcs template.FuncMap) (string, error) {
	tmpl, err := template.New(source + " target").Option("missingkey=error").Funcs(funcs).Parse(raw)
	if err != nil {
		return "", &TemplateRenderError{Source: source, Err: err}
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", &TemplateRenderError{Source: source, Err: err}
	}
	return buffer.String(), nil
}

func newRenderContext(source *recipe.Recipe, plan *Plan) renderContext {
	resolved := *source
	recipe.Normalize(&resolved)
	recipe.ApplyDefaults(&resolved)

	componentIDs := make([]string, 0, len(plan.Components))
	for _, current := range plan.Components {
		componentIDs = append(componentIDs, current.ID)
	}

	return renderContext{
		Recipe:       &resolved,
		ProjectName:  resolved.Project.Name,
		ModulePath:   resolved.Project.Module,
		Components:   append([]component.Component(nil), plan.Components...),
		ComponentIDs: componentIDs,
		GoModules:    append([]component.GoModule(nil), plan.GoModules...),
	}
}

func templateFuncs(componentIDs []string, r *recipe.Recipe) template.FuncMap {
	selected := make(map[string]struct{}, len(componentIDs))
	for _, id := range componentIDs {
		selected[id] = struct{}{}
	}
	drivers := map[string]struct{}{}
	if r != nil {
		for _, driver := range recipe.DatabaseDrivers(r.Database) {
			drivers[driver] = struct{}{}
		}
	}

	return template.FuncMap{
		"hasComponent": func(id string) bool {
			_, ok := selected[id]
			return ok
		},
		"hasDatabase": func(driver string) bool {
			_, ok := drivers[driver]
			return ok
		},
		"anyDatabaseEnabled": func() bool {
			for driver := range drivers {
				if driver != recipe.DatabaseDriverNone {
					return true
				}
			}
			return false
		},
		"anySQLDatabaseEnabled": func() bool {
			for driver := range drivers {
				if sqlDatabase(driver) {
					return true
				}
			}
			return false
		},
		"databaseEnabled": databaseEnabled,
		"configTag":       configTag(r),
		"sqlDatabase":     sqlDatabase,
		"sqlMigrations":   sqlMigrations,
		"schedulerEnabled": func() bool {
			return r != nil && r.TaskScheduler == recipe.TaskSchedulerGocron
		},
		"gormDistributedLock": func() bool {
			return r != nil && r.TaskScheduler == recipe.TaskSchedulerGocron && r.Database.Framework == recipe.DatabaseFrameworkGORM && hasSQLDatabase(drivers)
		},
		"lower": strings.ToLower,
		"title": title,
	}
}

func configTag(r *recipe.Recipe) func(string) string {
	return func(name string) string {
		if r == nil {
			return ""
		}
		switch r.Configuration.Format {
		case recipe.ConfigurationFormatJSON:
			return fmt.Sprintf("`json:%q`", name)
		case recipe.ConfigurationFormatTOML:
			return fmt.Sprintf("`toml:%q`", name)
		case recipe.ConfigurationFormatYAML:
			return fmt.Sprintf("`yaml:%q`", name)
		default:
			return ""
		}
	}
}

func databaseEnabled(driver string) bool {
	return driver != "" && driver != recipe.DatabaseDriverNone
}

func sqlDatabase(driver string) bool {
	switch driver {
	case recipe.DatabaseDriverPostgres, recipe.DatabaseDriverMySQL, recipe.DatabaseDriverSQLite:
		return true
	default:
		return false
	}
}

func hasSQLDatabase(drivers map[string]struct{}) bool {
	for driver := range drivers {
		if sqlDatabase(driver) {
			return true
		}
	}
	return false
}

func sqlMigrations(migrations string) bool {
	return migrations == recipe.DatabaseMigrationsGoose || migrations == recipe.DatabaseMigrationsMigrate
}

func title(value string) string {
	words := strings.Fields(value)
	for index, word := range words {
		words[index] = titleWord(word)
	}
	return strings.Join(words, " ")
}

func titleWord(value string) string {
	var builder strings.Builder
	first := true
	for _, current := range value {
		if first {
			builder.WriteRune(unicode.ToTitle(current))
			first = false
			continue
		}
		builder.WriteRune(current)
	}
	return builder.String()
}
