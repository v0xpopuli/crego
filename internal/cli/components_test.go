package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func (s *CliTestSuite) TestComponentsListCommand() {
	s.Run("prints grouped output in stable category order", func() {
		out, _, err := s.executeCLI("components", "list")

		s.Require().NoError(err)
		s.Require().Equal(componentCategoryOrder, topLevelHeaders(out))
		s.Require().Contains(out, "  gin - HTTP server built with Gin.")
		s.Require().Contains(out, "  framework:")
		s.Require().Contains(out, "    gorm - Database access through GORM.")
		s.Require().NotContains(out, "server.gin - HTTP server built with Gin.")
		s.Require().NotContains(out, "database.framework.gorm - Database access through GORM.")
		s.Require().NotContains(out, "status: planned, not yet generated")
		s.Require().NotContains(out, "db.")
		s.Require().NotContains(out, "database.package.")
	})

	s.Run("filters server category", func() {
		out, _, err := s.executeCLI("components", "list", "--category", "server")

		s.Require().NoError(err)
		s.Require().Equal([]string{"server"}, topLevelHeaders(out))
		s.Require().Contains(out, "  nethttp -")
		s.Require().Contains(out, "  chi -")
		s.Require().Contains(out, "  gin -")
		s.Require().Contains(out, "  echo -")
		s.Require().Contains(out, "  fiber -")
		s.Require().NotContains(out, "server.gin")
		s.Require().NotContains(out, "database.postgres")
	})

	s.Run("filters database category including frameworks", func() {
		out, _, err := s.executeCLI("components", "list", "--category", "database")

		s.Require().NoError(err)
		s.Require().Equal([]string{"database"}, topLevelHeaders(out))
		s.Require().Contains(out, "  none -")
		s.Require().Contains(out, "  postgres -")
		s.Require().Contains(out, "  mysql -")
		s.Require().Contains(out, "  sqlite -")
		s.Require().Contains(out, "  framework:")
		s.Require().Contains(out, "    pgx -")
		s.Require().Contains(out, "    sql -")
		s.Require().Contains(out, "    gorm -")
		s.Require().NotContains(out, "database.postgres")
		s.Require().NotContains(out, "database.framework.gorm")
	})

	s.Run("prints valid json", func() {
		out, _, err := s.executeCLI("components", "list", "--json")

		s.Require().NoError(err)
		var result componentsListOutput
		s.Require().NoError(json.Unmarshal([]byte(out), &result))
		s.Require().NotEmpty(result.Categories)
		s.Require().Equal("project", result.Categories[0].Category)
		s.Require().NotNil(result.Categories[0].Components)
	})

	s.Run("rejects unknown category", func() {
		_, _, err := s.executeCLI("components", "list", "--category", "routing")

		s.Require().Error(err)
		s.Require().Equal(1, ExitCode(err))
		s.Require().Contains(err.Error(), `unknown component category "routing"`)
		s.Require().Contains(err.Error(), "server")
		s.Require().Contains(err.Error(), "database")
	})
}

func (s *CliTestSuite) TestComponentListDisplayParts() {
	s.Run("strips category from flat ids", func() {
		group, id := componentListDisplayParts("server", "server.gin")

		s.Require().Empty(group)
		s.Require().Equal("gin", id)
	})

	s.Run("nests any dotted suffix after category", func() {
		group, id := componentListDisplayParts("observability", "observability.endpoint.health")

		s.Require().Equal("endpoint", group)
		s.Require().Equal("health", id)
	})

	s.Run("leaves foreign ids intact", func() {
		group, id := componentListDisplayParts("server", "database.postgres")

		s.Require().Empty(group)
		s.Require().Equal("database.postgres", id)
	})
}

