package components

import (
	"strings"
)

func Footer(styles Styles, showHelp bool) string {
	keys := "enter confirm  esc back  q cancel  ? help"
	if showHelp {
		keys = strings.Join([]string{
			"enter confirm/next",
			"esc back, or cancel on the first screen",
			"up/down move selection",
			"space toggle multi-select",
			"q or ctrl+c cancel",
			"? hide help",
		}, "\n")
	}
	return styles.Footer.Render(keys)
}
