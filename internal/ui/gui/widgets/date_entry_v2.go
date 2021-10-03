package widgets

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"time"
)

// DateEntryV2 is an entry widget which accepts a date
type DateEntryV2 struct {
	widget.Entry

	dateFormat   string
	dateBinding  binding.String
	dateListener binding.DataListener
}

// NewDateEntryV2 returns a new DateEntryV2 widget
func NewDateEntryV2(format string) *DateEntryV2 {
	newEntry := &DateEntryV2{
		dateFormat: format,
	}
	newEntry.ExtendBaseWidget(newEntry)
	newEntry.dateListener = binding.NewDataListener(newEntry.validateDate)
	return newEntry
}

func (d *DateEntryV2) Bind(data binding.String) {
	if d.dateBinding != nil {
		d.dateBinding.RemoveListener(d.dateListener)
	}
	d.dateBinding = data
	if data != nil {
		data.AddListener(d.dateListener)
	}
	d.Entry.Bind(data)
}

func (d *DateEntryV2) Binding() binding.String {
	return d.dateBinding
}

// MinSize is the minimum size of the widget
func (d *DateEntryV2) MinSize() fyne.Size {
	d.ExtendBaseWidget(d)
	entrySize := d.Entry.MinSize()
	return fyne.NewSize(entrySize.Width+theme.IconInlineSize(), entrySize.Height)
}

func (d *DateEntryV2) validateDate() {
	if d.dateBinding != nil {
		stringValue, err := d.dateBinding.Get()
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