func (s *CliTestSuite) TestComponentsShowCommand() {
	for _, id := range []string{
		"server.chi",
		"server.gin",
		"database.postgres",
		"database.framework.gorm",
	} {
		s.Run(id+" prints useful metadata", func() {
			out, _, err := s.executeCLI("components", "show", id)

			s.Require().NoError(err)
			s.Require().Contains(out, "id: "+id)
			s.Require().Contains(out, "category:")
			s.Require().Contains(out, "description:")
			s.Require().Contains(out, "requires:")
			s.Require().Contains(out, "conflicts:")
			s.Require().Contains(out, "files:")
			s.Require().Contains(out, "go_modules:")
			s.Require().Contains(out, "hooks:")
			s.Require().Contains(out, "support_status: planned-not-yet-generated")
		})
	}

	s.Run("prints valid json", func() {
		out, _, err := s.executeCLI("components", "show", "server.gin", "--json")

		s.Require().NoError(err)
		var result componentDetailOutput
		s.Require().NoError(json.Unmarshal([]byte(out), &result))
		s.Require().Equal("server.gin", result.ID)
		s.Require().Equal("server", result.Category)
		s.Require().NotNil(result.Requires)
		s.Require().NotNil(result.Conflicts)
		s.Require().NotNil(result.Files)
		s.Require().NotNil(result.GoModules)
		s.Require().NotNil(result.Hooks)
	})

	s.Run("rejects unknown component", func() {
		_, _, err := s.executeCLI("components", "show", "server.martini")

		s.Require().Error(err)
		s.Require().Equal(1, ExitCode(err))
		s.Require().Contains(err.Error(), "unknown component server.martini")
	})
}

func (s *CliTestSuite) TestExplainCommand() {
	s.Run("prints generation plan without writing files", func() {
		path := s.writeStarterRecipe()
		before := tempDirEntries(filepath.Dir(path))

		out, _, err := s.executeCLI("explain", "--recipe", path)

		s.Require().NoError(err)
		s.Require().Equal(before, tempDirEntries(filepath.Dir(path)))
		s.Require().Contains(out, "recipe: "+path)
		s.Require().Contains(out, "selected components:")
		s.Require().Contains(out, "project.web")
		s.Require().Contains(out, "database.postgres")
		s.Require().Contains(out, "generated files:")
		s.Require().Contains(out, "go modules:")
		s.Require().Contains(out, "hooks:")
		s.Require().NotContains(out, "status: planned, not yet generated")
	})

	s.Run("prints valid json", func() {
		path := s.writeStarterRecipe()

		out, _, err := s.executeCLI("explain", "--recipe", path, "--json")

		s.Require().NoError(err)
		var result explainOutput
		s.Require().NoError(json.Unmarshal([]byte(out), &result))
		s.Require().Equal(path, result.Recipe)
		s.Require().NotEmpty(result.Components)
		s.Require().NotNil(result.Files)
		s.Require().NotNil(result.GoModules)
		s.Require().NotNil(result.Hooks)
	})
}

func (s *CliTestSuite) TestComponentsAndExplainHelpExamples() {
	cmd := NewRootCommandWithWriters(VersionInfo{}, &bytes.Buffer{}, &bytes.Buffer{})
	for _, args := range [][]string{
		{"components"},
		{"components", "list"},
		{"components", "show"},
		{"explain"},
	} {
		found, _, err := cmd.Find(args)

		s.Require().NoError(err)
		s.Require().NotEmpty(strings.TrimSpace(found.Example))
		s.Require().Contains(found.Example, "crego "+strings.Join(args, " "))
	}
}

func (s *CliTestSuite) TestComponentsCommandRequiresSubcommand() {
	out, _, err := s.executeCLI("components")

	s.Require().Error(err)
	s.Require().EqualError(err, "components requires a subcommand: list or show")
	s.Require().Contains(out, "Explore available project components")
	s.Require().Contains(out, "list")
	s.Require().Contains(out, "show")
}

func topLevelHeaders(out string) []string {
	var headers []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, " ") || !strings.HasSuffix(line, ":") {
			continue
		}
		headers = append(headers, strings.TrimSuffix(line, ":"))
	}
	return headers
}

func tempDirEntries(path string) []string {
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
