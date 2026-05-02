package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type MultiSelect struct {
	Title    string
	Options  []SelectOption
	Cursor   int
	Selected map[string]bool
}

func NewMultiSelect(title string, options []SelectOption) MultiSelect {
	return MultiSelect{
		Title:    title,
		Options:  options,
		Selected: make(map[string]bool),
	}
}

func NewHuhMultiSelect(title string, values *[]string, options []SelectOption) *huh.MultiSelect[string] {
	items := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		items = append(items, huh.NewOption(option.Label, option.Value))
	}
	return huh.NewMultiSelect[string]().Title(title).Value(values).Options(items...)
}

func (m MultiSelect) Values() []string {
	values := make([]string, 0, len(m.Selected))
	for _, option := range m.Options {
		if m.Selected[option.Value] {
			values = append(values, option.Value)
		}
	}
	return values
}

func (m MultiSelect) Update(msg tea.Msg) MultiSelect {
	if m.Selected == nil {
		m.Selected = make(map[string]bool)
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok || len(m.Options) == 0 {
		return m
	}

	switch key.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Options)-1 {
			m.Cursor++
		}
	case " ":
		value := m.Options[m.Cursor].Value
		m.Selected[value] = !m.Selected[value]
	}
	return m
}

func (m MultiSelect) View(styles Styles) string {
	lines := make([]string, 0, len(m.Options)+1)
	if m.Title != "" {
		lines = append(lines, styles.Title.Render(m.Title))
	}

	for i, option := range m.Options {
		cursor := "  "
		if i == m.Cursor {
			cursor = "› "
		}

		check := "[ ] "
		if m.Selected != nil && m.Selected[option.Value] {
			check = "[x] "
		}

		labelStyle := styles.Option
		if i == m.Cursor {
			labelStyle = styles.Selected
		}
		label := cursor + check + option.Label
		if option.Description != "" {
			label += "  " + styles.Description.Render(option.Description)
		}
		lines = append(lines, labelStyle.Render(label))
	}

	return strings.Join(lines, "\n")
}
