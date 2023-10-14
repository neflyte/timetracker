package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
)

const (
	compactUICommandChanSize = 2
	compactUIOtherTaskLabel  = "Other..." // i18n
)

var (
	compactUIIdleTextStyle = fyne.TextStyle{
		Italic: true,
	}
	compactUIRunningTextStyle = fyne.TextStyle{
		Bold: true,
	}
)

/*
 * Event data structs
 */

// CompactUITaskEvent represents an event which starts or stops a task
type CompactUITaskEvent struct {
	// Task is the task object
	Task models.Task
	// TaskSynopsis is the synopsis of the task
	TaskSynopsis string
	// TaskIndex is the index of the task in the taskList
	TaskIndex int
}

// ShouldStopTask determines if the task event is meant to stop the running task
func (c *CompactUITaskEvent) ShouldStopTask() bool {
	return c.TaskIndex == -1 || c.TaskSynopsis == "" || c.Task == nil
}

// CompactUIManageEvent represents an event which opens the Manage window
type CompactUIManageEvent struct{}

// CompactUIReportEvent represents an event which opens the Report window
type CompactUIReportEvent struct{}

// CompactUIQuitEvent represents an event which exits the application
type CompactUIQuitEvent struct{}

// CompactUIAboutEvent represents an event which opens the About window
type CompactUIAboutEvent struct{}

// CompactUISelectTaskEvent represents an event which opens the task selector
type CompactUISelectTaskEvent struct{}

// CompactUICreateAndStartEvent represents an event which creates and starts a new task
type CompactUICreateAndStartEvent struct{}

/*
 * Main data struct
 */

var _ fyne.Widget = (*CompactUI)(nil)

// CompactUI is a compact user interface for the main Timetracker window
type CompactUI struct {
	log                  zerolog.Logger
	taskNameBinding      binding.String
	elapsedTimeBinding   binding.String
	selectedTask         models.Task
	createAndStartButton *widget.Button
	startStopButton      *widget.Button
	container            *fyne.Container
	taskNameLabel        *widget.Label
	elapsedTimeLabel     *widget.Label
	commandChan          chan rxgo.Item
	taskSelect           *widget.Select
	taskList             []string
	taskModels           models.TaskList
	widget.BaseWidget
	selectedTaskIndex int
	taskIsRunning     bool
}

// NewCompactUI creates a new instance of the compact user interface
func NewCompactUI() *CompactUI {
	compactui := &CompactUI{
		log:                logger.GetStructLogger("CompactUI"),
		taskList:           make([]string, 0),
		taskModels:         make(models.TaskList, 0),
		taskNameBinding:    binding.NewString(),
		elapsedTimeBinding: binding.NewString(),
		commandChan:        make(chan rxgo.Item, compactUICommandChanSize),
		selectedTaskIndex:  -1,
	}
	compactui.ExtendBaseWidget(compactui)
	compactui.initUI()
	err := compactui.initBindings()
	if err != nil {
		compactui.log.
			Err(err).
			Msg("error initializing bindings")
	}
	return compactui
}

/*
 * Initialization functions
 */

func (c *CompactUI) initUI() {
	c.taskSelect = widget.NewSelect(c.taskList, c.taskWasSelected)
	c.taskSelect.PlaceHolder = "Select a task"                                                               // i18n
	c.startStopButton = widget.NewButtonWithIcon("START", theme.MediaPlayIcon(), c.startStopButtonWasTapped) // i18n
	c.startStopButton.Disable()
	c.createAndStartButton = widget.NewButtonWithIcon("CREATE AND START", theme.ContentAddIcon(), c.createAndStartWasTapped) // i18n
	c.taskNameLabel = widget.NewLabelWithData(c.taskNameBinding)
	c.taskNameLabel.TextStyle = compactUIIdleTextStyle
	c.elapsedTimeLabel = widget.NewLabelWithData(c.elapsedTimeBinding)
	c.container = container.NewVBox(
		c.taskSelect,
		container.NewHBox(
			c.startStopButton,
			c.taskNameLabel,
			c.elapsedTimeLabel,
		),
		c.createAndStartButton,
		container.NewHBox(
			widget.NewButtonWithIcon("MANAGE", theme.SettingsIcon(), c.manageButtonWasTapped),       // i18n
			widget.NewButtonWithIcon("REPORT", theme.DocumentCreateIcon(), c.reportButtonWasTapped), // i18n
			widget.NewButtonWithIcon("QUIT", theme.LogoutIcon(), c.quitButtonWasTapped),             // i18n
			widget.NewButtonWithIcon("", theme.InfoIcon(), c.aboutButtonWasTapped),
		),
	)
}

