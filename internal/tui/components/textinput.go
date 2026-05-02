package components

import "github.com/charmbracelet/huh"

func NewTextInput(title string, value *string) *huh.Input {
	return huh.NewInput().Title(title).Value(value)
}
