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
	taskSynopsisBinding    binding.ExternalString
	taskDescription        string
	taskDescriptionBinding binding.ExternalString
	taskData               *models.TaskData
	taskDataChannel        chan rxgo.Item

	// TODO: Remove the split synopsis/descripting bindings and replace with a taskData binding instead

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
	te.taskSynopsisBinding = binding.BindString(&te.taskSynopsis)
	te.taskDescriptionBinding = binding.BindString(&te.taskDescription)
	te.taskData = new(models.TaskData)
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
	return te.taskData
}

func (te *TaskEditor) SetTask(task *models.TaskData) error {
	te.taskData = task
	if te.taskData != nil {
		err := te.taskSynopsisBinding.Set(te.taskData.Synopsis)
		if err != nil {
			return err
		}
		err = te.taskDescriptionBinding.Set(te.taskData.Description)
		if err != nil {
			return err
		}
	}
	te.taskDataChannel <- rxgo.Of(task)
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
	log := te.log.With().Str("func", "IsDirty").Logger()
	if te.taskData == nil {
		return false
	}
	editSynopsis, err := te.taskSynopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting synopsis from binding")
		return false
	}
	if te.taskData.Synopsis != editSynopsis {
		return true
	}
	editDescription, err := te.taskDescriptionBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting description from binding")
		return false
	}
	if te.taskData.Description != editDescription {
		return true
	}
	return false
}

func (te *TaskEditor) CreateRenderer() fyne.WidgetRenderer {
	te.ExtendBaseWidget(te)
	r := &taskEditorRenderer{
		log:                  logger.GetStructLogger("TaskEditor.taskEditorRenderer"),
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
			newTaskData, ok := item.(*models.TaskData)
			if ok {
				r.log.Debug().Msgf("setting taskData from observable; taskData=%s", newTaskData.String())
				r.taskData = newTaskData
			}
		},
		func(err error) {
			r.log.Err(err).Msg("error from taskData observable")
		},
		func() {
			r.log.Debug().Msg("taskData observable is finished")
		},
	)
	r.synopsisEntry.SetPlaceHolder("enter the task synopsis here")
	r.synopsisEntry.OnChanged = func(_ string) { r.updateButtonStates() }
	r.descriptionEntry.SetPlaceHolder("enter the task description here")
	r.descriptionEntry.MultiLine = true
	r.descriptionEntry.OnChanged = func(_ string) { r.updateButtonStates() }
	r.fieldContainer = container.NewVBox(
		r.synopsisLabel, r.synopsisEntry,
		r.descriptionLabel, r.descriptionEntry,
	)
	r.saveButton = widget.NewButtonWithIcon("SAVE", theme.ConfirmIcon(), r.doSaveTask)
	r.closeButton = widget.NewButtonWithIcon("CLOSE", theme.CancelIcon(), r.doCancelEdit)
	r.buttonContainer = container.NewCenter(container.NewHBox(r.closeButton, r.saveButton))
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
	taskData             *models.TaskData
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
	r.taskSavedChannel <- rxgo.Of(r.taskData)
}

func (r *taskEditorRenderer) doCancelEdit() {
	r.editCancelledChannel <- rxgo.Of(true)
}

func (r *taskEditorRenderer) updateButtonStates() {
	if r.isDirty() && !r.taskEditor.Disabled() {
		r.saveButton.Enable()
	} else {
		r.saveButton.Disable()
	}
	if r.taskEditor.Disabled() {
		r.closeButton.Disable()
	} else {
		r.closeButton.Enable()
	}
}

func (r *taskEditorRenderer) isDirty() bool {
	if r.taskData != nil {
		if r.taskData.Synopsis != r.synopsisEntry.Text {
			return true
		}
		if r.taskData.Description != r.descriptionEntry.Text {
			return true
		}
	}
	return false
}
