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
	startDateInput   *widget.Entry
	startDateBinding binding.String
	endDateInput     *widget.Entry
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
			return w.tableRows, w.tableColumns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(cell widget.TableCellID, object fyne.CanvasObject) {
			label, isLabel := object.(*widget.Label)
			if !isLabel {
				w.Logger.Error().Msgf("expected *widget.Label but got %T instead", object)
				return
			}
			labelText := ""
			if cell.Row == 0 {
				labelText = tableHeader[cell.Col]
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
			}
			label.SetText(labelText)
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
	dStart, dEnd, err := w.validateDateRange()
	if err != nil {
		// TODO: Show a more informative error
		dialog.NewError(err, w).Show()
		w.Logger.Err(err).Msg("unable to validate date range")
		return
	}
	// Disable run button and enable when this function is done
	w.runReportButton.Disable()
	defer w.runReportButton.Enable()
	// Clear table
	//  - set tableRows to zero
	w.tableRows = 0
	//  - refresh table
	w.resultTable.Refresh()
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
	// Populate table
	w.taskReport = reportData.Clone()
	w.tableColumns = len(tableHeader)
	w.tableRows = len(reportData) + 1
	w.resultTable.Refresh()
}

func (w *reportWindowData) validateDateRange() (startDate time.Time, endDate time.Time, err error) {
	// Parse the start date
	startDateString, _ := w.startDateBinding.Get()
	startDate, err = time.Parse(constants.TimestampDateLayout, startDateString)
	if err != nil {
		err = errors.InvalidTaskReportStartDate{
			StartDate: startDateString,
			Wrapped:   err,
		}
		return
	}
	// Parse the end date
	endDateString, _ := w.endDateBinding.Get()
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
