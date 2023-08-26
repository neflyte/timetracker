package dialogs

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/rs/zerolog"
)

// StopTaskDialog is the main interface for the Stop Task dialog
type StopTaskDialog interface {
	dialogBase
	SetCloseWindowCheckbox(hidden bool)
}

// stopTaskDialogData is the main data structure for the Stop Task dialog
type stopTaskDialogData struct {
	dialog.Dialog
	closeWindowBinding      binding.Bool
	messageLabel            *widget.Label
	closeWindowCheckbox     *widget.Check
	widgetContainer         *fyne.Container
	parentWindow            *fyne.Window
	callbackFunc            func(bool)
	log                     zerolog.Logger
	hideCloseWindowCheckbox bool
}

// NewStopTaskDialog creates a new instance of the Stop Task dialog
func NewStopTaskDialog(task models.TaskData, prefs fyne.Preferences, cb func(bool), parent fyne.Window) StopTaskDialog {
	newDialog := &stopTaskDialogData{
		log:                logger.GetStructLogger("stopTaskDialogData"),
		closeWindowBinding: binding.BindPreferenceBool(constants.PrefKeyCloseWindowStopTask, prefs),
		parentWindow:       &parent,
		messageLabel:       widget.NewLabel(fmt.Sprintf("Do you want to stop task %s?", task.Synopsis)), // i18n
		callbackFunc:       cb,
	}
	err := newDialog.Init()
	if err != nil {
		newDialog.log.
			Err(err).
			Msg("error initializing dialog")
	}
	return newDialog
}

// Init initializes the dialog
func (d *stopTaskDialogData) Init() error {
	d.closeWindowCheckbox = widget.NewCheckWithData("Close window after stopping task", d.closeWindowBinding) // i18n
	d.widgetContainer = container.NewVBox(
		d.messageLabel,
		d.closeWindowCheckbox,
	)
	d.Dialog = dialog.NewCustomConfirm(
		"Stop Running Task", // i18n
		"YES",               // i18n
		"NO",                // i18n
		d.widgetContainer,
		d.callbackFunc,
		*d.parentWindow,
	)
	return nil
}

func (d *stopTaskDialogData) SetCloseWindowCheckbox(hidden bool) {
	d.hideCloseWindowCheckbox = hidden
	defer d.closeWindowCheckbox.Refresh()
	if hidden {
		d.closeWindowCheckbox.Hide()
		return
	}
	d.closeWindowCheckbox.Show()
}
