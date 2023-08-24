package timesheet

import (
	"encoding/csv"
	"encoding/xml"
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
	exportFileMode   = 0644
	outputFormatText = "text"
	outputFormatCSV  = "csv"
	outputFormatJSON = "json"
	outputFormatXML  = "xml"
)

var (
	// ReportCmd represents the command that reports on completed tasks
	ReportCmd = &cobra.Command{
		Use:     "report",
		Aliases: []string{"r"},
		Short:   "Report on tasks completed in the specified time period",
		RunE:    reportTimesheets,
	}
	csvTableHeader     = []string{"task_id", "synopsis", "started_on", "duration"}
	reportStartDate    string
	reportEndDate      string
	exportCSVFile      string
	reportOutputFormat string
)

func init() {
	ReportCmd.Flags().StringVar(&reportStartDate, "startDate", "", "start date (YYYY-MM-DD)")
	ReportCmd.Flags().StringVar(&reportEndDate, "endDate", "", "end date (YYYY-MM-DD)")
	ReportCmd.Flags().BoolVar(&withDeleted, "deleted", false, "include deleted timesheets")
	ReportCmd.Flags().StringVar(&exportCSVFile, "exportCSV", "", "file to export report in CSV format")
	ReportCmd.Flags().StringVar(&reportOutputFormat, "outputFormat", outputFormatText, "output format (text, csv, json, xml; default text)")
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
	// Do CSV export to file if requested
	if exportCSVFile != "" {
		err = exportToCSV(reportData, exportCSVFile)
		if err != nil {
			return err
		}
		fmt.Printf("Exported %d records to file %s\n", len(reportData), exportCSVFile)
		return nil
	}
	// Print the report in the requested output format
	printReport(reportData, reportOutputFormat)
	return nil
}

func printReport(reportData models.TaskReport, reportFormat string) {
	log := logger.GetLogger("printReport")
	// Output using requested format
	switch reportFormat {
	case outputFormatText:
		printReportTable(reportData)
	case outputFormatCSV:
		cli.PrintCSV(log, reportData)
	case outputFormatJSON:
		jsonData := struct {
			TaskReport []models.TaskReportData `json:"Data"`
		}{
			TaskReport: reportData,
		}
		cli.PrintJSON(log, jsonData)
	case outputFormatXML:
		xmlData := struct {
			XMLName    xml.Name                `xml:"TaskReport"`
			TaskReport []models.TaskReportData `xml:"Data"`
		}{
			TaskReport: reportData,
		}
		cli.PrintXML(log, xmlData)
	}
}

func printReportTable(reportData models.TaskReport) {
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
}

func exportToCSV(reportData models.TaskReport, exportFile string) error {
	log := logger.GetLogger("exportToCSV")
	outFile, fileErr := os.OpenFile(exportFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, exportFileMode)
	if fileErr != nil {
		cli.PrintAndLogError(log, fileErr, "error opening output file %s", exportFile)
		return fileErr
	}
	defer func() {
		closeErr := outFile.Close()
		if closeErr != nil {
			cli.PrintAndLogError(log, closeErr, "error closing output file %s", exportFile)
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
		cli.PrintAndLogError(log, writeErr, "error exporting data to CSV file %s", exportFile)
		return writeErr
	}
	return nil
}
