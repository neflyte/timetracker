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

const (
	// prefKeyCloseWindow is the preferences key for the flag which causes the main window to close after creating a new task
	prefKeyCloseWindow = "close-window"
)

// CreateAndStartTaskDialog is the main interface for the Create and Start New Task dialog
type CreateAndStartTaskDialog interface {
	dialogBase
	GetTask() *models.TaskData
	Reset()
	HideCloseWindowCheckbox()
	ShowCloseWindowCheckbox()
}

// createAndStartTaskDialogData is the main data structure for the Create and Start New Task dialog
type createAndStartTaskDialogData struct {
	dialog.Dialog

	synopsisLabel           *widget.Label
	synopsisEntry           *widget.Entry
	synopsisBinding         binding.String
	descriptionLabel        *widget.Label
	descriptionEntry        *widget.Entry
	descriptionBinding      binding.String
	closeWindowCheckbox     *widget.Check
	closeWindowBinding      binding.Bool
	showCloseWindowCheckbox bool
	widgetContainer         *fyne.Container
	parentWindow            *fyne.Window
	log                     zerolog.Logger
	callbackFunc            func(bool)
}

// NewCreateAndStartTaskDialog creates a new instance of the Create and Start a New Task dialog
func NewCreateAndStartTaskDialog(prefs fyne.Preferences, cb func(bool), parent fyne.Window) CreateAndStartTaskDialog {
	newDialog := &createAndStartTaskDialogData{
		log:                     logger.GetStructLogger("createAndStartTaskDialogData"),
		synopsisLabel:           widget.NewLabel("Synopsis:"),
		descriptionLabel:        widget.NewLabel("Description:"),
		synopsisBinding:         binding.NewString(),
		descriptionBinding:      binding.NewString(),
		closeWindowBinding:      binding.BindPreferenceBool(prefKeyCloseWindow, prefs),
		showCloseWindowCheckbox: true,
		parentWindow:            &parent,
		callbackFunc:            cb,
	}
	err := newDialog.Init()
	if err != nil {
		newDialog.log.Err(err).Msg("error initializing dialog")
	}
	return newDialog
}

func (c *createAndStartTaskDialogData) Init() error {
	c.synopsisEntry = widget.NewEntryWithData(c.synopsisBinding)
	c.synopsisEntry.SetPlaceHolder("enter the task synopsis here")
	c.descriptionEntry = widget.NewEntryWithData(c.descriptionBinding)
	c.descriptionEntry.SetPlaceHolder("enter the task description here")
	c.descriptionEntry.MultiLine = true
	c.descriptionEntry.Wrapping = fyne.TextWrapWord
	c.closeWindowCheckbox = widget.NewCheckWithData("Close window after starting task", c.closeWindowBinding)
	c.widgetContainer = container.NewVBox(
		c.synopsisLabel,
		c.synopsisEntry,
		c.descriptionLabel,
		c.descriptionEntry,
		c.closeWindowCheckbox,
	)
	c.Dialog = dialog.NewCustomConfirm(
		"Create and start a new task",
		"CREATE AND START",
		"CLOSE",
		c.widgetContainer,
		c.doCallback,
		*c.parentWindow,
	)
	return nil
}

// HideCloseWindowCheckbox hides the Close Window checkbox
func (c *createAndStartTaskDialogData) HideCloseWindowCheckbox() {
	log := logger.GetFuncLogger(c.log, "HideCloseWindowCheckbox")
	c.showCloseWindowCheckbox = false
	log.Debug().Msgf("c.showCloseWindowCheckbox: %t", c.showCloseWindowCheckbox)
	c.closeWindowCheckbox.Hide()
}

// ShowCloseWindowCheckbox shows the Close Window checkbox
func (c *createAndStartTaskDialogData) ShowCloseWindowCheckbox() {
	log := logger.GetFuncLogger(c.log, "ShowCloseWindowCheckbox")
	c.showCloseWindowCheckbox = true
	log.Debug().Msgf("c.showCloseWindowCheckbox: %t", c.showCloseWindowCheckbox)
	c.closeWindowCheckbox.Show()
}

// GetTask returns the model.TaskData representing the newly input task details
func (c *createAndStartTaskDialogData) GetTask() *models.TaskData {
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
	task := models.NewTask()
	task.Data().Synopsis = synopsis
	task.Data().Description = description
	return task.Data()
}

// Reset clears the dialog fields for further re-use
func (c *createAndStartTaskDialogData) Reset() {
	log := logger.GetFuncLogger(c.log, "Reset")
	err := c.synopsisBinding.Set("")
	if err != nil {
		log.Err(err).Msg("error setting synopsis through binding")
	}
	err = c.descriptionBinding.Set("")
	if err != nil {
		log.Err(err).Msg("error setting description through binding")
	}
	c.ShowCloseWindowCheckbox() // reset close window checkbox to default
}

func (c *createAndStartTaskDialogData) doCallback(createAndStart bool) {
	log := logger.GetFuncLogger(c.log, "doCallback")
	if c.callbackFunc != nil {
		log.Debug().Msgf("callbackFunc is not nil; calling c.callbackFunc(%t)", createAndStart)
		c.callbackFunc(createAndStart)
	}
	// If the close window checkbox is visible, handle the close window case
	log.Debug().Msgf("showCloseWindowCheckbox: %t", c.showCloseWindowCheckbox)
	if c.showCloseWindowCheckbox {
		shouldCloseWindow, err := c.closeWindowBinding.Get()
		if err == nil && shouldCloseWindow && c.parentWindow != nil {
			log.Debug().Msgf("shouldCloseWindow: %t; err == nil and c.parentWindow != nil", shouldCloseWindow)
			(*c.parentWindow).Close()
		}
	}
	// Reset ourselves if we didn't cancel
	if createAndStart {
		c.Reset()
	}
}
