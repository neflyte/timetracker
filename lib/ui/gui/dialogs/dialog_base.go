package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// dialogBase is the base interface of all app dialogs
type dialogBase interface {
	dialog.Dialog
	Init() error // Init initializes the dialog
}

func ResizeDialogToWindowWithPadding(d dialog.Dialog, w fyne.Window, padding float32) {
	if d == nil || w == nil {
		return
	}
	dsize := d.MinSize()
	winsize := w.Canvas().Size()
	if dsize.Width < winsize.Width-padding {
		dsize.Width = winsize.Width - padding
	}
	if dsize.Height < winsize.Height-padding {
		dsize.Height = winsize.Height - padding
	}
	d.Resize(dsize)
}
