package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
)

const (
	dateEntryMinWidth = 150.0
)

// DateEntry is an entry widget which accepts a date
type DateEntry struct {
	widget.Entry
	log zerolog.Logger
}

// NewDateEntry returns a new DateEntry widget
func NewDateEntry() *DateEntry {
	newEntry := &DateEntry{
		log: logger.GetStructLogger("DateEntry"),
	}
	newEntry.ExtendBaseWidget(newEntry)
	return newEntry
}

// MinSize is the minimum size of the widget
func (d *DateEntry) MinSize() fyne.Size {
	entrySize := d.Entry.MinSize()
	minWidth := entrySize.Width
	if minWidth < dateEntryMinWidth {
		minWidth = dateEntryMinWidth
	}
	return fyne.NewSize(minWidth, entrySize.Height)
}
