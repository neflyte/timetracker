package windows

import "fyne.io/fyne/v2"

const (
	// minimumWindowWidth is the minimum width of a window in pixels
	minimumWindowWidth = 800.0
	// minimumWindowHeight is the minimum height of a window in pixels
	minimumWindowHeight = 600.0
	// prefKeyCloseWindowStopTask is the preferences key for the flag which causes the main window to close after creating a new task
	prefKeyCloseWindowStopTask = "close-window:stop-task"
	// dialogSizeOffset is the number of pixels to subtract from the parent window's size when setting a dialog's minimum size
	dialogSizeOffset = 50
)

var (
	// minimumWindowSize is the fyne.Size representation of the minimumWindowWidth and minimumWindowHeight
	minimumWindowSize = fyne.NewSize(minimumWindowWidth, minimumWindowHeight)
)
