package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
)

type reportWindow interface {
	windowBase
}

type reportWindowData struct {
	fyne.Window
	zerolog.Logger

	Container *fyne.Container

	headerContainer  *fyne.Container
	startDateLabel   *widget.Label
	endDateLabel     *widget.Label
	startDateInput   *widget.Entry
	startDateBinding binding.String
	endDateInput     *widget.Entry
	endDateBinding   binding.String
	runReportButton  *widget.Button

	resultTable *widget.Table
}

func newReportWindow(app fyne.App) reportWindow {
	newWindow := &reportWindowData{
		Window: app.NewWindow("Report"),
		Logger: logger.GetStructLogger("reportWindowData"),
	}
	err := newWindow.Init()
	if err != nil {
		newWindow.Logger.Err(err).Msg("error initializing window")
	}
	return newWindow
}

// Init initializes the window
func (w *reportWindowData) Init() error {
	w.startDateBinding = binding.NewString()
	w.endDateBinding = binding.NewString()
	w.startDateInput = widget.NewEntryWithData(w.startDateBinding)
	w.endDateInput = widget.NewEntryWithData(w.endDateBinding)
	w.startDateLabel = widget.NewLabel("Start date:")
	w.endDateLabel = widget.NewLabel("End date:")
	w.runReportButton = widget.NewButton("Run Report", w.doRunReport)
	w.headerContainer = container.NewHBox(
		w.startDateLabel,
		w.startDateInput,
		w.endDateLabel,
		w.endDateInput,
		w.runReportButton,
	)
	w.resultTable = widget.NewTable(
		func() (int, int) {

		},
		func() fyne.CanvasObject {

		},
		func(widget.TableCellID, fyne.CanvasObject) {

		},
	)
	w.Container = container.NewBorder(w.headerContainer, nil, nil, nil, w.resultTable)
	w.Window.SetContent(w.Container)
	w.Window.SetCloseIntercept(w.Hide)
	w.Window.SetFixedSize(true)
	w.Window.Resize(minimumWindowSize)
	return nil
}

func (w *reportWindowData) doRunReport() {
	// Validate date range
	// Disable run button
	// Clear table
	// Run query
	// Populate table
	// Enable run button
}
