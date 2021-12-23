package widgets

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/elliotchance/pie/pie"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/rs/zerolog"
)

const (
	// tasklistMinWidth is the minimum width of the widget in pixels
	tasklistMinWidth              = 500
	tasklistDescriptionTrimLength = 48
	tasklistLastStartedTasks      = 5
	tasklistSeparator             = "---"
	tasklistStringFormat          = "%s: %s"
)

// Tasklist is an extended widget.Select object that adds data binding for displaying a list of tasks
type Tasklist struct {
	widget.Select
	log              zerolog.Logger
	listBinding      binding.StringList
	selectionBinding binding.String
	tasks            []models.TaskData
}

// NewTasklist creates a new instance of a Tasklist widget
func NewTasklist() *Tasklist {
	tl := &Tasklist{
		log:              logger.GetStructLogger("Tasklist"),
		listBinding:      binding.NewStringList(),
		selectionBinding: binding.NewString(),
		tasks:            make([]models.TaskData, 0),
	}
	tl.ExtendBaseWidget(tl)
	tl.listBinding.AddListener(binding.NewDataListener(tl.listBindingChanged))
	tl.OnChanged = tl.selectionChanged
	return tl
}

// Refresh redraws the task list
func (t *Tasklist) Refresh() {
	t.refreshTaskList()
	t.Select.Refresh()
}

// SelectionBinding returns the Select widget's data binding for the current selection
func (t *Tasklist) SelectionBinding() binding.String {
	return t.selectionBinding
}

func (t *Tasklist) listBindingChanged() {
	log := logger.GetFuncLogger(t.log, "listBindingChanged")
	changed, err := t.listBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting list from binding")
		return
	}
	// If the currently selected option doesn't exist in the new list of options then we need to unset it
	if t.Select.Selected != "" && !pie.Strings(changed).Contains(t.Select.Selected) {
		t.Select.ClearSelected()
	}
	t.Select.Options = changed
}

func (t *Tasklist) selectionChanged(selection string) {
	log := logger.GetFuncLogger(t.log, "selectionChanged")
	err := t.selectionBinding.Set(selection)
	if err != nil {
		log.Err(err).Msg("error setting selectionBinding")
	}
}

func (t *Tasklist) refreshTaskList() {
	log := logger.GetFuncLogger(t.log, "refreshTaskList")
	log.Debug().Msg("refreshing task list data")
	td := models.NewTask()
	tasks, err := td.LoadAll(false)
	if err != nil {
		log.Err(err).Msg("unable to load task list")
		return
	}
	log.Trace().Msgf("len(tasks)=%d", len(tasks))
	allTasks := make([]models.TaskData, len(tasks))
	for idx, task := range tasks {
		allTasks[idx] = *task.Clone().Data()
	}
	// Get last 5 started tasks
	tsd := models.NewTimesheet()
	lastStartedTasks, err := tsd.LastStartedTasks(tasklistLastStartedTasks)
	if err != nil {
		log.Err(err).Msg("error loading last started tasks")
		return
	}
	log.Trace().Msgf("len(lastStartedTasks)=%d", len(lastStartedTasks))
	// Prepend the last started tasks to the front of the list
	t.tasks = make([]models.TaskData, 0)
	for _, task := range lastStartedTasks {
		t.tasks = append(t.tasks, *task.Clone().Data())
	}
	t.tasks = append(t.tasks, allTasks...)
	taskStrings := make([]string, len(t.tasks)+1)
	handledLastStarted := false
	for idx, task := range t.tasks {
		if idx+1 == len(lastStartedTasks) {
			taskStrings[idx] = fmt.Sprintf(
				tasklistStringFormat,
				task.Synopsis,
				utils.TrimWithEllipsis(task.Description, tasklistDescriptionTrimLength),
			)
			taskStrings[idx+1] = tasklistSeparator
			handledLastStarted = true
			continue
		}
		index := idx
		if handledLastStarted {
			index++
		}
		taskStrings[index] = fmt.Sprintf(
			tasklistStringFormat,
			task.Synopsis,
			utils.TrimWithEllipsis(task.Description, tasklistDescriptionTrimLength),
		)
	}
	log.Trace().Msgf("taskStrings=%#v", taskStrings)
	err = t.listBinding.Set(taskStrings)
	if err != nil {
		log.Err(err).Msg("unable to set list binding")
		return
	}
}

// MinSize returns the minimum size of this widget
func (t *Tasklist) MinSize() fyne.Size {
	minsize := t.Select.MinSize()
	if minsize.Width < tasklistMinWidth {
		minsize.Width = tasklistMinWidth
	}
	return minsize
}
