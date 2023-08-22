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

// CompactUITaskEvent represents an event which starts or stops a task
type CompactUITaskEvent struct {
	TaskSynopsis string
	TaskIndex    int
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
	c.taskSelect.PlaceHolder = "Select a task"                                                                               // i18n
	c.startStopButton = widget.NewButtonWithIcon("START", theme.MediaPlayIcon(), c.startStopButtonWasTapped)                 // i18n
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
	c.taskListBindingListener = binding.NewDataListener(c.taskListWasUpdated)
	c.taskListBinding.AddListener(c.taskListBindingListener)
	c.taskRunningBindingListener = binding.NewDataListener(c.taskRunningWasUpdated)
	c.taskRunningBinding.AddListener(c.taskRunningBindingListener)
	err := c.taskNameBinding.Set("idle") // i18n
	if err != nil {
		return err
	}
	err = c.taskRunningBinding.Set(false)
	if err != nil {
		return err
	}
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
	err := c.taskListBinding.Set(taskList)
	if err != nil {
		log.Err(err).
			Msg("unable to set task list binding")
		return
	}
	c.taskList = taskList
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
	log.Debug().Msg("set task running status")
}

// SelectTask attempts to select the task with the specified name. If the name is empty
// the selected task will be cleared.
func (c *CompactUI) SelectTask(name string) {
	if name == "" {
		c.taskSelect.ClearSelected()
		return
	}
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
	log.Debug().Msg("set task name")
}

// SetElapsedTime sets the display text of the elapsed time label
func (c *CompactUI) SetElapsedTime(elapsed string) {
	err := c.elapsedTimeBinding.Set(elapsed)
	if err != nil {
		log := logger.GetFuncLogger(c.log, "SetElapsedTime").
			With().Str("elapsed", elapsed).Logger()
		log.Err(err).
			Msg("unable to set elapsed time binding")
		return
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
	buttonText := "START" // i18n
	taskNameLabelStyle := fyne.TextStyle{
		Italic: true,
	}
	// if a task is actually running, use the correct button icon and text style
	if taskIsRunning {
		buttonIcon = theme.MediaStopIcon()
		buttonText = "STOP" // i18n
		taskNameLabelStyle = fyne.TextStyle{
			Bold: true,
		}
	}
	// update the start/stop button and task label
	c.startStopButton.SetIcon(buttonIcon)
	c.startStopButton.SetText(buttonText)
	c.taskNameLabel.TextStyle = taskNameLabelStyle
	if !taskIsRunning {
		c.taskNameLabel.SetText("idle") // i18n
	} else {
		c.taskNameLabel.Refresh()
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
	// TODO: if the list was updated and the selected task is in the new list, make that selection stand.
	// Clear the current selection
	c.selectedTask = nil
	c.selectedTaskIndex = -1
	c.taskSelect.ClearSelected()
	// Add the 'other' item to the list
	taskList = append(taskList, compactUIOtherTaskLabel)
	// Update the Select widget's list
	c.taskSelect.Options = taskList
	// Refresh the Select widget
	c.taskSelect.Refresh()
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
	c.selectedTask = models.NewTaskWithData(taskData)
	c.selectedTaskIndex = slices.Index(c.taskList, selection)
	if c.selectedTaskIndex == -1 {
		log.Error().
			Msg("selectedTaskIndex was -1; this is unexpected")
	}
	log.Debug().
		Bool("selectedTask", c.selectedTask != nil).
		Int("selectedTaskIndex", c.selectedTaskIndex).
		Msg("selected task")
}

func (c *CompactUI) startStopButtonWasTapped() {
	log := logger.GetFuncLogger(c.log, "startStopButtonWasTapped")
	taskIsRunning, err := c.taskRunningBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get value from task running binding")
		return
	}
	if taskIsRunning {
		log.Debug().
			Bool("taskIsRunning", taskIsRunning).
			Msg("sending task event to stop running task")
		c.commandChan <- rxgo.Of(CompactUITaskEvent{
			TaskSynopsis: "",
			TaskIndex:    -1,
		})
		return
	}
	synopsis := ""
	if c.selectedTask != nil {
		synopsis = c.selectedTask.Data().Synopsis
	}
	log.Debug().
		Str("synopsis", synopsis).
		Int("index", c.selectedTaskIndex).
		Msg("sending task event to start task")
	c.commandChan <- rxgo.Of(CompactUITaskEvent{
		TaskSynopsis: synopsis,
		TaskIndex:    c.selectedTaskIndex,
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
	c.selectedTask = nil
	c.selectedTaskIndex = -1
	c.commandChan <- rxgo.Of(CompactUISelectTaskEvent{})
	c.taskSelect.ClearSelected()
}

func (c *CompactUI) createAndStartWasTapped() {
	c.commandChan <- rxgo.Of(CompactUICreateAndStartEvent{})
}

func (c *CompactUI) aboutButtonWasTapped() {
	c.commandChan <- rxgo.Of(CompactUIAboutEvent{})
}
