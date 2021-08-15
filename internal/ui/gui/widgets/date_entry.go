package widgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"time"
)

// DateEntry is an entry widget which accepts a date
type DateEntry struct {
	widget.Entry

	minWidth      float32
	dateFormat    string
	stringBinding binding.String
}

// NewDateEntry returns a new DateEntry widget
func NewDateEntry(minWidth float32, placeholder string, dateFormat string, stringBinding binding.String) *DateEntry {
	newEntry := &DateEntry{
		dateFormat:    dateFormat,
		minWidth:      minWidth,
		stringBinding: stringBinding,
	}
	newEntry.ExtendBaseWidget(newEntry)
	newEntry.PlaceHolder = placeholder
	if newEntry.stringBinding != nil {
		newEntry.stringBinding.AddListener(binding.NewDataListener(newEntry.validateDate))
	}
	return newEntry
}

// MinSize is the minimum size of the widget
func (d *DateEntry) MinSize() fyne.Size {
	d.ExtendBaseWidget(d)
	entrySize := d.Entry.MinSize()
	minWidth := entrySize.Width
	if minWidth < d.minWidth {
		minWidth = d.minWidth
	}
	return fyne.NewSize(minWidth, entrySize.Height)
}

func (d *DateEntry) validateDate() {
	if d.stringBinding != nil {
		stringValue, err := d.stringBinding.Get()
		if err != nil {
			d.SetValidationError(err)
			return
		}
		if stringValue == "" {
			d.SetValidationError(errors.New("date cannot be empty"))
			return
		}
		_, err = time.Parse(d.dateFormat, stringValue)
		d.SetValidationError(err)
	}
}
