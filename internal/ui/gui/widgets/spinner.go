package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Spinner struct {
	widget.BaseWidget

	upIcon   *TappableIcon
	downIcon *TappableIcon
	upFunc   func()
	downFunc func()
}

func NewSpinner(upfunc func(), downfunc func()) *Spinner {
	s := &Spinner{
		upFunc:   upfunc,
		downFunc: downfunc,
	}
	s.ExtendBaseWidget(s) //nolint:typecheck
	s.upIcon = NewTappableIcon(theme.MenuDropUpIcon(), s.spinnerUp)
	s.downIcon = NewTappableIcon(theme.MenuDropDownIcon(), s.spinnerDown)
	return s
}

func (s *Spinner) spinnerUp() {
	if s.upFunc != nil {
		s.upFunc()
	}
}

func (s *Spinner) spinnerDown() {
	if s.downFunc != nil {
		s.downFunc()
	}
}

type spinnerRenderer struct {
	_       fyne.WidgetRenderer
	layout  fyne.Layout
	objects []fyne.CanvasObject
}

func (s *Spinner) CreateRenderer() fyne.WidgetRenderer {
	return spinnerRenderer{
		layout:  layout.NewVBoxLayout(),
		objects: []fyne.CanvasObject{s.upIcon, s.downIcon},
	}
}

func (s spinnerRenderer) Destroy() {
	// Do nothing
}

func (s spinnerRenderer) Layout(size fyne.Size) {
	s.layout.Layout(s.Objects(), s.MinSize())
}

func (s spinnerRenderer) MinSize() fyne.Size {
	return s.layout.MinSize(s.Objects())
}

func (s spinnerRenderer) Objects() []fyne.CanvasObject {
	return s.objects
}

func (s spinnerRenderer) Refresh() {
	s.Layout(s.MinSize())
}