func (c *CompactUI) initBindings() error {
	return c.taskNameBinding.Set("idle") // i18n
}

/*
 * Public functions
 */

// Observable returns an RxGo Observable for the widget's command channel
func (c *CompactUI) Observable() rxgo.Observable {
	return rxgo.FromEventSource(c.commandChan)
}

// SetTaskList sets the list of tasks displayed in the task selector widget
func (c *CompactUI) SetTaskList(taskModels models.TaskList) {
	log := logger.GetFuncLogger(c.log, "SetTaskList")
	if taskModels == nil {
		log.Error().
			Msg("cannot set task list with a nil list")
		return
	}
	c.taskModels = taskModels
	log = log.With().
		Int("modelCount", len(c.taskModels)).
		Logger()
	c.taskList = c.taskModels.Names()
	log.Debug().
		Int("taskCount", len(c.taskList)).
		Msg("set task list binding")
	c.taskListWasUpdated(c.taskList)
}

// SetRunning sets the "task is running" status of the UI
func (c *CompactUI) SetRunning(running bool) {
	log := logger.GetFuncLogger(c.log, "SetRunning").
		With().Bool("running", running).Logger()
	c.taskIsRunning = running
	log.Debug().
		Msg("set task running status")
	c.taskRunningWasUpdated()
}

// IsRunning returns the "task is running" status of the UI
func (c *CompactUI) IsRunning() bool {
	return c.taskIsRunning
}

// SelectTask attempts to select the specified task from the list. If the task is nil,
// the selected task will be cleared.
func (c *CompactUI) SelectTask(task models.Task) {
	log := logger.GetFuncLogger(c.log, "SelectTask")
	if task == nil {
		log.Debug().
			Msg("clearing selection; name was empty or taskList does not contain task")
		c.taskSelect.ClearSelected()
		return
	}
	log.Debug().
		Str("task", task.DisplayString()).
		Msg("set selected task")
	c.taskSelect.SetSelected(task.DisplayString())
}

// SetSelectedTask sets the selected task. The display will not be updated.
func (c *CompactUI) SetSelectedTask(task models.Task) {
	log := logger.GetFuncLogger(c.log, "SetSelectedTask")
	c.selectedTask = task
	c.selectedTaskIndex = -1
	selectedTaskPresent := task != nil
	if selectedTaskPresent {
		c.selectedTaskIndex = slices.Index(c.taskList, task.DisplayString())
	}
	log.Debug().
		Bool("selectedTaskPresent", selectedTaskPresent).
		Int("selectedTaskIndex", c.selectedTaskIndex).
		Msg("set selectedTask")
}

// SetTaskName sets the display text of the task name label
func (c *CompactUI) SetTaskName(name string) {
	log := logger.GetFuncLogger(c.log, "SetTaskName").
		With().Str("name", name).Logger()
	err := c.taskNameBinding.Set(name)
	if err != nil {
		log.Err(err).
			Msg("unable to set task name binding")
		return
	}
	log.Debug().Msg("set task name binding")
}

// SetElapsedTime sets the display text of the elapsed time label
func (c *CompactUI) SetElapsedTime(elapsed string) {
	err := c.elapsedTimeBinding.Set(elapsed)
	if err != nil {
		log := logger.GetFuncLogger(c.log, "SetElapsedTime").
			With().Str("elapsed", elapsed).
			Logger()
		log.Err(err).
			Msg("unable to set elapsed time binding")
	}
	// There is no log message here because this function could be called
	// often enough that it would only add noise to the logs.
}

// CreateRenderer returns a new WidgetRenderer for this widget
func (c *CompactUI) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}

/*
 * Private functions
 */

// updateStartStopButton enables or disables the start/stop button
func (c *CompactUI) updateStartStopButton() {
	log := logger.GetFuncLogger(c.log, "updateStartStopButton")
	enable := false
	switch c.IsRunning() {
	case true:
		enable = true
	case false:
		if c.selectedTask != nil {
			enable = true
		}
	}
	switch enable {
	case true:
		c.startStopButton.Enable()
	case false:
		c.startStopButton.Disable()
	}
	log.Debug().
		Bool("enable", enable).
		Msg("set start/stop button state")
}

