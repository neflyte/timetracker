package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

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
}

// NewTaskSelector returns a pointer to a new instance of TaskSelector
func NewTaskSelector() *TaskSelector {
	ts := &TaskSelector{
		log:              logger.GetStructLogger("TaskSelector"),
		tasksListBinding: binding.NewUntypedList(),
	}
	ts.ExtendBaseWidget(ts)
	ts.initUI()
	return ts
}

func (t *TaskSelector) initUI() {
	t.filterEntry = widget.NewEntry()
	t.filterEntry.SetPlaceHolder("Filter tasks")
	t.sortButton = widget.NewButton("Sort", t.doShowSortMenu)
	t.filterHBox = container.NewHBox(t.filterEntry, t.sortButton)
	t.tasksList = widget.NewListWithData(t.tasksListBinding, t.createTaskWidget, t.updateTaskWidget)
	t.container = container.NewVBox(t.filterHBox, t.tasksList)
}

func (t *TaskSelector) createTaskWidget() fyne.CanvasObject {
	return NewTask()
}

func (t *TaskSelector) updateTaskWidget(item binding.DataItem, canvasObject fyne.CanvasObject) {
	log := t.log.With().Str("func", "updateTaskWidget").Logger()
	// get task
	listBinding, ok := item.(binding.Untyped)
	if !ok {
		log.Error().Msgf("item is %T but should be binding.Untyped", item)
		return
	}
	taskIntf, err := listBinding.Get()
	if err != nil {
		log.Err(err).Msg("unable to get task interface from listBinding")
		return
	}
	task, ok := taskIntf.(models.Task)
	if !ok {
		log.Error().Msgf("taskIntf is %T but should be models.Task instead", taskIntf)
		return
	}
	// get widget
	taskWidget, ok := canvasObject.(*Task)
	if !ok {
		log.Error().Msgf("canvasObject is %T but should be *Task", canvasObject)
		return
	}
	// update widget
	taskWidget.SetTask(task)
}

func (t *TaskSelector) List() models.TaskList {
	log := t.log.With().Str("func", "List").Logger()
	taskListIntf, err := t.tasksListBinding.Get()
	if err != nil {
		log.Err(err).Msg("unable to get list from binding")
		return make([]models.Task, 0)
	}
	return models.TaskListFromSliceIntf(taskListIntf)
}

func (t *TaskSelector) SetList(tasks models.TaskList) {
	log := t.log.With().Str("func", "SetList").Logger()
	err := t.tasksListBinding.Set(models.TaskListToSliceIntf(tasks))
	if err != nil {
		log.Err(err).Msg("unable to set tasksListBinding")
	}
}

func (t *TaskSelector) doShowSortMenu() {
	t.log.Warn().Msg("doShowSortMenu(): IMPLEMENTATION MISSING")
}

// CreateRenderer returns a new WidgetRenderer for this widget.
func (t *TaskSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
