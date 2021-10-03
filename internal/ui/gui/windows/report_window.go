package windows

import (
	"encoding/csv"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/rs/zerolog"
	"time"
)

const (
	dateEntryMinWidth  = 150.0
	columnTaskID       = 0
	columnTaskSynopsis = 1
	columnStartDate    = 2
	columnDuration     = 3
)

var (
	tableColumnWidths = []float32{75, 250, 100}
	tableHeader       = []string{"Task ID", "Synopsis", "Started On", "Duration"}
	csvTableHeader    = []string{"task_id", "synopsis", "started_on", "duration"}
)

type reportWindow interface {
	windowBase
}

type reportWindowData struct {
	fyne.Window

	log zerolog.Logger

	Container *fyne.Container

	headerContainer  *fyne.Container
	startDateLabel   *widget.Label
	endDateLabel     *widget.Label
	startDateBinding binding.String
	startDatePicker  *widgets.DatePicker
	endDateInput     *widgets.DateEntry
	endDateBinding   binding.String
	runReportButton  *widget.Button
	exportButton     *widget.Button

	resultTable  *widget.Table
	tableColumns int
	tableRows    int

	taskReport models.TaskReport
}

func newReportWindow(app fyne.App) reportWindow {
	log := logger.GetLogger("newReportWindow")
	newWindow := &reportWindowData{
		Window:       app.NewWindow("Task Report"),
		log:          logger.GetStructLogger("reportWindowData"),
		tableRows:    0,
		tableColumns: 0,
		taskReport:   make([]models.TaskReportData, 0),
	}
	err := newWindow.Init()
	if err != nil {
		log.Err(err).Msg("error initializing window")
	}
	return newWindow
}

// Init initializes the window
func (w *reportWindowData) Init() error {
	// Header container
	w.startDateBinding = binding.NewString()
	w.endDateBinding = binding.NewString()
	w.startDatePicker = widgets.NewDatePicker(constants.TimestampDateLayout, &w.startDateBinding, w.Window.Canvas())
	w.endDateInput = widgets.NewDateEntry(dateEntryMinWidth, "YYYY-MM-DD", constants.TimestampDateLayout, w.endDateBinding)
	w.endDateInput.Bind(w.endDateBinding)
	w.startDateLabel = widget.NewLabel("Start date:")
	w.endDateLabel = widget.NewLabel("End date:")
	w.runReportButton = widget.NewButtonWithIcon("RUN", theme.MediaPlayIcon(), w.doRunReport)
	w.exportButton = widget.NewButtonWithIcon("EXPORT", theme.DownloadIcon(), w.doExport)
	w.exportButton.Disable() // Initialize the export button in a disabled state
	w.headerContainer = container.NewBorder(
		nil, nil,
		container.NewHBox(
			w.startDateLabel,
			w.startDatePicker,
			w.endDateLabel,
			w.endDateInput,
		),
		container.NewHBox(
			w.runReportButton,
			w.exportButton,
		),
	)
	// Result table
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
				w.log.Error().Msgf("expected *widget.Label but got %T instead", object)
				return
			}
			if cell.Row == 0 {
				labelText = tableHeader[cell.Col]
				label.TextStyle.Bold = true
			} else {
				taskReportData := w.taskReport[cell.Row-1]
				switch cell.Col {
				case columnTaskID:
					labelText = fmt.Sprintf("%d", taskReportData.TaskID)
				case columnTaskSynopsis:
					labelText = taskReportData.TaskSynopsis
				case columnStartDate:
					labelText = taskReportData.StartDate.Time.Format(constants.TimestampDateLayout)
				case columnDuration:
					labelText = taskReportData.Duration().String()
				default:
					labelText = ""
				}
				label.TextStyle.Bold = false
			}
			label.SetText(labelText)
		},
	)
	// Set table column widths
	for idx, colWidth := range tableColumnWidths {
		w.resultTable.SetColumnWidth(idx, colWidth)
	}
	w.Container = container.NewPadded(
		container.NewBorder(
			w.headerContainer, nil, nil, nil,
			w.resultTable,
		),
	)
	w.Window.SetContent(container.NewMax(w.Container))
	w.Window.SetCloseIntercept(w.Hide)
	w.Window.SetFixedSize(true)
	w.Window.Resize(minimumWindowSize)
	// FIXME: w.Window.Canvas().Focus(w.startDatePicker)
	return nil
}

// Show displays the window and focuses the start date entry widget
func (w *reportWindowData) Show() {
	// FIXME: w.Window.Canvas().Focus(w.startDatePicker)
	w.Window.Show()
}

func (w *reportWindowData) doRunReport() {
	// Validate date range
	dStart, dEnd, err := w.validateDateRange()
	if err != nil {
		// TODO: Show a more informative error
		dialog.NewError(err, w).Show()
		w.log.Err(err).Msg("unable to validate date range")
		return
	}
	// Disable run button and enable when this function is done
	w.exportButton.Disable()
	w.runReportButton.Disable()
	defer func() {
		w.runReportButton.Enable()
		if len(w.taskReport) > 0 {
			w.exportButton.Enable()
		}
		w.Content().Refresh()
	}()
	// Clear table
	//  - set tableRows to zero
	w.tableRows = 0
	//  - set taskReport to empty
	w.taskReport = make(models.TaskReport, 0)
	// Run query
	timesheet := models.NewTimesheet()
	// TODO: Add option to include deleted tasks
	reportData, err := timesheet.TaskReport(dStart, dEnd, false)
	if err != nil {
		// TODO: Show a more informative error
		dialog.NewError(err, w).Show()
		w.log.Err(err).Msgf("error running task report between %s and %s", dStart.Format(constants.TimestampDateLayout), dEnd.Format(constants.TimestampDateLayout))
		return
	}
	w.log.Debug().Msgf("loaded %d records for report", len(reportData))
	// Populate table
	w.taskReport = reportData.Clone()
	w.tableColumns = len(tableHeader)
	w.tableRows = len(w.taskReport) + 1
}

func (w *reportWindowData) doExport() {
	// Sanity check that there is data to export
	if len(w.taskReport) == 0 {
		dialog.NewError(
			fmt.Errorf("there is no data to export"),
			w,
		).Show()
		return
	}
	// Show file save dialog
	dialog.ShowFileSave(w.exportReportAsCSV, w)
}

func (w *reportWindowData) exportReportAsCSV(writeCloser fyne.URIWriteCloser, dialogErr error) {
	log := logger.GetFuncLogger(w.log, "exportReportAsCSV")
	if dialogErr != nil {
		log.Err(dialogErr).Msg("error opening CSV file for writing")
		return
	}
	// Write data to file
	if writeCloser != nil {
		// Collect data
		csvData := make([][]string, 0)
		csvData = append(csvData, csvTableHeader)
		for _, taskReportData := range w.taskReport {
			csvData = append(csvData, []string{
				fmt.Sprintf("%d", taskReportData.TaskID),
				taskReportData.TaskSynopsis,
				taskReportData.StartDate.Time.Format(constants.TimestampDateLayout),
				taskReportData.Duration().String(),
			})
		}
		csvOut := csv.NewWriter(writeCloser)
		defer func() {
			csvOut.Flush()
			err := writeCloser.Close()
			if err != nil {
				log.Err(err).Msg("error closing CSV file")
			}
		}()
		err := csvOut.WriteAll(csvData)
		if err != nil {
			log.Err(err).Msg("error exporting data to CSV")
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Export Successful", "The report data was exported successfully.", w)
	}
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
