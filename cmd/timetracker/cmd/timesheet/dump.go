package timesheet

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	ttErrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
)

var (
	// DumpCmd represents the command that prints all timesheets to the console
	DumpCmd = &cobra.Command{
		Use:     "dump",
		Aliases: []string{"d"},
		Short:   "Dumps all timesheets",
		RunE:    dumpTimesheets,
	}
	startDate   string
	endDate     string
	withDeleted bool
)

func init() {
	DumpCmd.Flags().StringVar(&startDate, "startDate", "", "start date (YYYY-MM-DD)")
	DumpCmd.Flags().StringVar(&endDate, "endDate", "", "end date (YYYY-MM-DD)")
	DumpCmd.Flags().BoolVar(&withDeleted, "deleted", false, "include deleted timesheets")
}

func dumpTimesheets(_ *cobra.Command, _ []string) (err error) {
	log := logger.GetLogger("dumpTimesheets")
	if startDate == "" || endDate == "" {
		return errors.New("both start date and end date must be specified")
	}
	var sheets []models.TimesheetData
	var dStart, dEnd time.Time
	dStart, err = time.Parse(constants.TimestampDateLayout, startDate)
	if err != nil {
		cli.PrintAndLogError(log, err, "error parsing %s as a start date", startDate)
		return err
	}
	dEnd, err = time.Parse(constants.TimestampDateLayout, endDate)
	if err != nil {
		cli.PrintAndLogError(log, err, "error parsing %s as an end date", endDate)
		return err
	}
	timesheet := models.NewTimesheet()
	timesheet.Data().StartTime = dStart
	err = timesheet.Data().StopTime.Scan(dEnd)
	if err != nil {
		return err
	}
	sheets, err = timesheet.SearchDateRange(withDeleted)
	if err != nil {
		cli.PrintAndLogError(log, err, ttErrors.ListTimesheetError)
		return err
	}
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "Timesheet ID"},
			{Text: "Task ID"},
			{Text: "Synopsis"},
			{Text: "Started At"},
			{Text: "Stopped At"},
			{Text: "Duration"},
		},
	}
	for _, sheet := range sheets {
		starttimedisplay := sheet.StartTime.Format(constants.TimestampLayout)
		stoptimedisplay := "RUNNING"
		durationdisplay := "(unknown)"
		if sheet.StopTime.Valid {
			stoptimedisplay = sheet.StopTime.Time.Format(constants.TimestampLayout)
			durationdisplay = sheet.StopTime.Time.Sub(sheet.StartTime).Truncate(time.Second).String()
		}
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(sheet.ID))},
			{Text: strconv.Itoa(int(sheet.Task.ID))},
			{Text: sheet.Task.Synopsis},
			{Text: starttimedisplay},
			{Text: stoptimedisplay},
			{Text: durationdisplay},
		}
		table.Body.Cells = append(table.Body.Cells, rec)
	}
	if len(table.Body.Cells) == 0 {
		fmt.Println(color.WhiteString("There are no timesheets"))
		return nil
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}
