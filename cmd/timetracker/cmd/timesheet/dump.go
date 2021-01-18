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
)

var (
	DumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dumps all timesheets",
		RunE:  dumpTimesheets,
	}
)

func dumpTimesheets(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("dumpTimesheets")
	timesheet := new(models.Timesheet)
	rows, err := database.DB.
		Model(timesheet).
		Select("timesheets.id, tasks.id, tasks.synopsis, timesheets.start_time, timesheets.stop_time").
		Joins("left join tasks on timesheets.task_id = tasks.id").
		Rows()
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
		stopTime := "(running)"
		if sheet.StopTime.Valid {
			stopTime = sheet.StopTime.Time.Format(`2006-01-02 15:04:05 PM`)
		}
		// log.Printf(sheet.String())
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(sheet.ID))},
			{Text: strconv.Itoa(int(task.ID))},
			{Text: task.Synopsis},
			{Text: sheet.StartTime.Format(`2006-01-02 15:04:05 PM`)},
			{Text: stopTime},
		}
		table.Body.Cells = append(table.Body.Cells, rec)
	}
	if len(table.Body.Cells) == 0 {
		fmt.Println(chalk.White, "There are no timesheets")
		return nil
	}
	fmt.Println(table.String())
	return nil
}
