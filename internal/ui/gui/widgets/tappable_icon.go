package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TappableIcon struct {
	widget.Icon

	OnTapped func()
}

func NewTappableIcon(res fyne.Resource, onTapped func()) *TappableIcon {
	icon := new(TappableIcon)
	icon.ExtendBaseWidget(icon)
	icon.SetResource(res)
	icon.OnTapped = onTapped
	return icon
}

func (t *TappableIcon) Tapped(_ *fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}
