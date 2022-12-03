package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TaskSelectorIntf interface {
}

type taskSelectorImpl struct {
	widget.BaseWidget
	container *fyne.Container
}

func NewTaskSelector() TaskSelectorIntf {
	ts := &taskSelectorImpl{}
	ts.ExtendBaseWidget(ts)
	return ts
}

func (t *taskSelectorImpl) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}