func (c *CompactUI) taskRunningWasUpdated() {
	// log := logger.GetFuncLogger(c.log, "taskRunningWasUpdated")
	defer c.updateStartStopButton()
	isRunning := c.IsRunning()
	// assume a task is not running
	buttonIcon := theme.MediaPlayIcon()
	buttonText := "START" // i18n
	taskNameLabelStyle := compactUIIdleTextStyle
	// if a task is actually running, use the correct button icon and text style
	if isRunning {
		buttonIcon = theme.MediaStopIcon()
		buttonText = "STOP" // i18n
		taskNameLabelStyle = compactUIRunningTextStyle
	}
	// update the start/stop button and task label
	c.startStopButton.SetIcon(buttonIcon)
	c.startStopButton.SetText(buttonText)
	if c.taskNameLabel.TextStyle != taskNameLabelStyle {
		c.taskNameLabel.TextStyle = taskNameLabelStyle
		defer c.taskNameLabel.Refresh()
	}
	if !isRunning {
		c.taskNameLabel.SetText("idle") // i18n
	}
}

func (c *CompactUI) taskListWasUpdated(taskNames []string) {
	log := logger.GetFuncLogger(c.log, "taskListWasUpdated")
	log.Debug().Msg("taskNames was updated")
	defer c.updateStartStopButton()
	// Append the 'other' item
	optionsList := make([]string, len(taskNames)+1)
	copy(optionsList, taskNames)
	optionsList[len(optionsList)-1] = compactUIOtherTaskLabel
	// Update the options
	c.taskSelect.SetOptions(optionsList)
	// If the selected task is in the new list, make that selection stand.
	if c.selectedTask != nil {
		c.selectedTaskIndex = c.taskModels.Index(c.selectedTask)
		if c.selectedTaskIndex == -1 {
			c.taskSelect.ClearSelected()
			return
		}
		log.Debug().
			Str("selectedTask", c.selectedTask.DisplayString()).
			Int("selectedTaskIndex", c.selectedTaskIndex).
			Msg("selected task is present")
	}
}

func (c *CompactUI) taskWasSelected(selection string) {
	log := logger.GetFuncLogger(c.log, "taskWasSelected").
		With().Str("selection", selection).
		Logger()
	if selection == "" {
		log.Debug().
			Msg("selection was empty; nothing to do")
		return
	}
	if selection == compactUIOtherTaskLabel {
		log.Debug().
			Msg("'other...' was selected")
		c.otherTaskWasSelected()
		return
	}
	selectedIndex := slices.Index(c.taskList, selection)
	if selectedIndex == -1 {
		log.Error().
			Str("selection", selection).
			Msg("unable to find task in taskList")
		return
	}
	if selectedIndex == c.selectedTaskIndex && c.IsRunning() {
		log.Warn().
			Int("selectedIndex", selectedIndex).
			Int("c.selectedTaskIndex", c.selectedTaskIndex).
			Msg("index values are equal and task is running; suppressing CompactUITaskEvent")
		return
	}
	c.selectedTaskIndex = selectedIndex
	c.selectedTask = c.taskModels[selectedIndex]
	log.Debug().
		Str("selectedTask", c.selectedTask.String()).
		Int("selectedTaskIndex", c.selectedTaskIndex).
		Msg("selected task")
	c.commandChan <- rxgo.Of(CompactUITaskEvent{
		TaskIndex:    c.selectedTaskIndex,
		TaskSynopsis: c.selectedTask.Data().Synopsis,
		Task:         c.selectedTask,
	})
}

func (c *CompactUI) startStopButtonWasTapped() {
	log := logger.GetFuncLogger(c.log, "startStopButtonWasTapped")
	synopsis := ""
	taskIndex := -1
	var selectedTask models.Task
	selectedTaskIsPresent := false
	if !c.IsRunning() && c.selectedTask != nil {
		synopsis = c.selectedTask.Data().Synopsis
		taskIndex = c.selectedTaskIndex
		selectedTask = c.selectedTask
		selectedTaskIsPresent = true
	}
	log.Debug().
		Str("synopsis", synopsis).
		Int("taskIndex", taskIndex).
		Bool("selectedTaskIsPresent", selectedTaskIsPresent).
		Msg("sending task event")
	c.commandChan <- rxgo.Of(CompactUITaskEvent{
		TaskSynopsis: synopsis,
		TaskIndex:    taskIndex,
		Task:         selectedTask,
	})
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
	c.commandChan <- rxgo.Of(CompactUISelectTaskEvent{})
}

func (c *CompactUI) createAndStartWasTapped() {
	c.commandChan <- rxgo.Of(CompactUICreateAndStartEvent{})
}

func (c *CompactUI) aboutButtonWasTapped() {
	c.commandChan <- rxgo.Of(CompactUIAboutEvent{})
}
