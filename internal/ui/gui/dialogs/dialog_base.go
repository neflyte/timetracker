package dialogs

import "fyne.io/fyne/v2/dialog"

// dialogBase is the base interface of all app dialogs
type dialogBase interface {
	dialog.Dialog
	Init() error // Init initializes the dialog
}
