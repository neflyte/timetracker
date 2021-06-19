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

// StopTaskDialog is the main data structure for the Stop Task dialog
type StopTaskDialog struct {
	confirmDialog       dialog.Dialog
	messageLabel        *widget.Label
	closeWindowCheckbox *widget.Check
	closeWindowBinding  binding.Bool
	widgetContainer     *fyne.Container
	parentWindow        *fyne.Window
	log                 zerolog.Logger
}

// NewStopTaskDialog creates a new instance of the Stop Task dialog
func NewStopTaskDialog(task models.TaskData, prefs fyne.Preferences, cb func(bool), parent fyne.Window) *StopTaskDialog {
	newDialog := &StopTaskDialog{
		log:                logger.GetStructLogger("CreateAndStartTaskDialog"),
		closeWindowBinding: binding.BindPreferenceBool(prefKeyCloseWindowStopTask, prefs),
		parentWindow:       &parent,
		messageLabel:       widget.NewLabel(fmt.Sprintf("Do you want to stop task %s?", task.Synopsis)),
	}
	newDialog.closeWindowCheckbox = widget.NewCheckWithData("Close window after stopping task", newDialog.closeWindowBinding)
	newDialog.widgetContainer = container.NewVBox(
		newDialog.messageLabel,
		newDialog.closeWindowCheckbox,
	)
	newDialog.confirmDialog = dialog.NewCustomConfirm(
		"Stop Running Task",
		"YES",
		"NO",
		newDialog.widgetContainer,
		cb,
		parent,
	)
	return newDialog
}

// Show shows the dialog on the screen
func (d *StopTaskDialog) Show() {
	d.confirmDialog.Show()
}
