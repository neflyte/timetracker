package widgets

import (
	slices "golang.org/x/exp/slices"
	"reflect"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
)

const (
	selectedTaskNone            = widget.ListItemID(-1)
	taskSelectorCommandChanSize = 2
	taskSelectorMinimumWidth    = float32(250)
)

// TaskSelectorSelectedEvent contains the task that is sent to the command channel when a selection happens
type TaskSelectorSelectedEvent struct {
	// SelectedTask is the selected task item; nil if there is no selection
	SelectedTask models.Task
}

// TaskSelectorErrorEvent contains an error that occurred during the lifetime of the widget
type TaskSelectorErrorEvent struct {
	// Err is the underlying error
	Err error
}

var _ fyne.Widget = (*TaskSelector)(nil)

// TaskSelector is the implementation of the task selector widget
type TaskSelector struct {
	filterBinding         binding.String
	filterBindingListener binding.DataListener
	tasksListBinding      binding.UntypedList
	container             *fyne.Container
	filterHBox            *fyne.Container
	filterEntry           *widget.Entry
	sortButton            *widget.Button
	tasksList             *widget.List
	commandChan           chan rxgo.Item
	log                   zerolog.Logger
	widget.BaseWidget
	selectedTask widget.ListItemID
}

// NewTaskSelector returns a pointer to a new, initialized instance of TaskSelector
func NewTaskSelector() *TaskSelector {
	ts := &TaskSelector{
		log:              logger.GetStructLogger("TaskSelector"),
		tasksListBinding: binding.NewUntypedList(),
		filterBinding:    binding.NewString(),
		selectedTask:     selectedTaskNone,
		commandChan:      make(chan rxgo.Item, taskSelectorCommandChanSize),
	}
	ts.ExtendBaseWidget(ts)
	ts.initUI()
	ts.filterBinding.AddListener(ts.filterBindingListener)
	return ts
}

// initUI initializes UI widgets
func (t *TaskSelector) initUI() {
	t.filterEntry = widget.NewEntryWithData(t.filterBinding)
	t.filterEntry.SetPlaceHolder("Filter tasks") // i18n
	t.filterEntry.Validator = nil
	t.filterBindingListener = binding.NewDataListener(func() {
		// Filter text has changed; re-filter the tasks.
		t.log.Debug().
			Msg("filterBinding changed; re-filtering tasks")
		t.FilterTasks()
	})
	t.sortButton = widget.NewButton("Sort", t.doShowSortMenu) // i18n
	t.filterHBox = container.NewBorder(nil, nil, nil, t.sortButton, t.filterEntry)
	t.tasksList = widget.NewListWithData(t.tasksListBinding, t.createTaskWidget, t.updateTaskWidget)
	t.tasksList.OnSelected = t.handleTaskSelected
	t.container = container.NewBorder(t.filterHBox, nil, nil, nil, t.tasksList)
}

// Observable returns an rxgo Observable for the widget's command channel
func (t *TaskSelector) Observable() rxgo.Observable {
	return rxgo.FromEventSource(t.commandChan)
}

// MinSize overrides the minimum size of this widget. A minimum width is enforced.
func (t *TaskSelector) MinSize() fyne.Size {
	minsize := t.BaseWidget.MinSize()
	if minsize.Width < taskSelectorMinimumWidth {
		minsize.Width = taskSelectorMinimumWidth
	}
	return minsize
}

// SetList sets the list of Task objects used in the selector
// Deprecated
func (t *TaskSelector) SetList(tasks models.TaskList) {
	log := logger.GetFuncLogger(t.log, "SetList")
	err := t.tasksListBinding.Set(tasks.ToSliceIntf())
	if err != nil {
		log.Err(err).
			Msg("unable to set tasksListBinding")
	}
	t.tasksList.UnselectAll()
	t.tasksList.Refresh()
	// Reset the selected task
	t.selectedTask = selectedTaskNone
}

// HasSelected indicates whether a selection has been made or not
func (t *TaskSelector) HasSelected() bool {
	return t.selectedTask != selectedTaskNone
}

// Selected returns the selected task or nil if there is no selection
func (t *TaskSelector) Selected() models.Task {
	if !t.HasSelected() {
		// nothing selected
		return nil
	}
	taskList := t.list()
	if len(taskList) == 0 {
		// empty list
		return nil
	}
	if t.selectedTask >= len(taskList) {
		// index out of bounds
		return nil
	}
	return taskList[t.selectedTask]
}

