package task

import (
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var (
	StartCmd = &cobra.Command{
		Use:   "start [task id]",
		Short: "Start a task",
		Args:  cobra.ExactArgs(1),
		RunE:  startTask,
	}
)

func startTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("startTask")
	taskid, err := strconv.Atoi(args[0])
	if err != nil {
		log.Printf("error converting argument (%s) into a number: %s\n", args[0], err)
		return err
	}
	task := new(models.Task)
	err = database.DB.Find(&task, uint(taskid)).Error
	if err != nil {
		log.Printf("error reading task id %d: %s\n", uint(taskid), err)
		return err
	}
	err = utils.StopRunningTask()
	if err != nil {
		log.Printf("error stopping running task: %s\n", err)
		return err
	}
	timesheet := new(models.Timesheet)
	timesheet.TaskID = task.ID
	timesheet.StartTime = time.Now()
	err = database.DB.Create(&timesheet).Error
	if err != nil {
		log.Printf("error creating timesheet for task id %d: %s\n", task.ID, err)
		return err
	}
	log.Printf("task id %d (timesheet id %d) started\n", task.ID, timesheet.ID)
	return nil
}
