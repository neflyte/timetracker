package task

import (
	"database/sql"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var (
	StopCmd = &cobra.Command{
		Use:   "stop [task id]",
		Short: "Stop a running task",
		Args:  cobra.ExactArgs(1),
		RunE:  stopTask,
	}
)

func stopTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("stopTask")
	taskid, err := strconv.Atoi(args[0])
	if err != nil {
		log.Printf("error converting argument (%s) into a number: %s\n", args[0], err)
		return err
	}
	timesheet := new(models.Timesheet)
	result := database.DB.Where("task_id = ? AND stop_time = ?", uint(taskid), nil).First(&timesheet)
	if result.Error != nil {
		log.Printf("error looking for started task: %s\n", result.Error)
		return result.Error
	}
	stoptime := new(sql.NullTime)
	err = stoptime.Scan(time.Now())
	if err != nil {
		log.Printf("error scanning time.Now() into sql.NullTime: %s\n", err)
		return err
	}
	timesheet.StopTime = *stoptime
	err = database.DB.Save(&timesheet).Error
	if err != nil {
		log.Printf("error stopping task id %d (timesheet id %d): %s\n", timesheet.TaskID, timesheet.ID, err)
		return err
	}
	log.Printf("task id %d (timesheet id %d) stopped\n", timesheet.TaskID, timesheet.ID)
	return nil
}
