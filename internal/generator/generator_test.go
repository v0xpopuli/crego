package generator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/component"
	"github.com/v0xpopuli/crego/internal/recipe"
	templatefs "github.com/v0xpopuli/crego/internal/templates"
)

type GeneratorTestSuite struct {
	suite.Suite
}

func TestGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(GeneratorTestSuite))
}

func (s *GeneratorTestSuite) TestRendersEmbeddedReadme() {
	outDir := s.T().TempDir()
	g := NewGenerator(templatefs.FS)

	result, err := g.Generate(context.Background(), generatorTestRecipe(), readmePlan("README.md"), Options{OutDir: outDir})

	s.Require().NoError(err)
	s.Require().Equal([]string{"README.md"}, result.FilesWritten)
	s.Require().Empty(result.FilesPlanned)

	content, err := os.ReadFile(filepath.Join(outDir, "README.md"))
	s.Require().NoError(err)
	s.Require().Contains(string(content), "# example")
	s.Require().Contains(string(content), "example.com/example")
}

func (s *GeneratorTestSuite) TestRenderFailureIncludesTemplatePath() {
	g := NewGenerator(fstest.MapFS{
		"bad.tmpl": &fstest.MapFile{Data: []byte("{{ .MissingField }}")},
	})
	plan := &Plan{Files: []component.TemplateFile{{Source: "bad.tmpl", Target: "README.md"}}}

	_, err := g.Generate(context.Background(), generatorTestRecipe(), plan, Options{OutDir: s.T().TempDir(), DryRun: true})

	s.Require().Error(err)
	var renderErr *TemplateRenderError
	s.Require().True(errors.As(err, &renderErr))
	s.Require().Equal("bad.tmpl", renderErr.Source)
	s.Require().Contains(err.Error(), "bad.tmpl")
}

func (s *GeneratorTestSuite) TestDryRunPlansFilesWithoutTouchingFilesystem() {
	outDir := s.T().TempDir()
	g := NewGenerator(templatefs.FS)
	plan := readmePlan("README.md")
	plan.Hooks = []component.Hook{{Name: "go-fmt"}}

	result, err := g.Generate(context.Background(), generatorTestRecipe(), plan, Options{OutDir: outDir, DryRun: true})

	s.Require().NoError(err)
	s.Require().Equal([]string{"README.md"}, result.FilesPlanned)
	s.Require().Empty(result.FilesWritten)
	s.Require().Equal([]string{"go-fmt"}, result.SkippedHooks)
	_, err = os.Stat(filepath.Join(outDir, "README.md"))
	s.Require().True(errors.Is(err, os.ErrNotExist))
	s.Require().Empty(readDirNames(outDir))
}

func (s *GeneratorTestSuite) TestDryRunIgnoresExistingOutputState() {
	outDir := s.T().TempDir()
	targetPath := filepath.Join(outDir, "README.md")
	s.Require().NoError(os.WriteFile(targetPath, []byte("old"), 0o644))
	g := NewGenerator(templatefs.FS)

	result, err := g.Generate(context.Background(), generatorTestRecipe(), readmePlan("README.md"), Options{OutDir: outDir, DryRun: true})

	s.Require().NoError(err)
	s.Require().Equal([]string{"README.md"}, result.FilesPlanned)

	content, err := os.ReadFile(targetPath)
	s.Require().NoError(err)
	s.Require().Equal("old", string(content))
}

func (s *GeneratorTestSuite) TestWritesNestedFilesWithRegularPermissions() {
	outDir := s.T().TempDir()
	g := NewGenerator(fstest.MapFS{
		"readme.tmpl": &fstest.MapFile{Data: []byte("project={{ .ProjectName }} module={{ .ModulePath }}")},
	})
	plan := &Plan{Files: []component.TemplateFile{{Source: "readme.tmpl", Target: "docs/README.md"}}}

	result, err := g.Generate(context.Background(), generatorTestRecipe(), plan, Options{OutDir: outDir})

	s.Require().NoError(err)
	s.Require().Equal([]string{"docs/README.md"}, result.FilesWritten)

	path := filepath.Join(outDir, "docs", "README.md")
	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Require().Equal("project=example module=example.com/example", string(content))

	info, err := os.Stat(path)
	s.Require().NoError(err)
	s.Require().Equal(regularFileMode, info.Mode().Perm())
}

func (s *GeneratorTestSuite) TestRejectsNonEmptyOutputDirectoryWithoutForce() {
	outDir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(outDir, "existing.txt"), []byte("keep"), 0o644))
	g := NewGenerator(templatefs.FS)

	_, err := g.Generate(context.Background(), generatorTestRecipe(), readmePlan("README.md"), Options{OutDir: outDir})

	s.Require().Error(err)
	var notEmptyErr *OutputDirectoryNotEmptyError
	s.Require().True(errors.As(err, &notEmptyErr))
	s.Require().Equal(outDir, notEmptyErr.Path)
}

func (s *GeneratorTestSuite) TestRejectsExistingTargetFileWithoutForce() {
	outDir := s.T().TempDir()
	targetPath := filepath.Join(outDir, "README.md")
	s.Require().NoError(os.WriteFile(targetPath, []byte("old"), 0o644))
	g := NewGenerator(templatefs.FS)

	_, err := g.Generate(context.Background(), generatorTestRecipe(), readmePlan("README.md"), Options{OutDir: outDir})

	s.Require().Error(err)
	var existsErr *FileExistsError
	s.Require().True(errors.As(err, &existsErr))
	s.Require().Equal(targetPath, existsErr.Path)
}

func (s *GeneratorTestSuite) TestForceOverwritesExistingTargetFile() {
	outDir := s.T().TempDir()
	targetPath := filepath.Join(outDir, "README.md")
	s.Require().NoError(os.WriteFile(targetPath, []byte("old"), 0o600))
	g := NewGenerator(templatefs.FS)

	result, err := g.Generate(context.Background(), generatorTestRecipe(), readmePlan("README.md"), Options{OutDir: outDir, Force: true})

	s.Require().NoError(err)
	s.Require().Equal([]string{"README.md"}, result.FilesWritten)

	content, err := os.ReadFile(targetPath)
	s.Require().NoError(err)
	s.Require().Contains(string(content), "# example")

	info, err := os.Stat(targetPath)
	s.Require().NoError(err)
	s.Require().Equal(regularFileMode, info.Mode().Perm())
}

func generatorTestRecipe() *recipe.Recipe {
	return &recipe.Recipe{
		Version: recipe.VersionV1,
		Project: recipe.ProjectConfig{
			Name:   "example",
			Module: "example.com/example",
			Type:   recipe.ProjectTypeWeb,
		},
	}
}

func readmePlan(target string) *Plan {
	return &Plan{
		Components: []component.Component{{ID: component.IDProjectWeb}},
		Files:      []component.TemplateFile{{Source: "project/README.md.tmpl", Target: target}},
	}
}

func readDirNames(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}
