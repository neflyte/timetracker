package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

var _ fyne.Widget = (*Task)(nil)

// Task is the implementation of the Task widget. This is essentially TasklistItem v2.
type Task struct {
	widget.BaseWidget
	log                zerolog.Logger
	task               models.Task
	taskID             uint
	container          *fyne.Container
	synopsis           *canvas.Text
	description        *canvas.Text
	synopsisBinding    binding.String
	descriptionBinding binding.String
}

// NewTask returns a pointer to a newly initialized Task
func NewTask() *Task {
	return NewTaskWithData(nil)
}

// NewTaskWithData returns a pointer to a new Task object initialized with the supplied taskData
func NewTaskWithData(taskData models.Task) *Task {
	t := &Task{
		log:                logger.GetStructLogger("Task"),
		synopsisBinding:    binding.NewString(),
		descriptionBinding: binding.NewString(),
	}
	t.ExtendBaseWidget(t)
	t.SetTask(taskData)
	t.initUI()
	t.synopsisBinding.AddListener(binding.NewDataListener(t.synopsisChanged))
	t.descriptionBinding.AddListener(binding.NewDataListener(t.descriptionChanged))
	return t
}

func (t *Task) initUI() {
	t.synopsis = canvas.NewText("", theme.ForegroundColor())
	t.description = canvas.NewText("", theme.ForegroundColor())
	t.description.TextSize = 10
	t.container = container.NewVBox(t.synopsis, t.description)
}

func (t *Task) synopsisChanged() {
	log := logger.GetFuncLogger(t.log, "synopsisChanged")
	synopsis, err := t.synopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting synopsis from binding")
		return
	}
	t.synopsis.Text = synopsis
	t.synopsis.Refresh()
}

func (t *Task) descriptionChanged() {
	log := logger.GetFuncLogger(t.log, "descriptionChanged")
	description, err := t.descriptionBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting description from binding")
		return
	}
	t.description.Text = description
	t.description.Refresh()
}

// Task returns the task currently represented by the widget
func (t *Task) Task() models.Task {
	return t.task
}

// SetTask sets the task to be represented by the widget
func (t *Task) SetTask(taskData models.Task) {
	log := logger.GetFuncLogger(t.log, "SetTask")
	t.taskID = 0
	t.task = taskData
	if taskData == nil {
		log.Debug().
			Msg("taskData is nil")
		return
	}
	if taskData.Data() == nil {
		log.Debug().
			Msg("taskData.Data() is nil")
		return
	}
	err := t.synopsisBinding.Set(taskData.Data().Synopsis)
	if err != nil {
		log.Err(err).
			Msg("unable to set synopsis binding")
	}
	err = t.descriptionBinding.Set(taskData.Data().Description)
	if err != nil {
		log.Err(err).
			Msg("unable to set description binding")
	}
	t.taskID = taskData.Data().ID
}

// CreateRenderer returns a new WidgetRenderer for this widget.
func (t *Task) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
