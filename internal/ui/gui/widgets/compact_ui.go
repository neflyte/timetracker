package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
)

const (
	compactUICommandChanSize = 2
	compactUIOtherTaskLabel  = "Other..." // i18n
)

// CompactUIStartTaskEvent represents an event which starts a task
type CompactUIStartTaskEvent struct {
	TaskSynopsis string
	TaskIndex    int
}

// CompactUIManageEvent represents an event which opens the Manage window
type CompactUIManageEvent struct{}

// CompactUIReportEvent represents an event which opens the Report window
type CompactUIReportEvent struct{}

// CompactUIQuitEvent represents an event which exits the application
type CompactUIQuitEvent struct{}

// CompactUIStopTaskEvent represents an event which starts a task
type CompactUIStopTaskEvent struct{}

// CompactUISelectTaskEvent represents an event which opens the task selector
type CompactUISelectTaskEvent struct{}

var _ fyne.Widget = (*CompactUI)(nil)

// CompactUI is a compact user interface for the main Timetracker window
type CompactUI struct {
	taskNameBinding            binding.String
	elapsedTimeBinding         binding.String
	taskRunningBinding         binding.Bool
	taskRunningBindingListener binding.DataListener
	container                  *fyne.Container
	taskSelect                 *widget.Select
	taskListBinding            binding.StringList
	taskListBindingListener    binding.DataListener
	startStopButton            *widget.Button
	taskNameLabel              *widget.Label
	elapsedTimeLabel           *widget.Label
	selectedTask               models.Task
	selectedTaskIndex          int
	commandChan                chan rxgo.Item
	log                        zerolog.Logger
	taskList                   []string
	widget.BaseWidget
}

// NewCompactUI creates a new instance of the compact user interface
func NewCompactUI() *CompactUI {
	compactui := &CompactUI{
		log:                logger.GetStructLogger("CompactUI"),
		taskList:           make([]string, 0),
		taskListBinding:    binding.NewStringList(),
		taskNameBinding:    binding.NewString(),
		taskRunningBinding: binding.NewBool(),
		elapsedTimeBinding: binding.NewString(),
		commandChan:        make(chan rxgo.Item, compactUICommandChanSize),
	}
	compactui.ExtendBaseWidget(compactui)
	compactui.initUI()
	compactui.initBindings()
	return compactui
}

// initialization functions

func (c *CompactUI) initUI() {
	c.taskSelect = widget.NewSelect(c.taskList, c.taskWasSelected)
	c.taskSelect.PlaceHolder = "Select a task" // i18n
	c.startStopButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), c.startStopButtonWasTapped)
	c.taskNameLabel = widget.NewLabelWithData(c.taskNameBinding)
	c.elapsedTimeLabel = widget.NewLabelWithData(c.elapsedTimeBinding)
	c.container = container.NewVBox(
		c.taskSelect,
		container.NewHBox(
			c.startStopButton,
			c.taskNameLabel,
			c.elapsedTimeLabel,
		),
		container.NewHBox(
			widget.NewButtonWithIcon("Manage", theme.SettingsIcon(), c.manageButtonWasTapped),       // i18n
			widget.NewButtonWithIcon("Report", theme.DocumentCreateIcon(), c.reportButtonWasTapped), // i18n
			widget.NewButtonWithIcon("Quit", theme.LogoutIcon(), c.quitButtonWasTapped),             // i18n
		),
	)
}

func (c *CompactUI) initBindings() {
	c.taskListBindingListener = binding.NewDataListener(c.taskListWasUpdated)
	c.taskListBinding.AddListener(c.taskListBindingListener)
	c.taskRunningBindingListener = binding.NewDataListener(c.taskRunningWasUpdated)
	c.taskRunningBinding.AddListener(c.taskRunningBindingListener)
}

// public methods

// Observable returns an rxgo Observable for the widget's command channel
func (c *CompactUI) Observable() rxgo.Observable {
	return rxgo.FromEventSource(c.commandChan)
}

