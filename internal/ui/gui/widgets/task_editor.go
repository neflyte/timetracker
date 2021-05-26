package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
)

const (
	TaskEditorEventChannelBufferSize = 2
	TaskEditorTaskSavedEventKey      = "task-saved"
	TaskEditorEditCancelledEventKey  = "edit-cancelled"
)

type TaskEditor struct {
	widget.DisableableWidget
	taskSynopsis           string
	taskDescription        string
	taskSynopsisBinding    binding.String
	taskDescriptionBinding binding.String
	taskData               models.TaskData

	taskDataChannel      chan rxgo.Item
	taskSavedChannel     chan rxgo.Item
	editCancelledChannel chan rxgo.Item

	taskSavedObservable     rxgo.Observable
	editCancelledObservable rxgo.Observable

	observerablesMap map[string]rxgo.Observable

	log zerolog.Logger
}

func NewTaskEditor() *TaskEditor {
	te := new(TaskEditor)
	te.ExtendBaseWidget(te)
	te.log = logger.GetStructLogger("TaskEditor")
	te.taskSavedChannel = make(chan rxgo.Item, TaskEditorEventChannelBufferSize)
	te.editCancelledChannel = make(chan rxgo.Item, TaskEditorEventChannelBufferSize)
	te.taskSavedObservable = rxgo.FromEventSource(te.taskSavedChannel)
	te.editCancelledObservable = rxgo.FromEventSource(te.editCancelledChannel)
	te.taskSynopsisBinding = binding.NewString()
	te.taskSynopsisBinding.AddListener(binding.NewDataListener(func() {
		editSynopsis, err := te.taskSynopsisBinding.Get()
		if err != nil {
			te.log.Err(err).Msg("error getting synopsis from binding")
		}
		te.log.Trace().Msgf("setting te.taskSynopsis=%s from binding datalistener", editSynopsis)
		te.taskSynopsis = editSynopsis
	}))
	te.taskDescriptionBinding = binding.NewString()
	te.taskDescriptionBinding.AddListener(binding.NewDataListener(func() {
		editDescription, err := te.taskDescriptionBinding.Get()
		if err != nil {
			te.log.Err(err).Msg("error getting description from binding")
		}
		te.log.Trace().Msgf("setting te.taskDescription=%s from binding datalistener", editDescription)
		te.taskDescription = editDescription
	}))
	te.taskData = models.TaskData{}
	te.taskDataChannel = make(chan rxgo.Item, TaskEditorEventChannelBufferSize)
	te.observerablesMap = map[string]rxgo.Observable{
		TaskEditorTaskSavedEventKey:     te.taskSavedObservable,
		TaskEditorEditCancelledEventKey: te.editCancelledObservable,
	}
	return te
}

func (te *TaskEditor) Observables() map[string]rxgo.Observable {
	return te.observerablesMap
}

func (te *TaskEditor) GetTask() *models.TaskData {
	return &te.taskData
}

