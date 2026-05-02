package tui

import (
	"bytes"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
	"github.com/v0xpopuli/crego/internal/tui/screens"
)

type TuiTestSuite struct {
	suite.Suite
	styles Styles
}

func TestTuiTestSuite(t *testing.T) {
	suite.Run(t, new(TuiTestSuite))
}

func (s *TuiTestSuite) SetupTest() {
	s.styles = NewStyles(&bytes.Buffer{}, true)
}

func (s *TuiTestSuite) TestRootModel() {
	s.Run("renders screen with footer", func() {
		model := NewModel(fakeScreen{name: "home"}, s.styles)

		view := model.View()

		s.Require().Contains(view, "screen: home")
		s.Require().Contains(view, "enter confirm")
	})

	s.Run("toggles help", func() {
		model := NewModel(fakeScreen{name: "home"}, s.styles)

		next, _ := model.Update(keyRunes('?'))

		rendered := next.(Model).View()
		s.Require().Contains(rendered, "space toggle multi-select")
	})

	s.Run("tracks terminal resize", func() {
		model := NewModel(fakeScreen{name: "home"}, s.styles)

		next, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

		updated := next.(Model)
		s.Require().Equal(120, updated.width)
		s.Require().Equal(40, updated.height)
	})

	s.Run("quits cleanly on q", func() {
		model := NewModel(fakeScreen{name: "home"}, s.styles)

		next, cmd := model.Update(keyRunes('q'))

		updated := next.(Model)
		s.Require().True(updated.Canceled())
		s.Require().NotNil(cmd)
	})

	s.Run("esc pops a screen before canceling", func() {
		model := NewModel(fakeScreen{name: "home"}, s.styles)
		next, _ := model.Update(PushScreen(fakeScreen{name: "next"}))

		next, cmd := next.(Model).Update(keyType(tea.KeyEsc))

		updated := next.(Model)
		s.Require().False(updated.Canceled())
		s.Require().Nil(cmd)
		s.Require().Len(updated.screens, 1)
		s.Require().Contains(updated.View(), "screen: home")
	})

	s.Run("error panel renders supplied error", func() {
		model := NewModel(fakeScreen{name: "home"}, s.styles)

		next, _ := model.Update(SetError(errors.New("recipe is invalid")))

		s.Require().Contains(next.(Model).View(), "Error: recipe is invalid")
	})
}

func (s *TuiTestSuite) TestDemoScreen() {
	s.Run("renders launch placeholder", func() {
		screen := NewDemoScreen(s.styles)

		view := screen.View()

		s.Require().Contains(view, "crego")
		s.Require().Contains(view, "Create Go services from composable recipes.")
		s.Require().Contains(view, "Continue")
		s.Require().Contains(view, "Cancel")
	})

	s.Run("cancel option emits cancel message", func() {
		screen := NewDemoScreen(s.styles)
		next, _ := screen.Update(keyType(tea.KeyDown))

		_, cmd := next.Update(keyType(tea.KeyEnter))

		s.Require().IsType(CancelMsg{}, cmd())
	})
}

func (s *TuiTestSuite) TestNoColorStyles() {
	var out bytes.Buffer
	styles := NewStyles(&out, true)
	model := NewModel(fakeScreen{name: "plain"}, styles)

	rendered := model.View()

	s.Require().NotContains(rendered, "\x1b[")
}

type fakeScreen struct {
	name string
}

func (f fakeScreen) Init() tea.Cmd {
	return nil
}

func (f fakeScreen) Update(_ tea.Msg) (screens.Screen, tea.Cmd) {
	return f, nil
}

func (f fakeScreen) View() string {
	return "screen: " + f.name
}

func keyRunes(r rune) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}})
}

func keyType(key tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: key})
}
