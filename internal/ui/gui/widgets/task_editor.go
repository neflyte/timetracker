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
	// taskEditorEventChannelBufferSize is the size of the event channels in this widget
	taskEditorEventChannelBufferSize = 2

	// TaskEditorTaskSavedEventKey is the map key to the taskSaved observable
	TaskEditorTaskSavedEventKey = "task-saved"
	// TaskEditorEditCancelledEventKey is the map key to the editCancelled observable
	TaskEditorEditCancelledEventKey = "edit-cancelled"
)

// TaskEditor is the main data structure for the TaskEditor widget
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

	observerablesMap map[string]rxgo.Observable

	log zerolog.Logger
}

// NewTaskEditor creates and initializes a new TaskEditor widget
func NewTaskEditor() *TaskEditor {
	te := new(TaskEditor)
	te.ExtendBaseWidget(te)
	te.log = logger.GetStructLogger("TaskEditor")
	te.taskSavedChannel = make(chan rxgo.Item, taskEditorEventChannelBufferSize)
	te.editCancelledChannel = make(chan rxgo.Item, taskEditorEventChannelBufferSize)
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
	te.taskDataChannel = make(chan rxgo.Item, taskEditorEventChannelBufferSize)
	te.observerablesMap = map[string]rxgo.Observable{
		TaskEditorTaskSavedEventKey:     rxgo.FromEventSource(te.taskSavedChannel),
		TaskEditorEditCancelledEventKey: rxgo.FromEventSource(te.editCancelledChannel),
	}
	return te
}

// Observables returns a map of Observable objects used by this widget
func (te *TaskEditor) Observables() map[string]rxgo.Observable {
	return te.observerablesMap
}

// getTask returns the current models.TaskData struct
func (te *TaskEditor) getTask() *models.TaskData {
	return &te.taskData
}

// SetTask sets the current models.TaskData struct
func (te *TaskEditor) SetTask(task *models.TaskData) error {
	if task != nil {
		te.taskDataChannel <- rxgo.Of(task)
		te.taskData = *task
		err := te.taskSynopsisBinding.Set(task.Synopsis)
		if err != nil {
			return err
		}
		err = te.taskDescriptionBinding.Set(task.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetDirtyTask returns the current models.TaskData structure updated with the data from the widget
func (te *TaskEditor) GetDirtyTask() *models.TaskData {
	if !te.IsDirty() {
		return te.getTask()
	}
	dirty := te.taskData.Clone()
	dirty.Synopsis = te.taskSynopsis
	dirty.Description = te.taskDescription
	return dirty
}

// IsDirty determines if the editor fields have been modified from their original values
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

// CreateRenderer creates and initializes a new fyne.WidgetRenderer
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
		r.taskDataChanged,
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

// Destroy is for internal use.
func (r *taskEditorRenderer) Destroy() {}

// Layout is a hook that is called if the widget needs to be laid out.
// This should never call Refresh.
func (r *taskEditorRenderer) Layout(size fyne.Size) {
	r.layout.Layout(r.Objects(), size)
}

// MinSize returns the minimum size of the widget that is rendered by this renderer.
func (r *taskEditorRenderer) MinSize() fyne.Size {
	return r.layout.MinSize(r.Objects())
}

// Objects returns all objects that should be drawn.
func (r *taskEditorRenderer) Objects() []fyne.CanvasObject {
	return r.canvasObjects
}

// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
// This might trigger a Layout.
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

func (r *taskEditorRenderer) taskDataChanged(item interface{}) {
	log := logger.GetFuncLogger(r.log, "taskDataChanged")
	log.Trace().Msg("taskData observable fired")
	newTaskData, ok := item.(models.TaskData)
	if ok {
		log.Trace().Msgf("setting taskData from observable (cloned); taskData=%s", newTaskData.String())
		cloned := newTaskData.Clone()
		r.taskData = *cloned
	}
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
