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
	DatePickerSubmitEventKey = "submit"
	DatePickerCancelEventKey = "cancel"

	dateStringFormat  = "2006-01-02"
	minYear           = 0
	maxYear           = 9999
	minMonth          = 1
	maxMonth          = 12
	minDay            = 1
	maxDay            = 31
	channelBufferSize = 2
)

type DatePicker struct {
	widget.BaseWidget
	fmt.Stringer

	log            zerolog.Logger
	yearBinding    binding.String
	monthBinding   binding.String
	dayBinding     binding.String
	submitChannel  chan rxgo.Item
	cancelChannel  chan rxgo.Item
	observablesMap map[string]rxgo.Observable
}

func NewDatePicker() *DatePicker {
	datePicker := &DatePicker{
		log:           logger.GetStructLogger("DatePicker"),
		yearBinding:   binding.NewString(),
		monthBinding:  binding.NewString(),
		dayBinding:    binding.NewString(),
		submitChannel: make(chan rxgo.Item, channelBufferSize),
		cancelChannel: make(chan rxgo.Item, channelBufferSize),
	}
	datePicker.ExtendBaseWidget(datePicker)
	datePicker.observablesMap = map[string]rxgo.Observable{
		DatePickerSubmitEventKey: rxgo.FromEventSource(datePicker.submitChannel),
		DatePickerCancelEventKey: rxgo.FromEventSource(datePicker.cancelChannel),
	}
	timeNow := time.Now()
	yearString := strconv.Itoa(timeNow.Year())
	err := datePicker.yearBinding.Set(yearString)
	if err != nil {
		datePicker.log.Err(err).Msgf("error setting yearBinding to %s", yearString)
	}
	monthString := strconv.Itoa(int(timeNow.Month()))
	err = datePicker.monthBinding.Set(monthString)
	if err != nil {
		datePicker.log.Err(err).Msgf("error setting monthBinding to %s", monthString)
	}
	dayString := strconv.Itoa(timeNow.Day())
	err = datePicker.dayBinding.Set(dayString)
	if err != nil {
		datePicker.log.Err(err).Msgf("error setting dayBinding to %s", dayString)
	}
	return datePicker
}

func (dp *DatePicker) String() string {
	yearValue, err := dp.yearBinding.Get()
	if err != nil {
		dp.log.Err(err).Msg("error getting value from yearBinding")
		yearValue = "????"
	}
	monthValue, err := dp.monthBinding.Get()
	if err != nil {
		dp.log.Err(err).Msg("error getting value from monthBinding")
		monthValue = "??"
	}
	dayValue, err := dp.dayBinding.Get()
	if err != nil {
		dp.log.Err(err).Msg("error getting value from dayBinding")
		dayValue = "??"
	}
	return fmt.Sprintf("%s-%s-%s", yearValue, monthValue, dayValue)
}

func (dp *DatePicker) Time() (*time.Time, error) {
	dateStringValue := dp.String()
	timeValue, err := time.Parse(dateStringFormat, dateStringValue)
	if err != nil {
		return nil, err
	}
	return &timeValue, nil
}

// Observables returns a map of Observable objects used by this widget
func (dp *DatePicker) Observables() map[string]rxgo.Observable {
	return dp.observablesMap
}

func (dp *DatePicker) CreateRenderer() fyne.WidgetRenderer {
	dp.ExtendBaseWidget(dp)
	r := &datePickerRenderer{
		log:           dp.log.With().Str("renderer", "datePickerRenderer").Logger(),
		canvasObjects: make([]fyne.CanvasObject, 0),
		yearEntry:     widget.NewEntryWithData(dp.yearBinding),
		monthEntry:    widget.NewEntryWithData(dp.monthBinding),
		dayEntry:      widget.NewEntryWithData(dp.dayBinding),
		submitChannel: dp.submitChannel,
		cancelChannel: dp.cancelChannel,
		layout:        layout.NewHBoxLayout(),
	}
	r.yearEntry.Validator = func(yearValue string) error {
		yearIntValue, err := strconv.Atoi(yearValue)
		if err != nil {
			return err
		}
		if yearIntValue < minYear {
			return fmt.Errorf("year cannot be earlier than %d", minYear)
		}
		if yearIntValue > maxYear {
			return fmt.Errorf("year cannot be later than %d", maxYear)
		}
		return nil
	}
	r.yearAddButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), r.doAddYear)
	r.yearSubButton = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), r.doSubYear)
	r.yearVBox = container.NewVBox(r.yearAddButton, r.yearSubButton)
	r.monthEntry.Validator = func(monthValue string) error {
		monthIntValue, err := strconv.Atoi(monthValue)
		if err != nil {
			return err
		}
		if monthIntValue < minMonth {
			return fmt.Errorf("month cannot be earlier than %d", minMonth)
		}
		if monthIntValue > maxMonth {
			return fmt.Errorf("month cannot be later than %d", maxMonth)
		}
		return nil
	}
	r.monthAddButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), r.doAddMonth)
	r.monthSubButton = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), r.doSubMonth)
	r.monthVBox = container.NewVBox(r.monthAddButton, r.monthSubButton)
	r.dayEntry.Validator = func(dayValue string) error {
		dayIntValue, err := strconv.Atoi(dayValue)
		if err != nil {
			return err
		}
		if dayIntValue < minDay {
			return fmt.Errorf("day cannot be earlier than %d", minDay)
		}
		if dayIntValue > maxDay {
			return fmt.Errorf("day cannot be later than %d", maxDay)
		}
		return nil
	}
	r.dayAddButton = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), r.doAddDay)
	r.daySubButton = widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), r.doSubDay)
	r.dayVBox = container.NewVBox(r.dayAddButton, r.daySubButton)
	r.okButton = widget.NewButtonWithIcon("", theme.ConfirmIcon(), r.doSubmit)
	r.cancelButton = widget.NewButtonWithIcon("", theme.CancelIcon(), r.doCancel)
	r.canvasObjects = append(
		r.canvasObjects,
		r.yearEntry, r.yearVBox,
		r.monthEntry, r.monthVBox,
		r.dayEntry, r.dayVBox,
		r.cancelButton,
		r.okButton,
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
	okButton       *widget.Button
	cancelButton   *widget.Button
	submitChannel  chan rxgo.Item
	cancelChannel  chan rxgo.Item
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

func (d *datePickerRenderer) validateEntries() error {
	err := d.yearEntry.Validate()
	if err != nil {
		return err
	}
	err = d.monthEntry.Validate()
	if err != nil {
		return err
	}
	err = d.dayEntry.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (d *datePickerRenderer) doSubmit() {
	err := d.validateEntries()
	if err != nil {
		d.log.Err(err).Msg("error validating entry fields; cannot submit")
		return
	}
	dateString := fmt.Sprintf(
		"%s-%s-%s",
		d.yearEntry.Text,
		d.monthEntry.Text,
		d.dayEntry.Text,
	)
	d.submitChannel <- rxgo.Of(dateString)
}

func (d *datePickerRenderer) doCancel() {
	d.cancelChannel <- rxgo.Of(true)
}
