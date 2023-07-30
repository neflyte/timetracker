package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

var _ fyne.Widget = (*CompactUI)(nil)

// CompactUI is a compact user interface for the main Timetracker window
type CompactUI struct {
	taskNameBinding    binding.String
	elapsedTimeBinding binding.String
	container          *fyne.Container
	taskSelect         *widget.Select
	startStopButton    *widget.Button
	taskNameLabel      *widget.Label
	elapsedTimeLabel   *widget.Label
	ellipsisButton     *widget.Button
	ellipsisMenu       *fyne.Menu
	ellipsisPopup      *widget.PopUpMenu
	selectedTask       models.Task
	log                zerolog.Logger
	taskList           []string
	widget.BaseWidget
}

// NewCompactUI creates a new instance of the compact user interface
func NewCompactUI() *CompactUI {
	compactui := &CompactUI{
		log:                logger.GetStructLogger("CompactUi"),
		taskList:           make([]string, 0),
		taskNameBinding:    binding.NewString(),
		elapsedTimeBinding: binding.NewString(),
	}
	compactui.ExtendBaseWidget(compactui)
	compactui.initUI()
	return compactui
}

func (c *CompactUI) initUI() {
	c.taskSelect = widget.NewSelect(c.taskList, c.taskWasSelected)
	c.taskSelect.PlaceHolder = "Select a task" // i18n
	c.startStopButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), c.startStopButtonWasTapped)
	c.taskNameLabel = widget.NewLabelWithData(c.taskNameBinding)
	c.elapsedTimeLabel = widget.NewLabelWithData(c.elapsedTimeBinding)
	c.ellipsisButton = widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), c.ellipsisButtonWasTapped)
	c.container = container.NewHBox(
		c.taskSelect,
		c.startStopButton,
		c.taskNameLabel,
		c.elapsedTimeLabel,
		c.ellipsisButton,
	)
	manageItem := fyne.NewMenuItem("Manage", c.manageMenuItemWasTapped)
	manageItem.Icon = theme.SettingsIcon()
	reportItem := fyne.NewMenuItem("Reports", c.reportMenuItemWasTapped)
	reportItem.Icon = theme.DocumentCreateIcon()
	quitItem := fyne.NewMenuItem("Quit", c.quitMenuItemWasTapped)
	quitItem.Icon = theme.LogoutIcon()
	c.ellipsisMenu = fyne.NewMenu("", manageItem, reportItem, quitItem)
	c.ellipsisPopup = widget.NewPopUpMenu(c.ellipsisMenu)
}

// RefreshTasks reloads the last n started tasks and updates the selection widget
func (c *CompactUI) RefreshTasks() {
	log := logger.GetFuncLogger(c.log, "RefreshTasks")
	lastStartedTasks, err := models.NewTimesheet().LastStartedTasks(10)
	if err != nil {
		log.Err(err).
			Msg("unable to get last started tasks")
		return
	}
	lastStartedTaskNames := make([]string, 0)
	for idx := range lastStartedTasks {
		lastStartedTaskNames = append(lastStartedTaskNames, lastStartedTasks[idx].Synopsis)
	}
	lastStartedTaskNames = append(lastStartedTaskNames, "Other...") // i18n
	c.taskList = lastStartedTaskNames
	c.taskSelect.Options = c.taskList
	c.taskSelect.Refresh()
}

func (c *CompactUI) taskWasSelected(selection string) {
	log := logger.GetFuncLogger(c.log, "taskWasSelected")
	task, err := models.NewTask().SearchBySynopsis(selection)
	if err != nil {
		log.Err(err).
			Str("selection", selection).
			Msg("unable to lookup selected task by synopsis")
		return
	}
	if len(task) == 0 {
		log.Error().
			Str("selection", selection).
			Msg("could not find selected task by synopsis")
		return
	}
	// TODO: handle the multiple result case
	c.selectedTask = models.NewTaskWithData(task[0])
}

func (c *CompactUI) startStopButtonWasTapped() {
	log := logger.GetFuncLogger(c.log, "startStopButtonWasTapped")
	c.startStopButton.Disable()
	defer c.startStopButton.Enable()
	runningTimesheets, err := models.NewTimesheet().SearchOpen()
	if err != nil {
		log.Err(err).
			Msg("unable to check for running task")
		return
	}
	if len(runningTimesheets) > 0 {
		_, stopErr := models.NewTask().StopRunningTask()
		if stopErr != nil {
			log.Err(stopErr).
				Msg("unable to stop running task")
			return
		}
		c.startStopButton.SetIcon(theme.MediaPlayIcon())
		return
	}
	if c.selectedTask == nil {
		// error: need to select a task
		return
	}
	// start task

	// update button
	c.startStopButton.SetIcon(theme.MediaStopIcon())
}

func (c *CompactUI) ellipsisButtonWasTapped() {
	// log := logger.GetFuncLogger(c.log, "ellipsisButtonWasTapped")
	c.ellipsisPopup.ShowAtPosition(c.ellipsisButton.Position())
}

func (c *CompactUI) manageMenuItemWasTapped() {

}

func (c *CompactUI) reportMenuItemWasTapped() {

}

func (c *CompactUI) quitMenuItemWasTapped() {

}

// CreateRenderer returns a new WidgetRenderer for this widget
func (c *CompactUI) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}
