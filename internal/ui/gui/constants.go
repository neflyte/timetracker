package gui

import "fyne.io/fyne/v2"

const (
	// minimumWindowWidth is the minimum width of a window in pixels
	minimumWindowWidth = 600.0
	// minimumWindowHeight is the minimum height of a window in pixels
	minimumWindowHeight = 400.0
	// manageWindowEventChannelBufferSize is the size of an event channel
	manageWindowEventChannelBufferSize = 2
)

var (
	// minimumWindowSize is the fyne.Size representation of the minimumWindowWidth and minimumWindowHeight
	minimumWindowSize = fyne.NewSize(minimumWindowWidth, minimumWindowHeight)
)
