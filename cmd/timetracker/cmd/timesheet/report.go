package timesheet

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
)

const (
	exportFileMode = 0644
)

var (
	// ReportCmd represents the command that reports on completed tasks
	ReportCmd = &cobra.Command{
		Use:     "report",
		Aliases: []string{"r"},
		Short:   "Report on tasks completed in the specified time period",
		RunE:    reportTimesheets,
	}
	csvTableHeader  = []string{"task_id", "synopsis", "started_on", "duration"}
	reportStartDate string
	reportEndDate   string
	exportCSVFile   string
)

func init() {
	ReportCmd.Flags().StringVar(&reportStartDate, "startDate", "", "start date (YYYY-MM-DD)")
	ReportCmd.Flags().StringVar(&reportEndDate, "endDate", "", "end date (YYYY-MM-DD)")
	ReportCmd.Flags().BoolVar(&withDeleted, "deleted", false, "include deleted timesheets")
	ReportCmd.Flags().StringVar(&exportCSVFile, "exportCSV", "", "file to export report in CSV format")
}

// FIXME: Figure out a way to reduce cyclomatic complexity so the lint exclusion can be removed
func reportTimesheets(_ *cobra.Command, _ []string) (err error) { //nolint:cyclop
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
	if exportCSVFile != "" {
		outFile, fileErr := os.OpenFile(exportCSVFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, exportFileMode)
		if fileErr != nil {
			cli.PrintAndLogError(log, fileErr, "error opening output file %s", exportCSVFile)
			return fileErr
		}
		defer func() {
			closeErr := outFile.Close()
			if closeErr != nil {
				cli.PrintAndLogError(log, closeErr, "error closing output file %s", exportCSVFile)
			}
		}()
		csvOut := csv.NewWriter(outFile)
		defer csvOut.Flush()
		csvData := make([][]string, 0)
		csvData = append(csvData, csvTableHeader)
		for _, taskReportData := range reportData {
			csvData = append(csvData, []string{
				fmt.Sprintf("%d", taskReportData.TaskID),
				taskReportData.TaskSynopsis,
				taskReportData.StartDate.Time.Format(constants.TimestampDateLayout),
				taskReportData.Duration().String(),
			})
		}
		writeErr := csvOut.WriteAll(csvData)
		if writeErr != nil {
			cli.PrintAndLogError(log, writeErr, "error exporting data to CSV file %s", exportCSVFile)
			return writeErr
		}
		fmt.Printf("Exported %d records to file %s\n", len(reportData), exportCSVFile)
		return nil
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
