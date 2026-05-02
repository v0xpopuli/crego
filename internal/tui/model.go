package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/v0xpopuli/crego/internal/tui/components"
	"github.com/v0xpopuli/crego/internal/tui/screens"
)

type Model struct {
	screens  []screens.Screen
	styles   Styles
	width    int
	height   int
	showHelp bool
	err      error
	canceled bool
}

func NewModel(initial screens.Screen, styles Styles) Model {
	return Model{
		screens: []screens.Screen{initial},
		styles:  styles,
	}
}

func (m Model) Init() tea.Cmd {
	if current := m.CurrentScreen(); current != nil {
		return current.Init()
	}
	return nil
}

func (m Model) CurrentScreen() screens.Screen {
	if len(m.screens) == 0 {
		return nil
	}
	return m.screens[len(m.screens)-1]
}

func (m Model) Canceled() bool {
	return m.canceled
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.canceled = true
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc":
			if managed, ok := m.CurrentScreen().(shellManaged); ok && managed.UsesShellLayout() {
				break
			}
			if len(m.screens) > 1 {
				m.screens = m.screens[:len(m.screens)-1]
				return m, nil
			}
			m.canceled = true
			return m, tea.Quit
		}
	case PushScreenMsg:
		if msg.Screen != nil {
			m.screens = append(m.screens, msg.Screen)
			return m, msg.Screen.Init()
		}
		return m, nil
	case PopScreenMsg:
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
			return m, nil
		}
		m.canceled = true
		return m, tea.Quit
	case ReplaceScreenMsg:
		if msg.Screen != nil {
			if len(m.screens) == 0 {
				m.screens = []screens.Screen{msg.Screen}
			} else {
				m.screens[len(m.screens)-1] = msg.Screen
			}
			return m, msg.Screen.Init()
		}
		return m, nil
	case SetErrorMsg:
		m.err = msg.Err
		return m, nil
	case ClearErrorMsg:
		m.err = nil
		return m, nil
	case CancelMsg:
		m.canceled = true
		return m, tea.Quit
	}

	current := m.CurrentScreen()
	if current == nil {
		return m, tea.Quit
	}

	next, cmd := current.Update(msg)
	if next != nil {
		m.screens[len(m.screens)-1] = next
	}
	return m, cmd
}

func (m Model) View() string {
	current := m.CurrentScreen()
	if current == nil {
		return ""
	}

	var parts []string
	if errorPanel := components.ErrorPanel(m.styles.Components(), m.err); errorPanel != "" {
		parts = append(parts, errorPanel)
	}
	parts = append(parts, current.View())
	if managed, ok := current.(shellManaged); !ok || !managed.UsesShellLayout() {
		parts = append(parts, components.Footer(m.styles.Components(), m.showHelp))
	}

	if managed, ok := current.(shellManaged); ok && managed.UsesShellLayout() {
		return strings.Join(parts, "\n\n")
	}
	return m.styles.App.Render(strings.Join(parts, "\n\n"))
}
