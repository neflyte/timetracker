package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type ErrorDialogWindow interface {
	Window() *fyne.Window
	Init() error
	Dialog() *dialog.Dialog
	SetDialog(dlg *dialog.Dialog)
	Show()
	Hide()
}

type errorDialogWindow struct {
	w           fyne.Window
	d           dialog.Dialog
	cb          func()
	initialSize *fyne.Size
	initialized bool
}

func NewErrorDialogWindow(app fyne.App, title string, err error, size *fyne.Size, callback func()) ErrorDialogWindow {
	edw := new(errorDialogWindow)
	edw.w = app.NewWindow(title)
	edw.w.SetFixedSize(true)
	edw.d = dialog.NewError(err, edw.w)
	edw.cb = callback
	edw.initialized = false
	edw.initialSize = &MinimumWindowSize
	if size != nil {
		edw.initialSize = size
	}
	err = edw.Init()
	if err != nil {
		return nil
	}
	return edw
}

func (e *errorDialogWindow) Window() *fyne.Window {
	return &e.w
}

func (e *errorDialogWindow) Init() error {
	if !e.initialized {
		e.d.SetOnClosed(e.Callback)
		if e.initialSize != nil {
			e.w.Resize(*e.initialSize)
			e.d.Resize(*e.initialSize)
		}
		e.initialized = true
	}
	return nil
}

func (e *errorDialogWindow) Dialog() *dialog.Dialog {
	return &e.d
}

func (e *errorDialogWindow) SetDialog(dlg *dialog.Dialog) {
	if dlg != nil {
		e.d = *dlg
	}
}

func (e *errorDialogWindow) Show() {
	if e.initialized {
		e.w.Show()
		e.d.Show()
	}
}

func (e *errorDialogWindow) Hide() {
	if e.initialized {
		e.d.Hide()
		e.w.Hide()
	}
}

func (e *errorDialogWindow) Callback() {
	if e.cb != nil {
		e.cb()
	}
	e.w.Hide()
	CloseWindow(e.w)
}