func (te *TaskEditor) SetTask(task *models.TaskData) error {
	log := logger.GetFuncLogger(te.log, "SetTask")
	if task != nil {
		log.Trace().Msgf("sending task %s to taskDataChannel", task.String())
		te.taskDataChannel <- rxgo.Of(task)
		log.Trace().Msgf("setting te.taskData=%s", task.String())
		te.taskData = *task
		log.Trace().Msgf("setting synopsis binding to %s", task.Synopsis)
		err := te.taskSynopsisBinding.Set(task.Synopsis)
		if err != nil {
			return err
		}
		log.Trace().Msgf("setting description binding to %s", task.Description)
		err = te.taskDescriptionBinding.Set(task.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func (te *TaskEditor) GetDirtyTask() *models.TaskData {
	if !te.IsDirty() {
		return te.GetTask()
	}
	dirty := te.taskData.Clone()
	dirty.Synopsis = te.taskSynopsis
	dirty.Description = te.taskDescription
	return dirty
}

func (te *TaskEditor) IsDirty() bool {
	log := logger.GetFuncLogger(te.log, "IsDirty")
	log.Trace().Msgf(
		"te.taskData.Synopsis=%s, te.taskSynopsis=%s",
		te.taskData.Synopsis,
		te.taskSynopsis,
	)
	if te.taskData.Synopsis != te.taskSynopsis {
		log.Debug().Msg("synopsis is different; returning true")
		return true
	}
	log.Trace().Msgf(
		"te.taskData.Description=%s, te.taskDescription=%s",
		te.taskData.Description,
		te.taskDescription,
	)
	if te.taskData.Description != te.taskDescription {
		log.Debug().Msg("description is different; returning true")
		return true
	}
	log.Debug().Msg("no differences; returning false")
	return false
}

func (te *TaskEditor) CreateRenderer() fyne.WidgetRenderer {
	te.ExtendBaseWidget(te)
	r := &taskEditorRenderer{
		log:                  te.log.With().Str("struct", "taskEditorRenderer").Logger(),
		taskEditor:           te,
		taskData:             te.taskData,
		taskDataObservable:   rxgo.FromEventSource(te.taskDataChannel),
		synopsisLabel:        widget.NewLabel("Synopsis:"),
		synopsisEntry:        widget.NewEntryWithData(te.taskSynopsisBinding),
		descriptionLabel:     widget.NewLabel("Description:"),
		descriptionEntry:     widget.NewEntryWithData(te.taskDescriptionBinding),
		taskSavedChannel:     te.taskSavedChannel,
		editCancelledChannel: te.editCancelledChannel,
	}
	r.taskDataObservable.ForEach(
		func(item interface{}) {
			r.log.Trace().Msg("taskData observable fired")
			newTaskData, ok := item.(models.TaskData)
			if ok {
				r.log.Trace().Msgf("setting taskData from observable (cloned); taskData=%s", newTaskData.String())
				cloned := newTaskData.Clone()
				r.taskData = *cloned
			}
		},
		func(err error) {
			r.log.Err(err).Msg("error from taskData observable")
		},
		func() {
			r.log.Trace().Msg("taskData observable is finished")
		},
	)
	te.taskSynopsisBinding.AddListener(binding.NewDataListener(func() {
		r.updateButtonStates()
	}))
	te.taskDescriptionBinding.AddListener(binding.NewDataListener(func() {
		r.updateButtonStates()
	}))
	r.synopsisEntry.SetPlaceHolder("enter the task synopsis here")
	r.descriptionEntry.SetPlaceHolder("enter the task description here")
	r.descriptionEntry.MultiLine = true
	r.descriptionEntry.Wrapping = fyne.TextWrapWord
	r.fieldContainer = container.NewVBox(
		r.synopsisLabel, r.synopsisEntry,
		r.descriptionLabel, r.descriptionEntry,
	)
	r.saveButton = widget.NewButtonWithIcon("SAVE", theme.ConfirmIcon(), r.doSaveTask)
	r.closeButton = widget.NewButtonWithIcon("CANCEL", theme.CancelIcon(), r.doCancelEdit)
	r.buttonContainer = container.NewBorder(nil, nil, nil, container.NewHBox(r.closeButton, r.saveButton))
	r.canvasObjects = []fyne.CanvasObject{
		r.buttonContainer, r.fieldContainer,
	}
	r.layout = layout.NewBorderLayout(
		nil,
		r.buttonContainer,
		nil,
		nil,
	)
	return r
}

type taskEditorRenderer struct {
	log                  zerolog.Logger
	taskEditor           *TaskEditor
	canvasObjects        []fyne.CanvasObject
	layout               fyne.Layout
	synopsisLabel        *widget.Label
	synopsisEntry        *widget.Entry
	descriptionLabel     *widget.Label
	descriptionEntry     *widget.Entry
	saveButton           *widget.Button
	closeButton          *widget.Button
	fieldContainer       *fyne.Container
	buttonContainer      *fyne.Container
	taskData             models.TaskData
	taskDataObservable   rxgo.Observable
	taskSavedChannel     chan rxgo.Item
	editCancelledChannel chan rxgo.Item
}

func (r *taskEditorRenderer) Destroy() {}

func (r *taskEditorRenderer) Layout(size fyne.Size) {
	r.layout.Layout(r.Objects(), size)
}

func (r *taskEditorRenderer) MinSize() fyne.Size {
	return r.layout.MinSize(r.Objects())
}

func (r *taskEditorRenderer) Objects() []fyne.CanvasObject {
	return r.canvasObjects
}

func (r *taskEditorRenderer) Refresh() {
	r.updateButtonStates()
	if r.taskEditor.Disabled() {
		r.synopsisEntry.Disable()
		r.descriptionEntry.Disable()
	} else {
		r.synopsisEntry.Enable()
		r.descriptionEntry.Enable()
	}
	r.Layout(r.MinSize())
}

func (r *taskEditorRenderer) doSaveTask() {
	dirtyTask := r.taskData.Clone()
	dirtyTask.Synopsis = r.synopsisEntry.Text
	dirtyTask.Description = r.descriptionEntry.Text
	r.taskSavedChannel <- rxgo.Of(*dirtyTask)
}

func (r *taskEditorRenderer) doCancelEdit() {
	r.editCancelledChannel <- rxgo.Of(true)
}

func (r *taskEditorRenderer) updateButtonStates() {
	if r.saveButton != nil {
		if r.isDirty() && (r.taskEditor != nil && !r.taskEditor.Disabled()) {
			r.saveButton.Enable()
		} else {
			r.saveButton.Disable()
		}
	}
	if r.closeButton != nil {
		if r.taskEditor.Disabled() {
			r.closeButton.Disable()
		} else {
			r.closeButton.Enable()
		}
	}
}

func (r *taskEditorRenderer) isDirty() bool {
	log := logger.GetFuncLogger(r.log, "isDirty")
	if r.taskData.Synopsis != r.synopsisEntry.Text {
		log.Debug().Msg("synopsis is different; returning true")
		return true
	}
	if r.taskData.Description != r.descriptionEntry.Text {
		log.Debug().Msg("description is different; returning true")
		return true
	}
	log.Debug().Msg("no differences; returning false")
	return false
}
