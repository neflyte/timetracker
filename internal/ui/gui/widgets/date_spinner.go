package widgets

import (
	"fmt"
	"fyne.io/fyne/v2"
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
	DateSpinnerSubmitEventKey = "submit"
	DateSpinnerCancelEventKey = "cancel"

	minYear           = 0
	maxYear           = 9999
	minMonth          = 1
	maxMonth          = 12
	minDay            = 1
	maxDay            = 31
	channelBufferSize = 1
)

type DateSpinner struct {
	widget.BaseWidget
	fmt.Stringer

	log            zerolog.Logger
	dateFormat     string
	yearBinding    binding.String
	monthBinding   binding.String
	dayBinding     binding.String
	submitChannel  chan rxgo.Item
	cancelChannel  chan rxgo.Item
	observablesMap map[string]rxgo.Observable
}

func NewDateSpinner(format string) *DateSpinner {
	dateSpinner := &DateSpinner{
		log:           logger.GetStructLogger("DateSpinner"),
		dateFormat:    format,
		yearBinding:   binding.NewString(),
		monthBinding:  binding.NewString(),
		dayBinding:    binding.NewString(),
		submitChannel: make(chan rxgo.Item, channelBufferSize),
		cancelChannel: make(chan rxgo.Item, channelBufferSize),
	}
	dateSpinner.ExtendBaseWidget(dateSpinner)
	dateSpinner.observablesMap = map[string]rxgo.Observable{
		DateSpinnerSubmitEventKey: rxgo.FromEventSource(dateSpinner.submitChannel),
		DateSpinnerCancelEventKey: rxgo.FromEventSource(dateSpinner.cancelChannel),
	}
	dateSpinner.SetTime(time.Now())
	return dateSpinner
}

func (dp *DateSpinner) String() string {
	return dp.Time().Format(dp.dateFormat)
}

func (dp *DateSpinner) SetString(date string) {
	dateValue, err := time.Parse(dp.dateFormat, date)
	if err != nil {
		dp.log.Err(err).Msgf("error parsing string %s as format %s", date, dp.dateFormat)
	} else {
		dp.SetTime(dateValue)
	}
}

func (dp *DateSpinner) Time() time.Time {
	return time.Date(
		dp.GetYear(),
		time.Month(dp.GetMonth()),
		dp.GetDay(),
		0, 0, 0, 0, nil,
	)
}

func (dp *DateSpinner) SetTime(t time.Time) {
	dp.SetYear(t.Year())
	dp.SetMonth(int(t.Month()))
	dp.SetDay(t.Day())
}

func (dp *DateSpinner) SetYear(year int) {
	if year >= minYear && year <= maxYear {
		yearString := strconv.Itoa(year)
		err := dp.yearBinding.Set(yearString)
		if err != nil {
			dp.log.Err(err).Msgf("error setting yearBinding to %s", yearString)
		}
	}
}

func (dp *DateSpinner) SetMonth(month int) {
	if month >= minMonth && month <= maxMonth {
		monthString := strconv.Itoa(month)
		err := dp.monthBinding.Set(monthString)
		if err != nil {
			dp.log.Err(err).Msgf("error setting monthBinding to %s", monthString)
		}
	}
}

func (dp *DateSpinner) SetDay(day int) {
	if day >= minDay && day <= maxDay {
		dayString := strconv.Itoa(day)
		err := dp.dayBinding.Set(dayString)
		if err != nil {
			dp.log.Err(err).Msgf("error setting dayBinding to %s", dayString)
		}
	}
}

func (dp *DateSpinner) GetYear() int {
	yearString, err := dp.yearBinding.Get()
	if err != nil {
		dp.log.Err(err).Msg("error getting value from yearBinding")
		return 0
	}
	yearInt, err := strconv.Atoi(yearString)
	if err != nil {
		dp.log.Err(err).Msgf("error converting yearString %s to an int", yearString)
		return 0
	}
	return yearInt
}