// Reset resets the widget to its default state
func (t *TaskSelector) Reset() {
	log := logger.GetFuncLogger(t.log, "Reset")
	t.tasksList.UnselectAll()
	t.selectedTask = selectedTaskNone
	t.filterBinding.RemoveListener(t.filterBindingListener)
	err := t.filterBinding.Set("")
	if err != nil {
		log.Err(err).
			Msg("error resetting filter binding")
	}
	t.filterBinding.AddListener(t.filterBindingListener)
}

// createTaskWidget creates new Task widgets for the tasksList widget
func (t *TaskSelector) createTaskWidget() fyne.CanvasObject {
	return NewTask()
}

// updateTaskWidget updates the supplied Task widget with the supplied Task model
func (t *TaskSelector) updateTaskWidget(item binding.DataItem, canvasObject fyne.CanvasObject) {
	log := logger.GetFuncLogger(t.log, "updateTaskWidget")
	// get task
	listBinding, ok := item.(binding.Untyped)
	if !ok {
		log.Error().
			Str("expected", "binding.Untyped").
			Str("actual", reflect.TypeOf(item).String()).
			Msg("item is of an unexpected type")
		return
	}
	taskIntf, err := listBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get task interface from listBinding")
		return
	}
	task, ok := taskIntf.(models.Task)
	if !ok {
		log.Error().
			Str("expected", "models.Task").
			Str("actual", reflect.TypeOf(taskIntf).String()).
			Msg("taskIntf is of an unexpected type")
		return
	}
	// get widget
	taskWidget, ok := canvasObject.(*Task)
	if !ok {
		log.Error().
			Str("expected", "*Task").
			Str("actual", reflect.TypeOf(canvasObject).String()).
			Msg("canvasObject is of an unexpected type")
		return
	}
	// update widget
	taskWidget.SetTask(task)
}

// handleTaskSelected unselects any existing selection and sets the selection to the provided ID
func (t *TaskSelector) handleTaskSelected(id widget.ListItemID) {
	// If there was an existing selection, unselect it and send a selection event
	if t.selectedTask != selectedTaskNone {
		t.tasksList.Unselect(t.selectedTask)
		t.commandChan <- rxgo.Of(TaskSelectorSelectedEvent{nil})
	}
	// Set the selected task
	t.selectedTask = id
	// Send a selection event with the new task ID
	t.commandChan <- rxgo.Of(TaskSelectorSelectedEvent{t.Selected()})
}

// list returns the slice of Task objects used in the selector
func (t *TaskSelector) list() models.TaskList {
	log := logger.GetFuncLogger(t.log, "List")
	taskListIntf, err := t.tasksListBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get list from binding")
		return make([]models.Task, 0)
	}
	return models.TaskListFromSliceIntf(taskListIntf)
}

// doShowSortMenu displays the task sort menu
func (t *TaskSelector) doShowSortMenu() {
	log := logger.GetFuncLogger(t.log, "doShowSortMenu")
	log.Warn().
		Msg("implementation missing")
}

// FilterTasks loads and filters tasks from the database
func (t *TaskSelector) FilterTasks() {
	var (
		err               error
		filteredTaskDatas []models.TaskData
	)
	log := logger.GetFuncLogger(t.log, "FilterTasks")
	// Get the filter text
	filterText := t.getFilterText()
	// Search (filter) tasks
	if filterText == "" {
		filteredTaskDatas, err = models.NewTask().LoadAll(false)
	} else {
		filteredTaskDatas, err = models.NewTask().Search(filterText)
	}
	if err != nil {
		log.Err(err).
			Str("filter", filterText).
			Msg("unable to filter tasks")
		t.commandChan <- rxgo.Of(TaskSelectorErrorEvent{Err: err})
		return
	}
	log.Debug().
		Str("filter", filterText).
		Int("count", len(filteredTaskDatas)).
		Msg("task filter results")
	// Update list binding with results of search
	taskList := models.TaskDatas(filteredTaskDatas).AsTaskList()
	slices.Reverse(taskList)
	err = t.tasksListBinding.Set(taskList.ToSliceIntf())
	if err != nil {
		log.Err(err).
			Msg("unable to set tasks list binding")
	}
}

// getFilterText returns the text to be used as a filter for tasks
func (t *TaskSelector) getFilterText() string {
	log := logger.GetFuncLogger(t.log, "getFilterText")
	filterText, err := t.filterBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get filter text from binding")
		return ""
	}
	return filterText
}

// CreateRenderer returns a new WidgetRenderer for this widget.
func (t *TaskSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
