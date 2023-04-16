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

var _ fyne.Widget = (*TaskEditorV2)(nil)

// TaskEditorV2 is the struct implementing the TaskEditorV2 widget.
// Use NewTaskEditorV2() to create a new instance of the widget.
type TaskEditorV2 struct {
	taskSynopsisBinding    binding.String
	taskDescriptionBinding binding.String
	container              *fyne.Container
	synopsisLabel          *widget.Label
	synopsisEntry          *widget.Entry
	descriptionLabel       *widget.Label
	descriptionEntry       *widget.Entry
	log                    zerolog.Logger
	widget.BaseWidget
	taskID uint
}

// NewTaskEditorV2 returns a pointer to a newly initialized instance of the TaskEditorV2 widget
func NewTaskEditorV2() *TaskEditorV2 {
	te := &TaskEditorV2{
		log:                    logger.GetStructLogger("TaskEditorV2"),
		taskSynopsisBinding:    binding.NewString(),
		taskDescriptionBinding: binding.NewString(),
	}
	te.ExtendBaseWidget(te)
	te.initUI()
	return te
}

// CreateRenderer returns a new WidgetRenderer for this widget.
func (t *TaskEditorV2) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}

func (t *TaskEditorV2) initUI() {
	t.synopsisLabel = widget.NewLabel("Synopsis:") // i18n
	t.synopsisEntry = widget.NewEntryWithData(t.taskSynopsisBinding)
	t.synopsisEntry.SetPlaceHolder("enter the task synopsis here") // i18n
	t.synopsisEntry.Validator = nil
	t.descriptionLabel = widget.NewLabel("Description:") // i18n
	t.descriptionEntry = widget.NewEntryWithData(t.taskDescriptionBinding)
	t.descriptionEntry.SetPlaceHolder("enter the task description here") // i18n
	t.descriptionEntry.Validator = nil
	t.descriptionEntry.MultiLine = true
	t.descriptionEntry.Wrapping = fyne.TextWrapWord
	t.container = container.NewVBox(
		container.NewBorder(nil, nil, t.synopsisLabel, nil, t.synopsisEntry),
		t.descriptionLabel,
		t.descriptionEntry,
	)
}

// Reset resets the editor widget to its default state
func (t *TaskEditorV2) Reset() {
	log := logger.GetFuncLogger(t.log, "Reset")
	t.taskID = 0
	err := t.taskSynopsisBinding.Set("")
	if err != nil {
		log.Err(err).
			Msg("error resetting synopsis binding")
	}
	err = t.taskDescriptionBinding.Set("")
	if err != nil {
		log.Err(err).
			Msg("error resetting description binding")
	}
}

// Task returns the task currently being edited
func (t *TaskEditorV2) Task() models.Task {
	log := logger.GetFuncLogger(t.log, "Task")
	synopsis, err := t.taskSynopsisBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("error getting synopsis from binding")
		return nil
	}
	description, err := t.taskDescriptionBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("error getting description from binding")
		return nil
	}
	task := models.NewTask()
	task.Data().Synopsis = synopsis
	task.Data().Description = description
	task.Data().ID = t.taskID
	log.Debug().
		Str("task", task.String()).
		Msg("returning task")
	return task
}

// SetTask sets the task to be edited
func (t *TaskEditorV2) SetTask(task models.Task) {
	log := logger.GetFuncLogger(t.log, "SetTask")
	if task == nil || task.Data() == nil {
		return
	}
	// Set the known task ID to zero in case updating the bindings fails.
	// We don't want to unintentionally update a task with bad data.
	t.taskID = 0
	// Update the synopsis binding
	err := t.taskSynopsisBinding.Set(task.Data().Synopsis)
	if err != nil {
		log.Err(err).
			Msg("error setting synopsis")
		return
	}
	// Update the description binding
	err = t.taskDescriptionBinding.Set(task.Data().Description)
	if err != nil {
		log.Err(err).
			Msg("error setting description")
		return
	}
	// Save the task's ID now that we've updated the bindings
	t.taskID = task.Data().ID
	// Log what we set
	log.Debug().
		Str("task", task.String()).
		Msg("task set successfully")
}
