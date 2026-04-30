package generator

import (
	"bytes"
	"context"
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
	funcs := templateFuncs(data.ComponentIDs)
	for _, file := range plan.Files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		content, err := renderTemplate(templates, file.Source, data, funcs)
		if err != nil {
			return nil, err
		}
		result = append(result, renderedFile{
			Source:  file.Source,
			Target:  file.Target,
			Content: content,
		})
	}
	return result, nil
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

func newRenderContext(source *recipe.Recipe, plan *Plan) renderContext {
	componentIDs := make([]string, 0, len(plan.Components))
	for _, current := range plan.Components {
		componentIDs = append(componentIDs, current.ID)
	}

	return renderContext{
		Recipe:       source,
		ProjectName:  source.Project.Name,
		ModulePath:   source.Project.Module,
		Components:   append([]component.Component(nil), plan.Components...),
		ComponentIDs: componentIDs,
	}
}

func templateFuncs(componentIDs []string) template.FuncMap {
	selected := make(map[string]struct{}, len(componentIDs))
	for _, id := range componentIDs {
		selected[id] = struct{}{}
	}

	return template.FuncMap{
		"hasComponent": func(id string) bool {
			_, ok := selected[id]
			return ok
		},
		"lower": strings.ToLower,
		"title": title,
	}
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
