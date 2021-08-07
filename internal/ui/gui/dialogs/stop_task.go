package dialogs

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

const (
	// prefKeyCloseWindowStopTask is the preferences key for the flag which causes the main window to close after creating a new task
	prefKeyCloseWindowStopTask = "close-window:stop-task"
)

// StopTaskDialog is the main interface for the Stop Task dialog
type StopTaskDialog interface {
	dialogBase
}

// stopTaskDialogData is the main data structure for the Stop Task dialog
type stopTaskDialogData struct {
	dialog.Dialog

	messageLabel        *widget.Label
	closeWindowCheckbox *widget.Check
	closeWindowBinding  binding.Bool
	widgetContainer     *fyne.Container
	parentWindow        *fyne.Window
	log                 zerolog.Logger
	callbackFunc        func(bool)
}

// NewStopTaskDialog creates a new instance of the Stop Task dialog
func NewStopTaskDialog(task models.TaskData, prefs fyne.Preferences, cb func(bool), parent fyne.Window) StopTaskDialog {
	newDialog := &stopTaskDialogData{
		log:                logger.GetStructLogger("createAndStartTaskDialogData"),
		closeWindowBinding: binding.BindPreferenceBool(prefKeyCloseWindowStopTask, prefs),
		parentWindow:       &parent,
		messageLabel:       widget.NewLabel(fmt.Sprintf("Do you want to stop task %s?", task.Synopsis)),
		callbackFunc:       cb,
	}
	err := newDialog.Init()
	if err != nil {
		newDialog.log.Err(err).Msg("error initializing dialog")
	}
	return newDialog
}

// Init initializes the dialog
func (d *stopTaskDialogData) Init() error {
	d.closeWindowCheckbox = widget.NewCheckWithData("Close window after stopping task", d.closeWindowBinding)
	d.widgetContainer = container.NewVBox(
		d.messageLabel,
		d.closeWindowCheckbox,
	)
	d.Dialog = dialog.NewCustomConfirm(
		"Stop Running Task",
		"YES",
		"NO",
		d.widgetContainer,
		d.callbackFunc,
		*d.parentWindow,
	)
	return nil
}
