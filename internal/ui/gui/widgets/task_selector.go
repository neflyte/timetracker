package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type taskSelectorImpl struct {
	widget.BaseWidget
	container        *fyne.Container
	filterHBox       *fyne.Container
	filterEntry      *widget.Entry
	sortButton       *widget.Button
	tasksList        *widget.List
	tasksListBinding binding.StringList
}

func NewTaskSelector() *taskSelectorImpl {
	ts := &taskSelectorImpl{}
	ts.ExtendBaseWidget(ts)
	ts.initUI()
	return ts
}

func (t *taskSelectorImpl) initUI() {
	t.filterEntry = widget.NewEntry()
	t.sortButton = widget.NewButton("", func() {})
	t.filterHBox = container.NewHBox(t.filterEntry, t.sortButton)
	t.tasksListBinding = binding.NewStringList()
	t.tasksList = widget.NewListWithData(
		t.tasksListBinding,
		func() fyne.CanvasObject {
			return NewTask()
		},
		func(item binding.DataItem, canvasObject fyne.CanvasObject) {
			listBinding, ok := item.(binding.StringList)
			if !ok {
				return
			}
		},
	)
	t.container = container.NewVBox(t.filterHBox)
}

func (t *taskSelectorImpl) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
