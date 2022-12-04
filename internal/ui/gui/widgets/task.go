package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/models"
)

const (
	taskDefaultSynopsis = "(?)"
	taskDefaultDescription
)

// taskImpl is the struct implementation of the TaskIntf interface
type taskImpl struct {
	widget.BaseWidget
	task        models.Task
	container   *fyne.Container
	synopsis    *widget.Label
	description *widget.Label
}

func NewTask() *taskImpl {
	return NewTaskWithData(nil)
}

func NewTaskWithData(taskData models.Task) *taskImpl {
	t := &taskImpl{
		task: taskData,
	}
	t.ExtendBaseWidget(t)
	t.initUI()
	return t
}

func (t *taskImpl) initUI() {
	synopsysText := ""
	descriptionText := ""
	if t.task != nil {
		synopsysText = t.task.Data().Synopsis
		descriptionText = t.task.Data().Description
	}
	t.synopsis = widget.NewLabelWithStyle(synopsysText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	t.description = widget.NewLabel(descriptionText)
	t.container = container.NewVBox(t.synopsis, t.description)
}

func (t *taskImpl) Task() models.Task {
	return t.task
}

func (t *taskImpl) SetTask(taskData models.Task) {
	didChange := false
	t.task = taskData
	if taskData == nil {
		return
	}
	if t.synopsis != nil {
		t.synopsis.SetText(taskData.Data().Synopsis)
		didChange = true
	}
	if t.description != nil {
		t.description.SetText(taskData.Data().Description)
		didChange = true
	}
	if didChange {
		t.Refresh()
	}
}

func (t *taskImpl) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
