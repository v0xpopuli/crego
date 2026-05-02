package tui

import (
	"io"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/v0xpopuli/crego/internal/tui/components"
)

type Styles struct {
	Renderer *lipgloss.Renderer
	HuhTheme *huh.Theme

	App         lipgloss.Style
	Title       lipgloss.Style
	Description lipgloss.Style
	Option      lipgloss.Style
	Selected    lipgloss.Style
	Footer      lipgloss.Style
	Help        lipgloss.Style
	Error       lipgloss.Style
	Preview     lipgloss.Style
}

func (s Styles) Components() components.Styles {
	return components.Styles{
		Title:    s.Title,
		Option:   s.Option,
		Selected: s.Selected,
		Footer:   s.Footer,
		Error:    s.Error,
		Preview:  s.Preview,
	}
}

func NewStyles(out io.Writer, noColor bool) Styles {
	if out == nil {
		out = io.Discard
	}
	renderer := lipgloss.NewRenderer(out)
	if noColor {
		renderer.SetColorProfile(termenv.Ascii)
	}

	theme := huh.ThemeCharm()
	if noColor {
		theme = huh.ThemeBase()
	}

	style := renderer.NewStyle
	muted := lipgloss.AdaptiveColor{Light: "241", Dark: "246"}
	accent := lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#8B87FF"}
	errorColor := lipgloss.AdaptiveColor{Light: "#C93434", Dark: "#FF6B6B"}

	return Styles{
		Renderer: renderer,
		HuhTheme: theme,
		App: style().
			Padding(1, 2),
		Title: style().
			Bold(true).
			Foreground(accent).
			MarginBottom(1),
		Description: style().
			Foreground(muted).
			MarginBottom(2),
		Option: style().
			PaddingLeft(2),
		Selected: style().
			Bold(true).
			Foreground(accent),
		Footer: style().
			Foreground(muted).
			MarginTop(2),
		Help: style().
			Foreground(muted).
			MarginTop(1),
		Error: style().
			Foreground(errorColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(0, 1).
			MarginBottom(1),
		Preview: style().
			Border(lipgloss.NormalBorder()).
			BorderForeground(muted).
			Padding(1, 2),
	}
}
