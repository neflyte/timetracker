package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TaskEditor struct {
	widget.BaseWidget
}

func (te *TaskEditor) CreateRenderer() fyne.WidgetRenderer {
	te.ExtendBaseWidget(te)
	return &taskEditorRenderer{}
}

type taskEditorRenderer struct {
	canvasObjects []fyne.CanvasObject
}

func (r *taskEditorRenderer) Destroy() {}

func (r *taskEditorRenderer) Layout(size fyne.Size) {}

func (r *taskEditorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *taskEditorRenderer) Objects() []fyne.CanvasObject {
	return r.canvasObjects
}

func (r *taskEditorRenderer) Refresh() {}
