package windows

import "fyne.io/fyne/v2"

const (
	// minimumWindowWidth is the minimum width of a window in pixels
	minimumWindowWidth = 425.0
	// minimumWindowHeight is the minimum height of a window in pixels
	minimumWindowHeight = 400.0
	// dialogSizeOffset is the number of pixels to subtract from the parent window's size when setting a dialog's minimum size
	dialogSizeOffset = 50
)

var (
	// minimumWindowSize is the fyne.Size representation of the minimumWindowWidth and minimumWindowHeight
	minimumWindowSize = fyne.NewSize(minimumWindowWidth, minimumWindowHeight)
)
