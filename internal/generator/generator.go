package generator

import (
	"context"
	"errors"
	"io/fs"

	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
)

type (
	Options struct {
		OutDir string
		DryRun bool
		Force  bool
	}

	Result struct {
		FilesWritten []string
		FilesPlanned []string
		SkippedHooks []string
	}

	Generator struct {
		Templates fs.FS
	}
)

func NewGenerator(templates fs.FS) *Generator {
	if templates == nil {
		templates = templatefs.FS
	}
	return &Generator{Templates: templates}
}

func (g *Generator) Generate(ctx context.Context, source *recipe.Recipe, plan *Plan, opts Options) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if source == nil {
		return nil, errors.New("generator recipe is required")
	}
	if opts.OutDir == "" {
		return nil, &OutDirRequiredError{}
	}
	if plan == nil {
		plan = &Plan{}
	}
	var templates fs.FS = templatefs.FS
	if g != nil && g.Templates != nil {
		templates = g.Templates
	}

	result := &Result{
		FilesWritten: []string{},
		FilesPlanned: []string{},
		SkippedHooks: skippedHooks(plan),
	}

	rendered, err := renderFiles(ctx, templates, source, plan)
	if err != nil {
		return nil, err
	}

	if opts.DryRun {
		targets, _, err := planFiles(opts.OutDir, rendered)
		if err != nil {
			return nil, err
		}
		result.FilesPlanned = targets
		return result, nil
	}

	written, err := writeFiles(ctx, opts.OutDir, rendered, opts.Force)
	if err != nil {
		return nil, err
	}
	result.FilesWritten = written
	return result, nil
}

func skippedHooks(plan *Plan) []string {
	hooks := make([]string, 0, len(plan.Hooks))
	for _, hook := range plan.Hooks {
		hooks = append(hooks, hook.Name)
	}
	return hooks
}
