package gui

import "fyne.io/fyne/v2"

const (
	// minimumWindowWidth is the minimum width of a window in pixels
	minimumWindowWidth = 600.0
	// minimumWindowHeight is the minimum height of a window in pixels
	minimumWindowHeight = 400.0
	// manageWindowEventChannelBufferSize is the size of an event channel
	manageWindowEventChannelBufferSize = 2

	// PrefKeyCloseWindow is the preferences key for the flag which causes the main window to close after creating a new task
	PrefKeyCloseWindow = "close-window"
)

var (
	// minimumWindowSize is the fyne.Size representation of the minimumWindowWidth and minimumWindowHeight
	minimumWindowSize = fyne.NewSize(minimumWindowWidth, minimumWindowHeight)
)
