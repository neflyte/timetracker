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
	taskNameBinding            binding.String
	elapsedTimeBinding         binding.String
	taskRunningBinding         binding.Bool
	taskListBinding            binding.StringList
	taskRunningBindingListener binding.DataListener
	taskListBindingListener    binding.DataListener
	selectedTask               models.Task
	taskSelect                 *widget.Select
	container                  *fyne.Container
	startStopButton            *widget.Button
	createAndStartButton       *widget.Button
	taskNameLabel              *widget.Label
	elapsedTimeLabel           *widget.Label
	commandChan                chan rxgo.Item
	taskList                   []string
	log                        zerolog.Logger
	widget.BaseWidget
	selectedTaskIndex int
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
	err := c.taskNameBinding.Set("idle") // i18n
	if err != nil {
		return err
	}
	err = c.taskRunningBinding.Set(false)
	if err != nil {
		return err
	}
	c.taskRunningBindingListener = binding.NewDataListener(c.taskRunningWasUpdated)
	c.taskRunningBinding.AddListener(c.taskRunningBindingListener)
	c.taskListBindingListener = binding.NewDataListener(c.taskListWasUpdated)
	c.taskListBinding.AddListener(c.taskListBindingListener)
	return nil
}

/*
 * Public functions
 */

// Observable returns an rxgo Observable for the widget's command channel
func (c *CompactUI) Observable() rxgo.Observable {
	return rxgo.FromEventSource(c.commandChan)
}

// SetTaskList sets the list of tasks displayed in the task selector widget
func (c *CompactUI) SetTaskList(taskList []string) {
	log := logger.GetFuncLogger(c.log, "SetTaskList")
	if taskList == nil {
		log.Error().
			Msg("cannot set task list with a nil list")
		return
	}
	log = log.With().
		Int("taskCount", len(taskList)).
		Logger()
	err := c.taskListBinding.Set(taskList)
	if err != nil {
		log.Err(err).
			Msg("unable to set task list binding")
		return
	}
	c.taskList = taskList
	log.Debug().
		Msg("set task list binding")
}

// SetRunning sets the "task is running" status of the UI
func (c *CompactUI) SetRunning(running bool) {
	log := logger.GetFuncLogger(c.log, "SetRunning").
		With().Bool("running", running).Logger()
	err := c.taskRunningBinding.Set(running)
	if err != nil {
		log.Err(err).
			Msg("unable to set task running binding")
		return
	}
	log.Debug().
		Msg("set task running status")
}

// SelectTask attempts to select the task with the specified name. If the name is empty
// or not present in the task list, the selected task will be cleared.
func (c *CompactUI) SelectTask(name string) {
	log := logger.GetFuncLogger(c.log, "SelectTask").
		With().Str("name", name).Logger()
	if name == "" {
		log.Debug().
			Msg("clearing selection; name was empty")
		c.taskSelect.ClearSelected()
		return
	} else if !slices.Contains(c.taskList, name) {
		log.Debug().
			Msg("clearing selection; taskList does not contain name")
		c.taskSelect.ClearSelected()
		return
	}
	log.Debug().
		Msg("set selected task")
	c.taskSelect.SetSelected(name)
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
			With().Str("elapsed", elapsed).Logger()
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

func (c *CompactUI) isRunning() bool {
	log := logger.GetFuncLogger(c.log, "isRunning")
	taskIsRunning, err := c.taskRunningBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get value from task running binding")
		return false
	}
	return taskIsRunning
}

// updateStartStopButton enables or disables the start/stop button
func (c *CompactUI) updateStartStopButton() {
	log := logger.GetFuncLogger(c.log, "updateStartStopButton")
	enable := false
	switch c.isRunning() {
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
	// assume a task is not running
	buttonIcon := theme.MediaPlayIcon()
	buttonText := "START" // i18n
	taskNameLabelStyle := fyne.TextStyle{
		Italic: true,
	}
	// if a task is actually running, use the correct button icon and text style
	if c.isRunning() {
		buttonIcon = theme.MediaStopIcon()
		buttonText = "STOP" // i18n
		taskNameLabelStyle = fyne.TextStyle{
			Bold: true,
		}
	}
	// update the start/stop button and task label
	c.startStopButton.SetIcon(buttonIcon)
	c.startStopButton.SetText(buttonText)
	if c.taskNameLabel.TextStyle != taskNameLabelStyle {
		c.taskNameLabel.TextStyle = taskNameLabelStyle
		defer c.taskNameLabel.Refresh()
	}
	if !c.isRunning() {
		c.taskNameLabel.SetText("idle") // i18n
	}
}

func (c *CompactUI) taskListWasUpdated() {
	log := logger.GetFuncLogger(c.log, "taskListWasUpdated")
	taskList, err := c.taskListBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get task list from binding")
		return
	}
	defer c.updateStartStopButton()
	// Update the Select widget's list after appending the 'other' item
	taskList = append(taskList, compactUIOtherTaskLabel)
	c.taskSelect.Options = taskList
	// If the selected task is in the new list, make that selection stand.
	if c.selectedTask != nil {
		c.selectedTaskIndex = slices.Index(taskList, c.selectedTask.Data().Synopsis)
		if c.selectedTaskIndex > -1 {
			c.taskSelect.SetSelectedIndex(c.selectedTaskIndex)
		} else {
			c.selectedTask = nil
			c.taskSelect.ClearSelected()
		}
	} else {
		// Need an explicit refresh here because the other code paths indirectly refresh
		// the taskSelect widget.
		c.taskSelect.Refresh()
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
			Msg("other task was selected")
		c.otherTaskWasSelected()
		return
	}
	log.Debug().Msg("searching for task by synopsis using selection")
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
	taskData := task[0]
	log.Debug().
		Uint("id", taskData.ID).
		Str("synopsis", taskData.Synopsis).
		Msg("found task")
	taskIndex := slices.Index(c.taskList, taskData.Synopsis)
	if taskIndex == -1 {
		log.Error().
			Uint("id", taskData.ID).
			Str("synopsis", taskData.Synopsis).
			Msg("could not find selected task in taskList; will not change selectedTask")
		return
	}
	c.selectedTaskIndex = taskIndex
	c.selectedTask = models.NewTaskWithData(taskData)
	log.Debug().
		Bool("selectedTask", c.selectedTask != nil).
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
	var task models.Task
	if !c.isRunning() && c.selectedTask != nil {
		synopsis = c.selectedTask.Data().Synopsis
		taskIndex = c.selectedTaskIndex
		task = c.selectedTask
	}
	log.Debug().
		Str("synopsis", synopsis).
		Int("index", taskIndex).
		Msg("sending task event to start task")
	c.commandChan <- rxgo.Of(CompactUITaskEvent{
		TaskSynopsis: synopsis,
		TaskIndex:    taskIndex,
		Task:         task,
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
