package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/models"
)

type TaskEditor struct {
	widget.BaseWidget
	taskSynopsis           string
	taskSynopsisBinding    binding.ExternalString
	taskDescription        string
	taskDescriptionBinding binding.ExternalString
	taskData               *models.TaskData
	isDirtyBinding         binding.ExternalBool
	isDirty                bool
	saveCB                 func()
	closeCB                func()
}

func NewTaskEditor(saveCallback func(), closeCallback func()) *TaskEditor {
	te := new(TaskEditor)
	te.ExtendBaseWidget(te)
	te.saveCB = saveCallback
	te.closeCB = closeCallback
	te.isDirtyBinding = binding.BindBool(&te.isDirty)
	te.taskSynopsisBinding = binding.BindString(&te.taskSynopsis)
	te.taskSynopsisBinding.AddListener(binding.NewDataListener(func() {
		newSyn, err := te.taskSynopsisBinding.Get()
		if err == nil {
			if te.taskData != nil {
				te.isDirty = te.taskData.Synopsis != newSyn
			}
		}
	}))
	te.taskDescriptionBinding = binding.BindString(&te.taskDescription)
	te.taskDescriptionBinding.AddListener(binding.NewDataListener(func() {
		newDesc, err := te.taskDescriptionBinding.Get()
		if err == nil {
			if te.taskData != nil {
				te.isDirty = te.taskData.Description != newDesc
			}
		}
	}))
	te.taskData = new(models.TaskData)
	return te
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
	return nil
}

func (te *TaskEditor) GetDirtyTask() *models.TaskData {
	if !te.IsDirty() {
		return te.GetTask()
	}
	dirty := new(models.TaskData)
	dirty.ID = te.GetTask().ID
	dirty.Synopsis = te.taskSynopsis
	dirty.Description = te.taskDescription
	dirty.CreatedAt = te.GetTask().CreatedAt
	dirty.UpdatedAt = te.GetTask().UpdatedAt
	return dirty
}

func (te *TaskEditor) IsDirty() bool {
	return te.isDirty
}

func (te *TaskEditor) CreateRenderer() fyne.WidgetRenderer {
	te.ExtendBaseWidget(te)
	r := &taskEditorRenderer{
		taskData:         te.taskData,
		synopsisLabel:    widget.NewLabel("Synopsis:"),
		synopsisEntry:    widget.NewEntryWithData(te.taskSynopsisBinding),
		descriptionLabel: widget.NewLabel("Description:"),
		descriptionEntry: widget.NewEntryWithData(te.taskDescriptionBinding),
		saveCallback:     te.saveCB,
		closeCallback:    te.closeCB,
	}
	r.synopsisEntry.SetPlaceHolder("enter the task synopsis here")
	// TODO: write a function that updates the enabled status of buttons when an entry changes
	r.synopsisEntry.OnChanged = func(changed string) {
		if r.taskData != nil {
			isDirty := changed != r.taskData.Synopsis
			if isDirty {
				r.saveButton.Enable()
			} else {
				r.saveButton.Disable()
			}
		}
	}
	r.descriptionEntry.SetPlaceHolder("enter the task description here")
	r.descriptionEntry.MultiLine = true
	r.fieldContainer = container.NewVBox(
		r.synopsisLabel, r.synopsisEntry,
		r.descriptionLabel, r.descriptionEntry,
	)
	r.saveButton = widget.NewButtonWithIcon("SAVE", theme.ConfirmIcon(), r.saveCallback)
	r.closeButton = widget.NewButtonWithIcon("CLOSE", theme.CancelIcon(), r.closeCallback)
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
	canvasObjects    []fyne.CanvasObject
	layout           fyne.Layout
	synopsisLabel    *widget.Label
	synopsisEntry    *widget.Entry
	descriptionLabel *widget.Label
	descriptionEntry *widget.Entry
	saveButton       *widget.Button
	closeButton      *widget.Button
	fieldContainer   *fyne.Container
	buttonContainer  *fyne.Container
	saveCallback     func()
	closeCallback    func()
	taskData         *models.TaskData
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
	r.Layout(r.MinSize())
}
