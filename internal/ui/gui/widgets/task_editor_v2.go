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
// taskEditorV2CommandChanSize = 2
)

var _ fyne.Widget = (*TaskEditorV2)(nil)

type TaskEditorV2 struct {
	widget.BaseWidget
	log                    zerolog.Logger
	container              *fyne.Container
	taskSynopsisBinding    binding.String
	taskDescriptionBinding binding.String
	synopsisLabel          *widget.Label
	synopsisEntry          *widget.Entry
	descriptionLabel       *widget.Label
	descriptionEntry       *widget.Entry
}

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
	t.container = container.NewVBox(t.synopsisLabel, t.synopsisEntry, t.descriptionLabel, t.descriptionEntry)
}

func (t *TaskEditorV2) Task() models.Task {
	log := t.log.With().Str("func", "Task").Logger()
	synopsis, err := t.taskSynopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting synopsis from binding")
		return nil
	}
	description, err := t.taskDescriptionBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting description from binding")
		return nil
	}
	task := models.NewTask()
	task.Data().Synopsis = synopsis
	task.Data().Description = description
	return task
}

func (t *TaskEditorV2) SetTask(task models.Task) {
	log := t.log.With().Str("func", "SetTask").Logger()
	if task == nil || task.Data() == nil {
		return
	}
	err := t.taskSynopsisBinding.Set(task.Data().Synopsis)
	if err != nil {
		log.Err(err).Msg("error setting synopsis")
		return
	}
	err = t.taskDescriptionBinding.Set(task.Data().Description)
	if err != nil {
		log.Err(err).Msg("error setting description")
		return
	}
}