// SetRunning sets the "task is running" status of the UI
func (c *CompactUI) SetRunning(running bool) {
	log := logger.GetFuncLogger(c.log, "SetRunning").
		With().Bool("running", running).Logger()
	err := c.taskRunningBinding.Set(running)
	if err != nil {
		log.Err(err).
			Msg("unable to set task running binding")
	}
	log.Debug().Msg("set task running status")
}

// CreateRenderer returns a new WidgetRenderer for this widget
func (c *CompactUI) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}

// private methods

func (c *CompactUI) taskRunningWasUpdated() {
	log := logger.GetFuncLogger(c.log, "taskRunningWasUpdated")
	taskIsRunning, err := c.taskRunningBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get value from task running binding")
		return
	}
	// assume a task is not running
	buttonIcon := theme.MediaPlayIcon()
	// if a task is actually running, use the correct button icon
	if taskIsRunning {
		buttonIcon = theme.MediaStopIcon()
	}
	c.startStopButton.SetIcon(buttonIcon)
}

func (c *CompactUI) taskListWasUpdated() {
	log := logger.GetFuncLogger(c.log, "taskListWasUpdated")
	taskList, err := c.taskListBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get task list from binding")
		return
	}
	c.taskList = taskList
	c.taskSelect.Options = c.taskList
	c.taskSelect.Refresh()
}

func (c *CompactUI) taskWasSelected(selection string) {
	log := logger.GetFuncLogger(c.log, "taskWasSelected").
		With().Str("selection", selection).
		Logger()
	if selection == compactUIOtherTaskLabel {
		log.Debug().
			Msg("other task was selected")
		c.otherTaskWasSelected()
		return
	}
	task, err := models.NewTask().SearchBySynopsis(selection)
	if err != nil {
		log.Err(err).
			Msg("unable to lookup selected task by synopsis")
		return
	}
	if len(task) == 0 {
		log.Error().
			Msg("could not find selected task by synopsis")
		return
	}
	c.selectedTask = models.NewTaskWithData(task[0])
	c.selectedTaskIndex = slices.Index(c.taskList, selection)
	if c.selectedTaskIndex == -1 {
		log.Error().
			Msg("selectedTaskIndex was -1; this is unexpected")
	}
}

func (c *CompactUI) startStopButtonWasTapped() {
	log := logger.GetFuncLogger(c.log, "startStopButtonWasTapped")
	//c.startStopButton.Disable()
	//defer c.startStopButton.Enable()
	//runningTimesheets, err := models.NewTimesheet().SearchOpen()
	//if err != nil {
	//	log.Err(err).
	//		Msg("unable to check for running task")
	//	return
	//}
	//if len(runningTimesheets) > 0 {
	//	_, stopErr := models.NewTask().StopRunningTask()
	//	if stopErr != nil {
	//		log.Err(stopErr).
	//			Msg("unable to stop running task")
	//		return
	//	}
	//	c.startStopButton.SetIcon(theme.MediaPlayIcon())
	//	return
	//}
	if c.selectedTask == nil || c.selectedTaskIndex == -1 {
		log.Error().
			Msg("no task is selected")
		return
	}
	// update button
	// c.startStopButton.SetIcon(theme.MediaStopIcon())
}

func (c *CompactUI) manageButtonWasTapped() {
	c.commandChan <- rxgo.Of(CompactUIManageEvent{})
}

func (c *CompactUI) reportButtonWasTapped() {
	c.commandChan <- rxgo.Of(CompactUIReportEvent{})
}

func (c *CompactUI) quitButtonWasTapped() {
	c.commandChan <- rxgo.Of(CompactUIQuitEvent{})
}

func (c *CompactUI) otherTaskWasSelected() {
	c.selectedTask = nil
	c.selectedTaskIndex = -1
	c.commandChan <- rxgo.Of(CompactUISelectTaskEvent{})
}
