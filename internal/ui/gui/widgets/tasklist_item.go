package widgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/rs/zerolog"
)

const (
	synopsisTrimLength    = 8
	descriptionTrimLength = 25
	itemLayoutColumns     = 3
)

// TasklistItem is the data struct for the TasklistItem widget
type TasklistItem struct {
	widget.BaseWidget
	log                zerolog.Logger
	synopsisBinding    binding.String
	descriptionBinding binding.String
	taskData           models.Task
}

// NewTasklistItem creates a new tasklistItem widget
func NewTasklistItem() *TasklistItem {
	item := new(TasklistItem)
	item.ExtendBaseWidget(item)
	item.log = logger.GetStructLogger("TasklistItem")
	item.synopsisBinding = binding.NewString()
	item.descriptionBinding = binding.NewString()
	item.taskData = models.NewTask()
	return item
}

// SetTask sets the current models.TaskData struct
func (i *TasklistItem) SetTask(newTask *models.TaskData) error {
	// log := logger.GetFuncLogger(i.log, "SetTask")
	if newTask == nil {
		return errors.New("nil values are not accepted")
	}
	i.taskData = models.NewTaskWithData(*newTask)
	err := i.synopsisBinding.Set(utils.TrimWithEllipsis(newTask.Synopsis, synopsisTrimLength))
	if err != nil {
		return err
	}
	err = i.descriptionBinding.Set(utils.TrimWithEllipsis(newTask.Description, descriptionTrimLength))
	if err != nil {
		return err
	}
	return nil
}

// CreateRenderer creates and returns a new fyne.WidgetRenderer
func (i *TasklistItem) CreateRenderer() fyne.WidgetRenderer {
	i.ExtendBaseWidget(i)
	r := &tasklistItemRenderer{
		tasklistItem:     i,
		layout:           layout.NewGridLayoutWithColumns(itemLayoutColumns),
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

// Destroy is for internal use.
func (r *tasklistItemRenderer) Destroy() {}

// Objects returns all objects that should be drawn.
func (r *tasklistItemRenderer) Objects() []fyne.CanvasObject {
	return r.canvasObjects
}

// Layout is a hook that is called if the widget needs to be laid out.
// This should never call Refresh.
func (r *tasklistItemRenderer) Layout(size fyne.Size) {
	r.layout.Layout(r.Objects(), size)
}

// MinSize returns the minimum size of the widget that is rendered by this renderer.
func (r *tasklistItemRenderer) MinSize() fyne.Size {
	return r.layout.MinSize(r.Objects())
}

// Refresh is a hook that is called if the widget has updated and needs to be redrawn.
// This might trigger a Layout.
func (r *tasklistItemRenderer) Refresh() {
	r.Layout(r.MinSize())
}
