package tui

import (
	"errors"

	"github.com/v0xpopuli/crego/internal/tui/screens"
)

var ErrCanceled = errors.New("tui canceled")

type (
	PushScreenMsg struct {
		Screen screens.Screen
	}

	PopScreenMsg struct{}

	ReplaceScreenMsg struct {
		Screen screens.Screen
	}

	SetErrorMsg struct {
		Err error
	}

	ClearErrorMsg struct{}

	CancelMsg struct{}
)

func PushScreen(screen screens.Screen) PushScreenMsg {
	return PushScreenMsg{Screen: screen}
}

func PopScreen() PopScreenMsg {
	return PopScreenMsg{}
}

func ReplaceScreen(screen screens.Screen) ReplaceScreenMsg {
	return ReplaceScreenMsg{Screen: screen}
}

func SetError(err error) SetErrorMsg {
	return SetErrorMsg{Err: err}
}

func ClearError() ClearErrorMsg {
	return ClearErrorMsg{}
}

func Cancel() CancelMsg {
	return CancelMsg{}
}
