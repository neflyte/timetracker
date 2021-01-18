package timesheet

import (
	"fmt"
	"github.com/alexeyco/simpletable"
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
		Use:   "dump",
		Short: "Dumps all timesheets",
		RunE:  dumpTimesheets,
	}
	startDate string
	endDate   string
)

func init() {
	DumpCmd.Flags().StringVar(&startDate, "startDate", "", "start date (YYYY-MM-DD)")
	DumpCmd.Flags().StringVar(&endDate, "endDate", "", "end date (YYYY-MM-DD)")
}

func dumpTimesheets(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("dumpTimesheets")
	timesheet := new(models.Timesheet)
	db := database.DB.
		Model(timesheet).
		Select("timesheets.id, tasks.id, tasks.synopsis, timesheets.start_time, timesheets.stop_time").
		Joins("left join tasks on timesheets.task_id = tasks.id")
	if startDate != "" && endDate != "" {
		dStart, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return err
		}
		dEnd, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return err
		}
		db = db.Where("timesheets.start_time >= ? AND timesheets.stop_time <= ?", dStart, dEnd)
	}
	rows, err := db.Rows()
	if err != nil {
		log.Printf("error getting result rows: %s\n", err)
		return err
	}
	defer database.CloseRows(rows)
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
	for rows.Next() {
		sheet := new(models.Timesheet)
		task := new(models.Task)
		err = rows.Scan(&sheet.ID, &task.ID, &task.Synopsis, &sheet.StartTime, &sheet.StopTime)
		if err != nil {
			fmt.Println(chalk.Red, "Error scanning result row:", chalk.White, chalk.Dim.TextStyle(err.Error()))
			log.Err(err).Msg("error scanning result row")
			return err
		}
		starttimedisplay := sheet.StartTime.Format(`2006-01-02 15:04:05 PM`)
		stoptimedisplay := "(running)"
		durationdisplay := "(unknown)"
		if sheet.StopTime.Valid {
			stoptimedisplay = sheet.StopTime.Time.Format(`2006-01-02 15:04:05 PM`)
			diff := sheet.StopTime.Time.Sub(sheet.StartTime)
			durationdisplay = diff.String()
		}
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(sheet.ID))},
			{Text: strconv.Itoa(int(task.ID))},
			{Text: task.Synopsis},
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
