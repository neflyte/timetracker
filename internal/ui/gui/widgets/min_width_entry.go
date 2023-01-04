package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// MinWidthEntry is an entry widget which has a minimum width
type MinWidthEntry struct {
	widget.Entry
	minWidth float32
}

// NewDateEntry returns a new MinWidthEntry widget
func NewDateEntry(minWidth float32, placeholder string) *MinWidthEntry {
	newEntry := &MinWidthEntry{
		minWidth: minWidth,
	}
	newEntry.ExtendBaseWidget(newEntry)
	newEntry.SetPlaceHolder(placeholder)
	return newEntry
}

// MinSize is the minimum size of the widget
func (d *MinWidthEntry) MinSize() fyne.Size {
	d.ExtendBaseWidget(d)
	entrySize := d.Entry.MinSize()
	minWidth := entrySize.Width
	if minWidth < d.minWidth {
		minWidth = d.minWidth
	}
	return fyne.NewSize(minWidth, entrySize.Height)
}
