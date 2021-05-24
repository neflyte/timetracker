package widgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

type TasklistItem struct {
	widget.BaseWidget
	log                zerolog.Logger
	synopsis           string
	description        string
	synopsisBinding    binding.ExternalString
	descriptionBinding binding.ExternalString
	taskData           *models.TaskData
}

func NewTasklistItem() *TasklistItem {
	item := new(TasklistItem)
	item.ExtendBaseWidget(item)
	item.log = logger.GetStructLogger("TasklistItem")
	item.synopsisBinding = binding.BindString(&item.synopsis)
	item.descriptionBinding = binding.BindString(&item.description)
	item.taskData = new(models.TaskData)
	return item
}

func (i *TasklistItem) GetTask() *models.TaskData {
	return i.taskData
}

func (i *TasklistItem) SetTask(newTask *models.TaskData) error {
	if newTask == nil {
		return errors.New("nil values are not accepted")
	}
	i.taskData = newTask
	err := i.synopsisBinding.Set(newTask.Synopsis)
	if err != nil {
		return err
	}
	err = i.descriptionBinding.Set(newTask.Description)
	if err != nil {
		return err
	}
	return nil
}

func (i *TasklistItem) CreateRenderer() fyne.WidgetRenderer {
	i.ExtendBaseWidget(i)
	r := &tasklistItemRenderer{
		tasklistItem:     i,
		layout:           layout.NewVBoxLayout(),
		synopsisLabel:    widget.NewLabelWithData(i.synopsisBinding),
		descriptionLabel: widget.NewLabelWithData(i.descriptionBinding),
	}
	r.synopsisLabel.TextStyle = fyne.TextStyle{Bold: true, Italic: false, Monospace: false}
	r.canvasObjects = []fyne.CanvasObject{r.synopsisLabel, r.descriptionLabel}
	return r
}

type tasklistItemRenderer struct {
	tasklistItem     *TasklistItem
	canvasObjects    []fyne.CanvasObject
	layout           fyne.Layout
	synopsisLabel    *widget.Label
	descriptionLabel *widget.Label
}

func (r *tasklistItemRenderer) Destroy() {}

func (r *tasklistItemRenderer) Objects() []fyne.CanvasObject {
	return r.canvasObjects
}

func (r *tasklistItemRenderer) Layout(size fyne.Size) {
	r.layout.Layout(r.Objects(), size)
}

func (r *tasklistItemRenderer) MinSize() fyne.Size {
	return r.layout.MinSize(r.Objects())
}

func (r *tasklistItemRenderer) Refresh() {
	// TODO: update enabled/disabled states here
	r.Layout(r.MinSize())
}
