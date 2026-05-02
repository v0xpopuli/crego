package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CliTestSuite struct {
	suite.Suite
}

func TestCliTestSuite(t *testing.T) {
	suite.Run(t, new(CliTestSuite))
}

func (s *CliTestSuite) TestRootCommand() {
	s.Run("contains planned commands", func() {
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd := NewRootCommandWithWriters(VersionInfo{}, &out, &errOut)

		for _, name := range []string{
			"new",
			"configure",
			"generate",
			"recipe",
			"components",
			"explain",
			"version",
		} {
			child, _, err := cmd.Find([]string{name})

			s.Require().NoError(err)
			s.Require().Equal(name, child.Name())
		}
	})

	s.Run("contains global flags", func() {
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd := NewRootCommandWithWriters(VersionInfo{}, &out, &errOut)

		for _, flag := range []string{"no-color", "verbose", "debug", "config"} {
			s.Require().NotNil(cmd.PersistentFlags().Lookup(flag))
		}
	})
}

func (s *CliTestSuite) TestVersionCommand() {
	s.Run("prints build metadata", func() {
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd := NewRootCommandWithWriters(VersionInfo{
			Version: "1.2.3",
			Commit:  "abc123",
			Built:   "2026-04-30T12:00:00Z",
		}, &out, &errOut)
		cmd.SetArgs([]string{"version"})

		err := cmd.Execute()

		s.Require().NoError(err)
		s.Require().Contains(out.String(), "version: 1.2.3")
		s.Require().Contains(out.String(), "commit: abc123")
		s.Require().Contains(out.String(), "built: 2026-04-30T12:00:00Z")
	})
}

func (s *CliTestSuite) TestPlaceholderCommands() {
	s.Run("non-interactive new explains missing module", func() {
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd := NewRootCommandWithWriters(VersionInfo{}, &out, &errOut)
		cmd.SetArgs([]string{"new", "--non-interactive"})

		err := cmd.Execute()

		s.Require().EqualError(err, "module path is required for non-interactive new")
	})
}
