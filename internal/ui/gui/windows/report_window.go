package windows

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/constants"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/rs/zerolog"
)

const (
	dateEntryMinWidth  = 120.0
	columnTaskID       = 0
	columnTaskSynopsis = 1
	columnStartDate    = 2
	columnDuration     = 3
)

var (
	tableColumnWidths = []float32{75, 250, 100, 100}
	tableHeader       = []string{"Task ID", "Synopsis", "Started On", "Duration"}
	csvTableHeader    = []string{"task_id", "synopsis", "started_on", "duration"}
)

type reportWindow interface {
	windowBase
}

type reportWindowData struct {
	fyne.Window
	log              zerolog.Logger
	container        *fyne.Container
	headerContainer  *fyne.Container
	startDateLabel   *widget.Label
	endDateLabel     *widget.Label
	startDateEntry   *widgets.MinWidthEntry
	startDateBinding binding.String
	endDateEntry     *widgets.MinWidthEntry
	endDateBinding   binding.String
	runReportButton  *widget.Button
	exportButton     *widget.Button
	resultTable      *widget.Table
	tableColumns     int
	tableRows        int
	taskReport       models.TaskReport
}

func newReportWindow(app fyne.App) reportWindow {
	log := logger.GetLogger("newReportWindow")
	newWindow := &reportWindowData{
		Window:           app.NewWindow("Task Report"), // i18n
		log:              logger.GetStructLogger("reportWindowData"),
		startDateBinding: binding.NewString(),
		endDateBinding:   binding.NewString(),
		tableRows:        0,
		tableColumns:     0,
		taskReport:       make([]models.TaskReportData, 0),
	}
	err := newWindow.Init()
	if err != nil {
		log.Err(err).
			Msg("error initializing window")
	}
	return newWindow
}

// Init initializes the window
func (w *reportWindowData) Init() error {
	// Header container
	w.startDateEntry = widgets.NewMinWidthEntry(dateEntryMinWidth, "YYYY-MM-DD") // l10n
	w.startDateEntry.Bind(w.startDateBinding)
	w.startDateEntry.Validator = w.dateValidator
	w.endDateEntry = widgets.NewMinWidthEntry(dateEntryMinWidth, "YYYY-MM-DD") // l10n
	w.endDateEntry.Bind(w.endDateBinding)
	w.endDateEntry.Validator = w.dateValidator
	w.startDateLabel = widget.NewLabel("Start date:")                                         // i18n
	w.endDateLabel = widget.NewLabel("End date:")                                             // i18n
	w.runReportButton = widget.NewButtonWithIcon("RUN", theme.MediaPlayIcon(), w.doRunReport) // i18n
	w.exportButton = widget.NewButtonWithIcon("EXPORT", theme.DownloadIcon(), w.doExport)     // i18n
	w.exportButton.Disable()                                                                  // Initialize the export button in a disabled state
	w.headerContainer = container.NewBorder(
		nil, nil,
		container.NewHBox(
			w.startDateLabel, w.startDateEntry,
			w.endDateLabel, w.endDateEntry,
		),
		container.NewHBox(
			w.runReportButton, w.exportButton,
		),
	)
	// Result table
	w.resultTable = widget.NewTable(w.resultTableLength, w.resultTableCreate, w.resultTableUpdate)
	// Set table column widths
	for idx, colWidth := range tableColumnWidths {
		w.resultTable.SetColumnWidth(idx, colWidth)
	}
	w.container = container.NewPadded(
		container.NewBorder(
			w.headerContainer, nil, nil, nil,
			w.resultTable,
		),
	)
	w.Window.SetContent(w.container)
	w.Window.SetIcon(icons.IconV2)
	w.Window.SetCloseIntercept(w.Window.Hide)
	w.Window.Resize(minimumWindowSize)
	w.Window.Canvas().Focus(w.startDateEntry)
	return nil
}

// Show displays the window and focuses the start date entry widget
func (w *reportWindowData) Show() {
	w.Window.Canvas().Focus(w.startDateEntry)
	w.Window.Show()
}

func (w *reportWindowData) resultTableLength() (int, int) {
	return w.tableRows, w.tableColumns
}

func (w *reportWindowData) resultTableCreate() fyne.CanvasObject {
	return widget.NewLabel("")
}

func (w *reportWindowData) resultTableUpdate(cell widget.TableCellID, object fyne.CanvasObject) {
	var (
		log       = logger.GetFuncLogger(w.log, "resultTableUpdate")
		labelText string
	)
	label, isLabel := object.(*widget.Label)
	if !isLabel {
		log.Error().
			Str("unexpectedType", reflect.TypeOf(object).String()).
			Msg("expected *widget.Label but got unexpected type")
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
}

func (w *reportWindowData) dateValidator(entry string) error {
	if len(entry) == 0 {
		return nil
	}
	_, err := time.Parse(constants.TimestampDateLayout, entry)
	return err
}

func (w *reportWindowData) doRunReport() {
	log := logger.GetFuncLogger(w.log, "doRunReport")
	// Validate date range
	dStart, dEnd, err := w.validateDateRange()
	if err != nil {
		log.Err(err).
			Msg("unable to validate date range")
		// TODO: Show a more informative error
		dialog.NewError(err, w).Show()
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
		w.Window.Content().Refresh()
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
		log.Err(err).
			Str("startDate", dStart.Format(constants.TimestampDateLayout)).
			Str("endDate", dEnd.Format(constants.TimestampDateLayout)).
			Msg("error running task report")
		return
	}
	log.Debug().
		Int("count", len(reportData)).
		Msgf("loaded records for report")
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
		log.Err(dialogErr).
			Msg("error opening CSV file for writing")
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
				log.Err(err).
					Msg("error closing CSV file")
			}
		}()
		err := csvOut.WriteAll(csvData)
		if err != nil {
			log.Err(err).
				Msg("error exporting data to CSV")
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Export Successful", "The report data was exported successfully.", w)
	}
}

func (w *reportWindowData) validateDateRange() (startDate time.Time, endDate time.Time, err error) {
	log := logger.GetFuncLogger(w.log, "validateDateRange")
	// Parse the start date
	startDateString, err := w.startDateBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("error getting start date from binding")
		return
	}
	startDate, err = time.Parse(constants.TimestampDateLayout, startDateString)
	if err != nil {
		log.Err(err).
			Str("startDate", startDateString).
			Str("timestampFormat", constants.TimestampDateLayout).
			Msg("error parsing start date timestamp")
		err = tterrors.InvalidTaskReportStartDate{
			StartDate: startDateString,
			Wrapped:   err,
		}
		return
	}
	// Parse the end date
	endDateString, err := w.endDateBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("error getting end date from binding")
		return
	}
	endDate, err = time.Parse(constants.TimestampDateLayout, endDateString)
	if err != nil {
		log.Err(err).
			Str("endDate", endDateString).
			Str("timestampFormat", constants.TimestampDateLayout).
			Msg("error parsing end date timestamp")
		err = tterrors.InvalidTaskReportEndDate{
			EndDate: endDateString,
			Wrapped: err,
		}
		return
	}
	// Check if end date happens before start date
	if endDate.Before(startDate) {
		log.Error().
			Str("startDate", startDate.String()).
			Str("endDate", endDate.String()).
			Msg("end date cannot happen before start date")
		err = tterrors.InvalidTaskReportEndDate{
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
