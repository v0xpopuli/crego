package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type (
	HelpMode int

	LayoutProps struct {
		Title      string
		Subtitle   string
		Step       int
		TotalSteps int
		Version    string
		Sidebar    string
		Body       string
		Preview    string
		StackLine  string
		Help       string
		Error      string
		Width      int
		Height     int
	}

	shellManaged interface {
		UsesShellLayout() bool
	}
)

const (
	HelpSelect HelpMode = iota
	HelpInput
	HelpMulti
	HelpPreview
)

func HelpLine(mode HelpMode) string {
	switch mode {
	case HelpInput:
		return "tab next · enter continue · esc back · q quit"
	case HelpMulti:
		return "space toggle · enter continue · esc back · q quit"
	case HelpPreview:
		return "↑↓ move · enter confirm · esc back · q quit"
	default:
		return "↑↓ move · enter select · esc back · q quit"
	}
}

func RenderShell(styles Styles, props LayoutProps) string {
	width := props.Width
	if width <= 0 {
		width = 100
	}
	width = max(width-1, 48)
	height := props.Height
	if height <= 0 {
		height = 30
	}
	height = max(height, 18)

	innerWidth := width - 2
	header := padLine(renderShellHeader(styles, props, innerWidth), innerWidth)
	footerLines := splitFixed(renderShellFooter(styles, props, innerWidth), innerWidth)
	if len(footerLines) == 0 {
		footerLines = []string{""}
	}
	bodyHeight := max(height-5-len(footerLines), 6)
	bodyLines := renderShellBody(styles, props, innerWidth, bodyHeight)

	rows := make([]string, 0, height)
	rows = append(rows, topBorder(styles, innerWidth))
	rows = append(rows, borderedLine(styles, header, innerWidth))
	rows = append(rows, separator(styles, innerWidth))
	for _, line := range bodyLines {
		rows = append(rows, borderedLine(styles, line, innerWidth))
	}
	rows = append(rows, separator(styles, innerWidth))
	for _, line := range footerLines {
		rows = append(rows, borderedLine(styles, line, innerWidth))
	}
	rows = append(rows, bottomBorder(styles, innerWidth))
	return strings.Join(rows, "\n")
}

func renderShellHeader(styles Styles, props LayoutProps, width int) string {
	left := styles.Title.Render(props.Title)
	if props.Step > 0 && props.TotalSteps > 0 {
		left = fmt.Sprintf("%s  %s", left, styles.Description.Render(fmt.Sprintf("Step %d/%d: %s", props.Step, props.TotalSteps, props.Subtitle)))
	} else if props.Subtitle != "" {
		left = fmt.Sprintf("%s  %s", left, styles.Description.Render(props.Subtitle))
	}
	if props.Version == "" {
		return left
	}
	space := width - lipgloss.Width(left) - lipgloss.Width(props.Version)
	if space < 1 {
		space = 1
	}
	return left + strings.Repeat(" ", space) + styles.Description.Render(props.Version)
}

func renderShellBody(styles Styles, props LayoutProps, width int, height int) []string {
	body := props.Body
	if props.Error != "" {
		body = strings.Join(nonEmpty(props.Error, body), "\n\n")
	}
	if props.Sidebar == "" {
		if width > 126 && props.Preview != "" {
			bodyWidth := width - 45
			return joinColumns(
				renderColumn(body, bodyWidth, height),
				verticalRule(styles, height),
				renderColumn(props.Preview, 42, height),
			)
		}
		return renderColumn(body, width, height)
	}
	if width < 74 {
		compact := strings.Join(nonEmpty(props.Sidebar, body, props.Preview), "\n\n")
		return renderColumn(compact, width, height)
	}

	sidebarWidth := clamp(width/6, 20, 30)
	previewWidth := clamp(width/3, 42, 72)
	if props.Preview == "" {
		props.Preview = styles.Description.Render("Files tree\n\nWill update as the stack resolves.")
	}
	bodyWidth := max(width-sidebarWidth-previewWidth-2, 24)

	return joinColumns(
		renderColumn(props.Sidebar, sidebarWidth, height),
		verticalRule(styles, height),
		renderColumn(body, bodyWidth, height),
		verticalRule(styles, height),
		renderColumn(props.Preview, previewWidth, height),
	)
}

func renderShellFooter(styles Styles, props LayoutProps, width int) string {
	lines := nonEmpty(props.StackLine, styles.Footer.Render(props.Help))
	if len(lines) == 0 {
		return ""
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

func nonEmpty(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			out = append(out, part)
		}
	}
	return out
}

func clamp(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func max(value int, min int) int {
	if value < min {
		return min
	}
	return value
}

func verticalRule(styles Styles, height int) []string {
	lines := make([]string, height)
	for index := range lines {
		lines[index] = styles.Description.Render("│")
	}
	return lines
}

func renderColumn(text string, width int, height int) []string {
	rendered := lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		Height(height).
		PaddingTop(1).
		PaddingLeft(1).
		Render(text)
	return fixedLines(rendered, width, height)
}

func joinColumns(columns ...[]string) []string {
	height := 0
	for _, column := range columns {
		if len(column) > height {
			height = len(column)
		}
	}
	lines := make([]string, height)
	for row := 0; row < height; row++ {
		var builder strings.Builder
		for _, column := range columns {
			if row < len(column) {
				builder.WriteString(column[row])
			}
		}
		lines[row] = builder.String()
	}
	return lines
}

func splitFixed(text string, width int) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return fixedLines(text, width, lipgloss.Height(text))
}

func fixedLines(text string, width int, height int) []string {
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, height)
	for _, line := range raw {
		if len(lines) >= height {
			break
		}
		lines = append(lines, padLine(line, width))
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines
}

func padLine(line string, width int) string {
	lineWidth := lipgloss.Width(line)
	if lineWidth >= width {
		return line
	}
	return line + strings.Repeat(" ", width-lineWidth)
}

func topBorder(styles Styles, width int) string {
	return styles.Description.Render("┌" + strings.Repeat("─", width) + "┐")
}

func separator(styles Styles, width int) string {
	return styles.Description.Render("├" + strings.Repeat("─", width) + "┤")
}

func bottomBorder(styles Styles, width int) string {
	return styles.Description.Render("└" + strings.Repeat("─", width) + "┘")
}

func borderedLine(styles Styles, line string, width int) string {
	return styles.Description.Render("│") + padLine(line, width) + styles.Description.Render("│")
}