func (dp *DateSpinner) GetMonth() int {
	monthString, err := dp.monthBinding.Get()
	if err != nil {
		dp.log.Err(err).Msg("error getting value from monthBinding")
		return 0
	}
	monthInt, err := strconv.Atoi(monthString)
	if err != nil {
		dp.log.Err(err).Msgf("error converting monthString %s to an int", monthString)
		return 0
	}
	return monthInt
}

func (dp *DateSpinner) GetDay() int {
	dayString, err := dp.dayBinding.Get()
	if err != nil {
		dp.log.Err(err).Msg("error getting value from dayBinding")
		return 0
	}
	dayInt, err := strconv.Atoi(dayString)
	if err != nil {
		dp.log.Err(err).Msgf("error converting dayString %s to an int", dayString)
		return 0
	}
	return dayInt
}

// Observables returns a map of Observable objects used by this widget
func (dp *DateSpinner) Observables() map[string]rxgo.Observable {
	return dp.observablesMap
}

func (dp *DateSpinner) CreateRenderer() fyne.WidgetRenderer {
	dp.ExtendBaseWidget(dp)
	r := &dateSpinnerRenderer{
		log:           dp.log.With().Str("renderer", "dateSpinnerRenderer").Logger(),
		canvasObjects: make([]fyne.CanvasObject, 0),
		yearEntry:     NewNumericalEntry(),
		monthEntry:    NewNumericalEntry(),
		dayEntry:      NewNumericalEntry(),
		submitChannel: dp.submitChannel,
		cancelChannel: dp.cancelChannel,
		layout:        layout.NewHBoxLayout(),
	}
	r.yearEntry.Bind(dp.yearBinding)
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
	r.yearSpinner = NewSpinner(r.doAddYear, r.doSubYear)
	r.monthEntry.Bind(dp.monthBinding)
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
	r.monthSpinner = NewSpinner(r.doAddMonth, r.doSubMonth)
	r.dayEntry.Bind(dp.dayBinding)
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
	r.daySpinner = NewSpinner(r.doAddDay, r.doSubDay)
	r.okButton = widget.NewButtonWithIcon("", theme.ConfirmIcon(), r.doSubmit)
	r.cancelButton = widget.NewButtonWithIcon("", theme.CancelIcon(), r.doCancel)
	r.canvasObjects = append(
		r.canvasObjects,
		r.yearEntry, r.yearSpinner, widget.NewSeparator(),
		r.monthEntry, r.monthSpinner, widget.NewSeparator(),
		r.dayEntry, r.daySpinner,
		r.cancelButton,
		r.okButton,
	)
	return r
}

type dateSpinnerRenderer struct {
	log           zerolog.Logger
	canvasObjects []fyne.CanvasObject
	layout        fyne.Layout

	yearEntry     *NumericalEntry
	yearSpinner   *Spinner
	monthEntry    *NumericalEntry
	monthSpinner  *Spinner
	dayEntry      *NumericalEntry
	daySpinner    *Spinner
	okButton      *widget.Button
	cancelButton  *widget.Button
	submitChannel chan rxgo.Item
	cancelChannel chan rxgo.Item
}

func (d dateSpinnerRenderer) Destroy() {
	// Does nothing
}

func (d dateSpinnerRenderer) Layout(size fyne.Size) {
	d.layout.Layout(d.canvasObjects, size)
}

func (d dateSpinnerRenderer) MinSize() fyne.Size {
	return d.layout.MinSize(d.canvasObjects)
}

func (d dateSpinnerRenderer) Objects() []fyne.CanvasObject {
	return d.canvasObjects
}

func (d dateSpinnerRenderer) Refresh() {
	d.Layout(d.MinSize())
}

func (d *dateSpinnerRenderer) doAddYear() {
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

func (d *dateSpinnerRenderer) doSubYear() {
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

func (d *dateSpinnerRenderer) doAddMonth() {
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

func (d *dateSpinnerRenderer) doSubMonth() {
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

func (d *dateSpinnerRenderer) doAddDay() {
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

func (d *dateSpinnerRenderer) doSubDay() {
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

func (d *dateSpinnerRenderer) validateEntries() error {
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

func (d *dateSpinnerRenderer) doSubmit() {
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

func (d *dateSpinnerRenderer) doCancel() {
	d.cancelChannel <- rxgo.Of(true)
}
