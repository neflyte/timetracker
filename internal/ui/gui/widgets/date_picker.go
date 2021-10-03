package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
	"time"
)

type DatePicker struct {
	widget.BaseWidget

	log          zerolog.Logger
	dateBinding  binding.String
	dateFormat   string
	parentCanvas fyne.Canvas
}

func NewDatePicker(format string, bind *binding.String, canvas fyne.Canvas) *DatePicker {
	dp := &DatePicker{
		log:          logger.GetStructLogger("DatePicker"),
		parentCanvas: canvas,
		dateFormat:   format,
	}
	dp.ExtendBaseWidget(dp)
	if bind != nil {
		dp.dateBinding = *bind
	} else {
		dp.dateBinding = binding.NewString()
	}
	timeNow := time.Now().Format(format)
	err := dp.dateBinding.Set(timeNow)
	if err != nil {
		dp.log.Err(err).Msgf("error setting dateBinding to %s", timeNow)
	}
	return dp
}

func (d *DatePicker) Bind(bind binding.String) {
	d.dateBinding = bind
}

func (d *DatePicker) Binding() binding.String {
	return d.dateBinding
}

func (d *DatePicker) CreateRenderer() fyne.WidgetRenderer {
	r := datePickerRenderer{
		log:          d.log.With().Str("renderer", "datePickerRenderer").Logger(),
		dateSpinner:  NewDateSpinner(d.dateFormat),
		dateEntry:    NewDateEntryV2(d.dateFormat),
		widgetLayout: layout.NewHBoxLayout(),
	}
	r.dateEntry.Bind(d.dateBinding)
	r.dateSpinnerIcon = NewTappableIcon(theme.MenuDropDownIcon(), r.doShowSpinner)
	r.dateSpinnerPopup = widget.NewPopUp(r.dateSpinner, d.parentCanvas)
	r.dateSpinner.Observables()[DateSpinnerSubmitEventKey].ForEach(
		func(value interface{}) {
			dateStringValue, ok := value.(string)
			if ok {
				err := d.dateBinding.Set(dateStringValue)
				if err != nil {
					r.log.Err(err).Msg("error setting dateBinding value")
				}
				r.dateSpinnerPopup.Hide()
			}
		},
		func(err error) {
			r.log.Err(err).Msg("error from dateSpinner submit observable")
		},
		func() {
			r.log.Debug().Msg("dateSpinner submit observable finished")
		},
	)
	r.dateSpinner.Observables()[DateSpinnerCancelEventKey].ForEach(
		func(value interface{}) {
			cancelled, ok := value.(bool)
			if ok && cancelled {
				r.log.Debug().Msg("dateSpinner cancel observable fired")
			}
			r.dateSpinnerPopup.Hide()
		},
		func(err error) {
			r.log.Err(err).Msg("error from dateSpinner cancel observable")
		},
		func() {
			r.log.Debug().Msg("dateSpinner cancel observable finished")
		},
	)
	r.objects = []fyne.CanvasObject{
		r.dateEntry,
		r.dateSpinnerIcon,
	}
	r.sizeObjects = []fyne.CanvasObject{
		r.dateEntry,
		r.dateSpinnerIcon,
		r.dateSpinnerPopup,
	}
	return r
}

type datePickerRenderer struct {
	_ fyne.WidgetRenderer

	log          zerolog.Logger
	widgetLayout fyne.Layout
	objects      []fyne.CanvasObject
	sizeObjects  []fyne.CanvasObject

	dateEntry        *DateEntryV2
	dateSpinnerIcon  *TappableIcon
	dateSpinner      *DateSpinner
	dateSpinnerPopup *widget.PopUp
}

func (d datePickerRenderer) doShowSpinner() {
	if d.dateEntry.Binding() != nil {
		dateString, err := d.dateEntry.Binding().Get()
		if err != nil {
			d.log.Err(err).Msg("error getting string from dateEntry binding")
		} else {
			d.dateSpinner.SetString(dateString)
			d.dateSpinnerPopup.ShowAtPosition(d.getPopupPosition())
		}
	}
}

func (d datePickerRenderer) getPopupPosition() fyne.Position {
	entryPos := d.dateEntry.Position()
	return fyne.NewPos(
		entryPos.X,
		entryPos.Y+d.dateEntry.MinSize().Height+theme.Padding(),
	)
}

func (d datePickerRenderer) Destroy() {
	d.dateEntry.Unbind()
}

func (d datePickerRenderer) Layout(size fyne.Size) {
	d.widgetLayout.Layout(d.Objects(), size)
}

func (d datePickerRenderer) MinSize() fyne.Size {
	return d.widgetLayout.MinSize(d.sizeObjects)
}

func (d datePickerRenderer) Objects() []fyne.CanvasObject {
	return d.objects
}

func (d datePickerRenderer) Refresh() {
	d.Layout(d.MinSize())
}
