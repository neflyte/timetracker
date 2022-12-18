package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	"reflect"
)

const (
	selectedTaskNone            = widget.ListItemID(-1)
	taskSelectorCommandChanSize = 2
)

// TaskSelectorSelectedEvent contains the task that is sent to the command channel when a selection happens
type TaskSelectorSelectedEvent struct {
	// SelectedTask is the selected task item; nil if there is no selection
	SelectedTask models.Task
}

var _ fyne.Widget = (*TaskSelector)(nil)

// TaskSelector is the implementation of the task selector widget
type TaskSelector struct {
	widget.BaseWidget
	log              zerolog.Logger
	container        *fyne.Container
	filterHBox       *fyne.Container
	filterEntry      *widget.Entry
	sortButton       *widget.Button
	tasksList        *widget.List
	tasksListBinding binding.UntypedList
	selectedTask     widget.ListItemID
	commandChan      chan rxgo.Item
}

// NewTaskSelector returns a pointer to a new, initialized instance of TaskSelector
func NewTaskSelector() *TaskSelector {
	ts := &TaskSelector{
		log:              logger.GetStructLogger("TaskSelector"),
		tasksListBinding: binding.NewUntypedList(),
		selectedTask:     selectedTaskNone,
		commandChan:      make(chan rxgo.Item, taskSelectorCommandChanSize),
	}
	ts.ExtendBaseWidget(ts)
	ts.initUI()
	return ts
}

// initUI initializes UI widgets
func (t *TaskSelector) initUI() {
	t.filterEntry = widget.NewEntry()
	t.filterEntry.SetPlaceHolder("Filter tasks")
	t.sortButton = widget.NewButton("Sort", t.showSortMenu)
	t.filterHBox = container.NewBorder(nil, nil, nil, t.sortButton, t.filterEntry)
	t.tasksList = widget.NewListWithData(t.tasksListBinding, t.createTaskWidget, t.updateTaskWidget)
	t.tasksList.OnSelected = t.taskWasSelected
	// t.tasksList.OnUnselected = t.taskWasUnselected
	t.container = container.NewBorder(t.filterHBox, nil, nil, nil, t.tasksList)
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

func (t *TaskSelector) taskWasSelected(id widget.ListItemID) {
	if t.selectedTask != selectedTaskNone {
		t.tasksList.Unselect(t.selectedTask)
		t.commandChan <- rxgo.Of(TaskSelectorSelectedEvent{nil})
	}
	t.selectedTask = id
	t.commandChan <- rxgo.Of(TaskSelectorSelectedEvent{t.Selected()})
}

// Observable returns an rxgo Observable for the widget's command channel
func (t *TaskSelector) Observable() rxgo.Observable {
	return rxgo.FromEventSource(t.commandChan)
}

// List returns the list of Task objects used in the selector
func (t *TaskSelector) List() models.TaskList {
	log := logger.GetFuncLogger(t.log, "List")
	taskListIntf, err := t.tasksListBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("unable to get list from binding")
		return make([]models.Task, 0)
	}
	return models.TaskListFromSliceIntf(taskListIntf)
}

// SetList sets the list of Task objects used in the selector
func (t *TaskSelector) SetList(tasks models.TaskList) {
	log := logger.GetFuncLogger(t.log, "SetList")
	err := t.tasksListBinding.Set(models.TaskListToSliceIntf(tasks))
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
	taskList := t.List()
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

func (t *TaskSelector) showSortMenu() {
	t.log.Warn().
		Msg("showSortMenu(): IMPLEMENTATION MISSING")
}

// CreateRenderer returns a new WidgetRenderer for this widget.
func (t *TaskSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
