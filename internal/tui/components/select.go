package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type (
	Select struct {
		Title   string
		Options []SelectOption
		Cursor  int
	}

	SelectOption struct {
		Label       string
		Value       string
		Description string
	}
)

func NewSelect(title string, options []SelectOption) Select {
	return Select{Title: title, Options: options}
}

func NewHuhSelect(title string, value *string, options []SelectOption) *huh.Select[string] {
	items := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		items = append(items, huh.NewOption(option.Label, option.Value))
	}
	return huh.NewSelect[string]().Title(title).Value(value).Options(items...)
}

func (s Select) Selected() SelectOption {
	if len(s.Options) == 0 {
		return SelectOption{}
	}
	if s.Cursor < 0 {
		s.Cursor = 0
	}
	if s.Cursor >= len(s.Options) {
		s.Cursor = len(s.Options) - 1
	}
	return s.Options[s.Cursor]
}

func (s Select) Update(msg tea.Msg) Select {
	key, ok := msg.(tea.KeyMsg)
	if !ok || len(s.Options) == 0 {
		return s
	}

	switch key.String() {
	case "up", "k":
		if s.Cursor > 0 {
			s.Cursor--
		}
	case "down", "j":
		if s.Cursor < len(s.Options)-1 {
			s.Cursor++
		}
	}
	return s
}

func (s Select) View(styles Styles) string {
	lines := make([]string, 0, len(s.Options)+1)
	if s.Title != "" {
		lines = append(lines, styles.Title.Render(s.Title))
	}

	for i, option := range s.Options {
		prefix := "  "
		labelStyle := styles.Option
		if i == s.Cursor {
			prefix = "› "
			labelStyle = styles.Selected
		}
		label := prefix + option.Label
		if option.Description != "" {
			label += "  " + styles.Description.Render(option.Description)
		}
		lines = append(lines, labelStyle.Render(label))
	}

	return strings.Join(lines, "\n")
}
