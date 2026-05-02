package tui

import (
	"errors"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/v0xpopuli/crego/internal/tui/components"
	"github.com/v0xpopuli/crego/internal/tui/screens"
)

type (
	App struct {
		model Model
		in    io.Reader
		out   io.Writer
	}

	AppOptions struct {
		In      io.Reader
		Out     io.Writer
		NoColor bool
	}
)

func NewApp(initial screens.Screen, opts AppOptions) *App {
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	styles := NewStyles(opts.Out, opts.NoColor)
	return &App{
		model: NewModel(initial, styles),
		in:    opts.In,
		out:   opts.Out,
	}
}

func NewDemoApp(opts AppOptions) *App {
	if opts.Out == nil {
		opts.Out = io.Discard
	}
	styles := NewStyles(opts.Out, opts.NoColor)
	return &App{
		model: NewModel(NewDemoScreen(styles), styles),
		in:    opts.In,
		out:   opts.Out,
	}
}

func (a *App) Run() error {
	options := []tea.ProgramOption{
		tea.WithOutput(a.out),
		tea.WithAltScreen(),
	}
	if a.in != nil {
		options = append(options, tea.WithInput(a.in))
	}

	model, err := tea.NewProgram(a.model, options...).Run()
	if err != nil {
		return err
	}

	if finalModel, ok := model.(Model); ok && finalModel.Canceled() {
		return ErrCanceled
	}
	return nil
}

type DemoScreen struct {
	styles      Styles
	selectInput components.Select
}

func NewDemoScreen(styles Styles) DemoScreen {
	return DemoScreen{
		styles: styles,
		selectInput: components.NewSelect("", []components.SelectOption{
			{Label: "Continue", Value: "continue"},
			{Label: "Cancel", Value: "cancel"},
		}),
	}
}

func (s DemoScreen) Init() tea.Cmd {
	return nil
}

func (s DemoScreen) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		if s.selectInput.Selected().Value == "cancel" {
			return s, func() tea.Msg { return Cancel() }
		}
		return s, func() tea.Msg {
			return SetError(errors.New("configure wizard is not implemented yet"))
		}
	}

	s.selectInput = s.selectInput.Update(msg)
	return s, nil
}

func (s DemoScreen) View() string {
	return stringsJoin(
		s.styles.Title.Render("crego"),
		s.styles.Description.Render("Create Go services from composable recipes."),
		s.selectInput.View(s.styles.Components()),
	)
}

func stringsJoin(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, "\n")
}
