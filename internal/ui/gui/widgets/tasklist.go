package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/elliotchance/pie/pie"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

const (
	// tasklistMinWidth is the minimum width of the widget in pixels
	tasklistMinWidth = 350
)

// Tasklist is an extended widget.Select object that adds data binding for displaying a list of tasks
type Tasklist struct {
	widget.Select
	log              zerolog.Logger
	listBinding      binding.StringList
	selectionBinding binding.String
}

// NewTasklist creates a new instance of a Tasklist widget
func NewTasklist() *Tasklist {
	tl := &Tasklist{
		log:              logger.GetStructLogger("Tasklist"),
		listBinding:      binding.NewStringList(),
		selectionBinding: binding.NewString(),
	}
	tl.ExtendBaseWidget(tl)
	tl.listBinding.AddListener(binding.NewDataListener(tl.listBindingChanged))
	tl.OnChanged = tl.selectionChanged
	return tl
}

// Refresh redraws the task list
func (t *Tasklist) Refresh() {
	t.refreshTaskList()
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
	t.Select.Refresh()
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
	td := new(models.TaskData)
	tasks, err := td.LoadAll(false)
	if err != nil {
		log.Err(err).Msg("unable to load task list")
		return
	}
	log.Trace().Msgf("len(tasks)=%d", len(tasks))
	taskStrings := make([]string, len(tasks))
	for idx, task := range tasks {
		taskStrings[idx] = task.String()
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
