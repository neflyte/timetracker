package gui

import "fyne.io/fyne/v2"

const (
	// MinimumWindowWidth is the minimum width of a window in pixels
	MinimumWindowWidth = 600.0
	// MinimumWindowHeight is the minimum height of a window in pixels
	MinimumWindowHeight = 400.0
	// ManageWindowEventChannelBufferSize is the size of an event channel
	ManageWindowEventChannelBufferSize = 2
)

var (
	// MinimumWindowSize is the fyne.Size representation of the MinimumWindowWidth and MinimumWindowHeight
	MinimumWindowSize = fyne.NewSize(MinimumWindowWidth, MinimumWindowHeight)
)
