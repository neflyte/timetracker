package widgets

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	"strconv"
	"time"
)

const (
	dateStringFormat  = "2006-01-02"
	minYear           = 0
	maxYear           = 9999
	minMonth          = 1
	maxMonth          = 12
	minDay            = 1
	maxDay            = 31
	channelBufferSize = 1
)

type DatePicker struct {
	widget.DisableableWidget
	fmt.Stringer

	log          zerolog.Logger
	yearValue    string
	monthValue   string
	dayValue     string
	yearBinding  binding.String
	monthBinding binding.String
	dayBinding   binding.String
	yearChan     chan rxgo.Item
	monthChan    chan rxgo.Item
	dayChan      chan rxgo.Item
}

func NewDatePicker() *DatePicker {
	datePicker := &DatePicker{
		log:          logger.GetStructLogger("DatePicker"),
		yearBinding:  binding.NewString(),
		monthBinding: binding.NewString(),
		dayBinding:   binding.NewString(),
		yearChan:     make(chan rxgo.Item, channelBufferSize),
		monthChan:    make(chan rxgo.Item, channelBufferSize),
		dayChan:      make(chan rxgo.Item, channelBufferSize),
	}
	datePicker.ExtendBaseWidget(datePicker)
	timeNow := time.Now()
	_ = datePicker.yearBinding.Set(strconv.Itoa(timeNow.Year()))
	_ = datePicker.monthBinding.Set(strconv.Itoa(int(timeNow.Month())))
	_ = datePicker.dayBinding.Set(strconv.Itoa(timeNow.Day()))
	return datePicker
}

func (dp *DatePicker) String() string {
	return fmt.Sprintf("%s-%s-%s", dp.yearValue, dp.monthValue, dp.dayValue)
}

func (dp *DatePicker) Time() (*time.Time, error) {
	dateStringValue := fmt.Sprintf("%s-%s-%s", dp.yearValue, dp.monthValue, dp.dayValue)
	timeValue, err := time.Parse(dateStringFormat, dateStringValue)
	if err != nil {
		return nil, err
	}
	return &timeValue, nil
}

func (dp *DatePicker) CreateRenderer() fyne.WidgetRenderer {
	dp.ExtendBaseWidget(dp)
	r := &datePickerRenderer{
		log:           dp.log.With().Str("renderer", "datePickerRenderer").Logger(),
		canvasObjects: make([]fyne.CanvasObject, 0),
		yearEntry:     widget.NewEntryWithData(dp.yearBinding),
		monthEntry:    widget.NewEntryWithData(dp.monthBinding),
		dayEntry:      widget.NewEntryWithData(dp.dayBinding),
		layout:        layout.NewHBoxLayout(),
	}
	r.yearAddButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), r.doAddYear)
	r.yearSubButton = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), r.doSubYear)
	r.yearVBox = container.NewVBox(r.yearAddButton, r.yearSubButton)
	r.monthAddButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), r.doAddMonth)
	r.monthSubButton = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), r.doSubMonth)
	r.monthVBox = container.NewVBox(r.monthAddButton, r.monthSubButton)
	r.dayAddButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), r.doAddDay)
	r.daySubButton = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), r.doSubDay)
	r.dayVBox = container.NewVBox(r.dayAddButton, r.daySubButton)
	r.canvasObjects = append(
		r.canvasObjects,
		r.yearEntry, r.yearVBox,
		r.monthEntry, r.monthVBox,
		r.dayEntry, r.dayVBox,
	)
	return r
}

type datePickerRenderer struct {
	log           zerolog.Logger
	canvasObjects []fyne.CanvasObject
	layout        fyne.Layout

	yearEntry      *widget.Entry
	yearAddButton  *widget.Button
	yearSubButton  *widget.Button
	yearVBox       *fyne.Container
	monthEntry     *widget.Entry
	monthAddButton *widget.Button
	monthSubButton *widget.Button
	monthVBox      *fyne.Container
	dayEntry       *widget.Entry
	dayAddButton   *widget.Button
	daySubButton   *widget.Button
	dayVBox        *fyne.Container
}

func (d datePickerRenderer) Destroy() {
	// Does nothing
}

func (d datePickerRenderer) Layout(size fyne.Size) {
	d.layout.Layout(d.canvasObjects, size)
}

func (d datePickerRenderer) MinSize() fyne.Size {
	return d.layout.MinSize(d.canvasObjects)
}

func (d datePickerRenderer) Objects() []fyne.CanvasObject {
	return d.canvasObjects
}

func (d datePickerRenderer) Refresh() {
	d.Layout(d.MinSize())
}

func (d *datePickerRenderer) doAddYear() {
	yearValue, err := strconv.Atoi(d.yearEntry.Text)
	if err != nil {
		d.log.Err(err).Msgf("error converting %s to an int", d.yearEntry.Text)
		return
	}
	if yearValue >= maxYear {
		return
	}
	yearValue++
	d.yearEntry.SetText(strconv.Itoa(yearValue))
}

func (d *datePickerRenderer) doSubYear() {
	yearValue, err := strconv.Atoi(d.yearEntry.Text)
	if err != nil {
		d.log.Err(err).Msgf("error converting %s to an int", d.yearEntry.Text)
		return
	}
	if yearValue <= minYear {
		return
	}
	yearValue--
	d.yearEntry.SetText(strconv.Itoa(yearValue))
}

func (d *datePickerRenderer) doAddMonth() {
	monthValue, err := strconv.Atoi(d.monthEntry.Text)
	if err != nil {
		d.log.Err(err).Msgf("error converting %s to an int", d.monthEntry.Text)
		return
	}
	if monthValue >= maxMonth {
		return
	}
	monthValue++
	d.monthEntry.SetText(strconv.Itoa(monthValue))
}

func (d *datePickerRenderer) doSubMonth() {
	monthValue, err := strconv.Atoi(d.monthEntry.Text)
	if err != nil {
		d.log.Err(err).Msgf("error converting %s to an int", d.monthEntry.Text)
		return
	}
	if monthValue <= minMonth {
		return
	}
	monthValue--
	d.monthEntry.SetText(strconv.Itoa(monthValue))
}

func (d *datePickerRenderer) doAddDay() {
	dayValue, err := strconv.Atoi(d.dayEntry.Text)
	if err != nil {
		d.log.Err(err).Msgf("error converting %s to an int", d.dayEntry.Text)
		return
	}
	if dayValue >= maxDay {
		return
	}
	dayValue++
	d.dayEntry.SetText(strconv.Itoa(dayValue))
}

func (d *datePickerRenderer) doSubDay() {
	dayValue, err := strconv.Atoi(d.dayEntry.Text)
	if err != nil {
		d.log.Err(err).Msgf("error converting %s to an int", d.dayEntry.Text)
		return
	}
	if dayValue <= minDay {
		return
	}
	dayValue--
	d.dayEntry.SetText(strconv.Itoa(dayValue))
}
