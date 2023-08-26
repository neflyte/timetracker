package windows

import "fyne.io/fyne/v2"

// windowBase is the base interface of all app windows
type windowBase interface {
	fyne.Window
	Init() error // Init initializes the window
}

// resizeToMinimum resizes a window to a minimum size if it is not at least that size already
func resizeToMinimum(window fyne.Window, minimumWidth float32, minimumHeight float32) {
	windowSize := window.Canvas().Size()
	if windowSize.Width < minimumWidth {
		windowSize.Width = minimumWidth
	}
	if windowSize.Height < minimumHeight {
		windowSize.Height = minimumHeight
	}
	window.Resize(windowSize)
}
