package windows

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/rs/zerolog"
	"time"
)

var tableHeader = []string{"Task ID", "Synopsis", "Started On", "Duration"}

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
	startDateInput   *widgets.DateEntry
	startDateBinding binding.String
	endDateInput     *widgets.DateEntry
	endDateBinding   binding.String
	runReportButton  *widget.Button

	resultTable  *widget.Table
	tableColumns int
	tableRows    int

	taskReport models.TaskReport
}

func newReportWindow(app fyne.App) reportWindow {
	newWindow := &reportWindowData{
		Window:       app.NewWindow("Report"),
		Logger:       logger.GetStructLogger("reportWindowData"),
		tableRows:    0,
		tableColumns: 0,
		taskReport:   make([]models.TaskReportData, 0),
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
	w.startDateInput = widgets.NewDateEntry()
	w.startDateInput.PlaceHolder = "YYYY-MM-DD"
	w.startDateInput.Validator = func(value string) error {
		if value == "" {
			return nil
		}
		_, err := time.Parse(constants.TimestampDateLayout, value)
		return err
	}
	w.startDateInput.Bind(w.startDateBinding)
	w.endDateInput = widgets.NewDateEntry()
	w.endDateInput.PlaceHolder = "YYYY-MM-DD"
	w.endDateInput.Validator = func(value string) error {
		if value == "" {
			return nil
		}
		_, err := time.Parse(constants.TimestampDateLayout, value)
		return err
	}
	w.endDateInput.Bind(w.endDateBinding)
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
			return w.tableRows, w.tableColumns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(cell widget.TableCellID, object fyne.CanvasObject) {
			var labelText string

			label, isLabel := object.(*widget.Label)
			if !isLabel {
				w.Logger.Error().Msgf("expected *widget.Label but got %T instead", object)
				return
			}
			if cell.Row == 0 {
				labelText = tableHeader[cell.Col]
				label.TextStyle.Bold = true
			} else {
				taskReportData := w.taskReport[cell.Row-1]
				switch cell.Col {
				case 0:
					labelText = fmt.Sprintf("%d", taskReportData.TaskID)
				case 1:
					labelText = taskReportData.TaskSynopsis
				case 2:
					labelText = taskReportData.StartDate.Time.Format(constants.TimestampDateLayout)
				case 3:
					labelText = taskReportData.Duration().String()
				default:
					labelText = ""
				}
				label.TextStyle.Bold = false
			}
			label.SetText(labelText)
		},
	)
	w.resultTable.SetColumnWidth(0, 75)
	w.resultTable.SetColumnWidth(1, 250)
	w.resultTable.SetColumnWidth(2, 100)
	w.Container = container.NewMax(container.NewBorder(
		w.headerContainer, nil, nil, nil,
		container.NewPadded(container.NewVBox(widget.NewSeparator(), w.resultTable)),
	))
	w.Window.SetContent(w.Container)
	w.Window.SetCloseIntercept(w.Hide)
	w.Window.SetFixedSize(true)
	w.Window.Resize(minimumWindowSize)
	return nil
}

// Show displays the window and focuses the start date entry widget
func (w *reportWindowData) Show() {
	w.Window.Canvas().Focus(w.startDateInput)
	w.Window.Show()
}

func (w *reportWindowData) doRunReport() {
	// Validate date range
	dStart, dEnd, err := w.validateDateRange()
	if err != nil {
		// TODO: Show a more informative error
		dialog.NewError(err, w).Show()
		w.Logger.Err(err).Msg("unable to validate date range")
		return
	}
	// Disable run button and enable when this function is done
	w.runReportButton.Disable()
	defer func() {
		w.runReportButton.Enable()
		w.resultTable.Refresh()
	}()
	// Clear table
	//  - set tableRows to zero
	w.tableRows = 0
	// Run query
	timesheet := models.NewTimesheet()
	// TODO: Add option to include deleted tasks
	reportData, err := timesheet.TaskReport(dStart, dEnd, false)
	if err != nil {
		// TODO: Show a more informative error
		dialog.NewError(err, w).Show()
		w.Logger.Err(err).Msgf("error running task report between %s and %s", dStart.Format(constants.TimestampDateLayout), dEnd.Format(constants.TimestampDateLayout))
		return
	}
	w.Logger.Debug().Msgf("loaded %d records for report", len(reportData))
	// Populate table
	w.taskReport = reportData.Clone()
	w.tableColumns = len(tableHeader)
	w.tableRows = len(w.taskReport) + 1
}

func (w *reportWindowData) validateDateRange() (startDate time.Time, endDate time.Time, err error) {
	// Parse the start date
	startDateString, err := w.startDateBinding.Get()
	if err != nil {
		return
	}
	startDate, err = time.Parse(constants.TimestampDateLayout, startDateString)
	if err != nil {
		err = errors.InvalidTaskReportStartDate{
			StartDate: startDateString,
			Wrapped:   err,
		}
		return
	}
	// Parse the end date
	endDateString, err := w.endDateBinding.Get()
	if err != nil {
		return
	}
	endDate, err = time.Parse(constants.TimestampDateLayout, endDateString)
	if err != nil {
		err = errors.InvalidTaskReportEndDate{
			EndDate: endDateString,
			Wrapped: err,
		}
		return
	}
	// Check if end date happens before start date
	if endDate.Before(startDate) {
		err = errors.InvalidTaskReportEndDate{
			EndDate: endDateString,
			Wrapped: fmt.Errorf(
				"end date (%s) cannot happen before start date (%s)",
				endDate.Format(constants.TimestampDateLayout),
				startDate.Format(constants.TimestampDateLayout),
			),
		}
		return
	}
	// Date range is valid
	return
}
