package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type ConfirmDialogWindow interface {
	Window() *fyne.Window
	Init() error
	Dialog() *dialog.ConfirmDialog
	SetDialog(dlg *dialog.ConfirmDialog)
	Show()
	Hide()
}

type confirmDialogWindow struct {
	w           fyne.Window
	d           dialog.ConfirmDialog
	cb          func(bool)
	initialSize *fyne.Size
	result      bool
	initialized bool
}

func NewConfirmDialogWindow(app fyne.App, title string, message string, size *fyne.Size, callback func(bool)) ConfirmDialogWindow {
	cdw := new(confirmDialogWindow)
	cdw.w = app.NewWindow(title)
	cdw.w.SetFixedSize(true)
	cdlg := dialog.NewConfirm(title, message, cdw.Callback, cdw.w)
	cdw.d = *cdlg
	cdw.cb = callback
	cdw.result = false
	cdw.initialized = false
	cdw.initialSize = &MinimumWindowSize
	if size != nil {
		cdw.initialSize = size
	}
	err := cdw.Init()
	if err != nil {
		return nil
	}
	return cdw
}

func (d *confirmDialogWindow) Window() *fyne.Window {
	return &d.w
}

func (d *confirmDialogWindow) Init() error {
	if !d.initialized {
		if d.initialSize != nil {
			d.w.Resize(*d.initialSize)
			d.d.Resize(*d.initialSize)
		}
		d.initialized = true
	}
	return nil
}

func (d *confirmDialogWindow) Dialog() *dialog.ConfirmDialog {
	return &d.d
}

func (d *confirmDialogWindow) SetDialog(dlg *dialog.ConfirmDialog) {
	if dlg != nil {
		d.d = *dlg
	}
}

func (d *confirmDialogWindow) Show() {
	if d.initialized {
		d.w.Show()
		d.d.Show()
	}
}

func (d *confirmDialogWindow) Hide() {
	if d.initialized {
		d.d.Hide()
		d.w.Hide()
	}
}

func (d *confirmDialogWindow) Callback(b bool) {
	d.result = b
	d.cb(b)
	d.w.Hide()
	CloseWindow(d.w)
}
