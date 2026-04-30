package generator

import "github.com/v0xpopuli/crego/internal/component"

type Plan struct {
	Components []component.Component
	Files      []component.TemplateFile
	GoModules  []component.GoModule
	Hooks      []component.Hook
}
