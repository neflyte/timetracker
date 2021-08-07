package windows

import "fyne.io/fyne/v2"

// windowBase is the base interface of all app windows
type windowBase interface {
	fyne.Window
	Init() error // Init initializes the window
}
