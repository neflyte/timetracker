package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

// CreateAndStartTaskDialog is the main data structure for the Create and Start New Task dialog
type CreateAndStartTaskDialog struct {
	confirmDialog      dialog.Dialog
	synopsisLabel      *widget.Label
	synopsisEntry      *widget.Entry
	synopsisBinding    binding.String
	descriptionLabel   *widget.Label
	descriptionEntry   *widget.Entry
	descriptionBinding binding.String
	widgetContainer    *fyne.Container
	parentWindow       *fyne.Window
	log                zerolog.Logger
}

// NewCreateAndStartTaskDialog creates a new instance of the Create and Start a New Task dialog
func NewCreateAndStartTaskDialog(cb func(bool), parent fyne.Window) *CreateAndStartTaskDialog {
	newDialog := &CreateAndStartTaskDialog{
		log:                logger.GetStructLogger("CreateAndStartTaskDialog"),
		synopsisLabel:      widget.NewLabel("Synopsis:"),
		descriptionLabel:   widget.NewLabel("Description:"),
		synopsisBinding:    binding.NewString(),
		descriptionBinding: binding.NewString(),
		parentWindow:       &parent,
	}
	newDialog.synopsisEntry = widget.NewEntryWithData(newDialog.synopsisBinding)
	newDialog.synopsisEntry.SetPlaceHolder("enter the task synopsis here")
	newDialog.descriptionEntry = widget.NewEntryWithData(newDialog.descriptionBinding)
	newDialog.descriptionEntry.SetPlaceHolder("enter the task description here")
	newDialog.descriptionEntry.MultiLine = true
	newDialog.descriptionEntry.Wrapping = fyne.TextWrapWord
	newDialog.widgetContainer = container.NewVBox(
		newDialog.synopsisLabel,
		newDialog.synopsisEntry,
		newDialog.descriptionLabel,
		newDialog.descriptionEntry,
	)
	newDialog.confirmDialog = dialog.NewCustomConfirm(
		"Create and start a new task",
		"CREATE AND START",
		"CLOSE",
		newDialog.widgetContainer,
		cb,
		parent,
	)
	return newDialog
}

// Show shows the dialog on the screen
func (c *CreateAndStartTaskDialog) Show() {
	c.confirmDialog.Show()
}

// GetTask returns the model.TaskData representing the newly input task details
func (c *CreateAndStartTaskDialog) GetTask() *models.TaskData {
	log := logger.GetFuncLogger(c.log, "GetTask")
	synopsis, err := c.synopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting synopsis from binding")
		return nil
	}
	description, err := c.descriptionBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting description from binding")
		return nil
	}
	taskData := models.NewTaskData()
	taskData.Synopsis = synopsis
	taskData.Description = description
	return taskData
}

// Reset clears the dialog fields for further re-use
func (c *CreateAndStartTaskDialog) Reset() {
	log := logger.GetFuncLogger(c.log, "Reset")
	err := c.synopsisBinding.Set("")
	if err != nil {
		log.Err(err).Msg("error setting synopsis through binding")
	}
	err = c.descriptionBinding.Set("")
	if err != nil {
		log.Err(err).Msg("error setting description through binding")
	}
}
