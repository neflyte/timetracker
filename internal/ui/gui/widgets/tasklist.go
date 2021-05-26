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
	tasklistMinWidth = 350 // tasklistMinWidth is the minimum width of the widget in pixels
)

type Tasklist struct {
	widget.Select
	log         zerolog.Logger
	listBinding binding.StringList
}

func NewTasklist() *Tasklist {
	tl := &Tasklist{
		log:         logger.GetStructLogger("Tasklist"),
		listBinding: binding.NewStringList(),
	}
	tl.ExtendBaseWidget(tl)
	tl.listBinding.AddListener(binding.NewDataListener(tl.listBindingChanged))
	return tl
}

func (t *Tasklist) Refresh() {
	t.refreshTaskList()
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

func (t *Tasklist) MinSize() fyne.Size {
	minsize := t.Select.MinSize()
	if minsize.Width < tasklistMinWidth {
		minsize.Width = tasklistMinWidth
	}
	return minsize
}
