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

const (
	synopsisTrimLength    = 8
	descriptionTrimLength = 25
)

type TasklistItem struct {
	widget.BaseWidget
	log                zerolog.Logger
	synopsisBinding    binding.String
	descriptionBinding binding.String
	taskData           *models.TaskData
}

func NewTasklistItem() *TasklistItem {
	item := new(TasklistItem)
	item.ExtendBaseWidget(item)
	item.log = logger.GetStructLogger("TasklistItem")
	item.synopsisBinding = binding.NewString()
	item.descriptionBinding = binding.NewString()
	item.taskData = new(models.TaskData)
	return item
}

func (i *TasklistItem) GetTask() *models.TaskData {
	return i.taskData
}

func (i *TasklistItem) SetTask(newTask *models.TaskData) error {
	// log := logger.GetFuncLogger(i.log, "SetTask")
	if newTask == nil {
		return errors.New("nil values are not accepted")
	}
	i.taskData = newTask
	err := i.synopsisBinding.Set(i.trimWithEllipsis(newTask.Synopsis, synopsisTrimLength))
	if err != nil {
		return err
	}
	err = i.descriptionBinding.Set(i.trimWithEllipsis(newTask.Description, descriptionTrimLength))
	if err != nil {
		return err
	}
	return nil
}

func (i *TasklistItem) trimWithEllipsis(toTrim string, trimLength int) string {
	if len(toTrim) <= trimLength {
		return toTrim
	}
	return toTrim[0:trimLength-2] + `…`
}

func (i *TasklistItem) CreateRenderer() fyne.WidgetRenderer {
	i.ExtendBaseWidget(i)
	r := &tasklistItemRenderer{
		tasklistItem:     i,
		layout:           layout.NewGridLayoutWithColumns(3),
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
	r.Layout(r.MinSize())
}
