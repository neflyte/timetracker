package timesheet

import (
	"errors"
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var (
	// ReportCmd represents the command that reports on completed tasks
	ReportCmd = &cobra.Command{
		Use:     "report",
		Aliases: []string{"r"},
		Short:   "Report on tasks completed in the specified time period",
		RunE:    reportTimesheets,
	}
	reportStartDate string
	reportEndDate   string
)

func init() {
	ReportCmd.Flags().StringVar(&reportStartDate, "startDate", "", "start date (YYYY-MM-DD)")
	ReportCmd.Flags().StringVar(&reportEndDate, "endDate", "", "end date (YYYY-MM-DD)")
	ReportCmd.Flags().BoolVar(&withDeleted, "deleted", false, "include deleted timesheets")
}

func reportTimesheets(_ *cobra.Command, _ []string) (err error) {
	log := logger.GetLogger("reportTimesheets")
	if reportStartDate == "" || reportEndDate == "" {
		return errors.New("both start date and end date must be specified")
	}
	var dStart, dEnd time.Time
	dStart, err = time.Parse(constants.TimestampDateLayout, reportStartDate)
	if err != nil {
		cli.PrintAndLogError(log, err, "error parsing %s as the start date", reportStartDate)
		return err
	}
	dEnd, err = time.Parse(constants.TimestampDateLayout, reportEndDate)
	if err != nil {
		cli.PrintAndLogError(log, err, "error parsing %s as the end date", reportEndDate)
		return err
	}
	timesheet := models.NewTimesheet()
	reportData, reportErr := timesheet.TaskReport(dStart, dEnd, withDeleted)
	if reportErr != nil {
		cli.PrintAndLogError(log, reportErr, "error running task report between %s and %s", reportStartDate, reportEndDate)
		return reportErr
	}
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "Task ID"},
			{Text: "Synopsis"},
			{Text: "Started On"},
			{Text: "Duration"},
		},
	}
	for _, reportDataEntry := range reportData {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: strconv.Itoa(int(reportDataEntry.TaskID))},
			{Text: reportDataEntry.TaskSynopsis},
			{Text: reportDataEntry.StartDate.Time.Format(constants.TimestampDateLayout)},
			{Text: reportDataEntry.Duration().String()},
		})
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}
