package timesheet

import (
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"strconv"
	"time"
)

var (
	DumpCmd = &cobra.Command{
		Use:     "dump",
		Aliases: []string{"d"},
		Short:   "Dumps all timesheets",
		RunE:    dumpTimesheets,
	}
	startDate string
	endDate   string
)

func init() {
	DumpCmd.Flags().StringVar(&startDate, "startDate", "", "start date (YYYY-MM-DD)")
	DumpCmd.Flags().StringVar(&endDate, "endDate", "", "end date (YYYY-MM-DD)")
}

func dumpTimesheets(_ *cobra.Command, _ []string) (err error) {
	log := logger.GetLogger("dumpTimesheets")
	db := database.DB.Joins("Task")
	sheets := make([]models.Timesheet, 0)
	if startDate != "" && endDate != "" {
		var dStart, dEnd time.Time
		dStart, err = time.Parse(constants.TimestampDateLayout, startDate)
		if err != nil {
			return err
		}
		dEnd, err = time.Parse(constants.TimestampDateLayout, endDate)
		if err != nil {
			return err
		}
		db = db.Where("start_time >= ? AND stop_time <= ?", dStart, dEnd)
	}
	db = db.Find(&sheets)
	if db.Error != nil {
		log.Printf("error querying for timesheets: %s\n", db.Error)
		return db.Error
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
		fmt.Println(chalk.White, "There are no timesheets")
		return nil
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}
