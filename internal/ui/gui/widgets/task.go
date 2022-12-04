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

const (
	taskDefaultSynopsis = "(?)"
	taskDefaultDescription
)

var _ fyne.Widget = (*Task)(nil)

// Task is the implementation of the Task widget. This is essentially TasklistItem v2.
type Task struct {
	widget.BaseWidget
	log                zerolog.Logger
	task               models.Task
	container          *fyne.Container
	synopsis           *widget.Label
	description        *widget.Label
	synopsisBinding    binding.String
	descriptionBinding binding.String
}

func NewTask() *Task {
	return NewTaskWithData(nil)
}

func NewTaskWithData(taskData models.Task) *Task {
	t := &Task{
		log:                logger.GetStructLogger("Task"),
		synopsisBinding:    binding.NewString(),
		descriptionBinding: binding.NewString(),
	}
	t.ExtendBaseWidget(t)
	t.SetTask(taskData)
	t.initUI()
	return t
}

func (t *Task) initUI() {
	t.synopsis = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	t.synopsis.Bind(t.synopsisBinding)
	t.description = widget.NewLabelWithData(t.descriptionBinding)
	t.container = container.NewVBox(t.synopsis, t.description)
}

func (t *Task) Task() models.Task {
	return t.task
}

func (t *Task) SetTask(taskData models.Task) {
	log := t.log.With().Str("func", "SetTask").Logger()
	t.task = taskData
	synopsysText := taskDefaultSynopsis
	descriptionText := taskDefaultDescription
	if taskData != nil {
		synopsysText = taskData.Data().Synopsis
		descriptionText = taskData.Data().Description
	}
	err := t.synopsisBinding.Set(synopsysText)
	if err != nil {
		log.Err(err).Msg("unable to set synopsis binding")
	}
	err = t.descriptionBinding.Set(descriptionText)
	if err != nil {
		log.Err(err).Msg("unable to set description binding")
	}
}

func (t *Task) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
