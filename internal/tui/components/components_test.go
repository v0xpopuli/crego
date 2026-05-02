package components_test

import (
	"bytes"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/tui"
	"github.com/v0xpopuli/crego/internal/tui/components"
)

type ComponentsTestSuite struct {
	suite.Suite
	styles components.Styles
}

func TestComponentsTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentsTestSuite))
}

func (s *ComponentsTestSuite) SetupTest() {
	s.styles = tui.NewStyles(&bytes.Buffer{}, true).Components()
}

func (s *ComponentsTestSuite) TestSelect() {
	selectInput := components.NewSelect("Action", []components.SelectOption{
		{Label: "First", Value: "first"},
		{Label: "Second", Value: "second"},
	})

	selectInput = selectInput.Update(keyType(tea.KeyDown))

	s.Require().Equal("second", selectInput.Selected().Value)
	s.Require().Contains(selectInput.View(s.styles), "Second")
}

func (s *ComponentsTestSuite) TestMultiSelect() {
	input := components.NewMultiSelect("Features", []components.SelectOption{
		{Label: "HTTP", Value: "http"},
		{Label: "SQL", Value: "sql"},
	})

	input = input.Update(keyType(tea.KeySpace))
	input = input.Update(keyType(tea.KeyDown))
	input = input.Update(keyType(tea.KeySpace))

	s.Require().Equal([]string{"http", "sql"}, input.Values())
	s.Require().Contains(input.View(s.styles), "[x] HTTP")
}

func (s *ComponentsTestSuite) TestErrorAndPreviewPanels() {
	s.Require().Empty(components.ErrorPanel(s.styles, nil))
	s.Require().Contains(components.ErrorPanel(s.styles, errors.New("failed")), "Error: failed")
	s.Require().Contains(components.Preview(s.styles, "plan"), "plan")
}

func keyType(key tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: key})
}
