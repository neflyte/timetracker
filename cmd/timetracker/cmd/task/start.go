package task

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
	"time"
)

var (
	// StartCmd represents the command to start a task
	StartCmd = &cobra.Command{
		Use:     "start [task id/synopsis]",
		Aliases: []string{"s"},
		Short:   "Start a task",
		Args:    cobra.ExactArgs(1),
		RunE:    startTask,
	}
)

func startTask(_ *cobra.Command, args []string) (err error) {
	log := logger.GetLogger("startTask")
	taskData := models.NewTaskData()
	taskData.ID, taskData.Synopsis = taskData.Resolve(args[0])
	taskdisplay := taskData.Synopsis
	if taskData.ID > 0 {
		taskdisplay = fmt.Sprintf("%d", taskData.ID)
	}
	log.Debug().Msgf("taskdisplay=%s", taskdisplay)
	// Load the task to make sure it exists
	err = models.Task(taskData).Load(false)
	if err != nil {
		cli.PrintAndLogError(log, err, errors.LoadTaskError)
		return err
	}
	// Stop any running task
	stoppedTimesheet, err := models.Task(taskData).StopRunningTask()
	if err != nil && err.Error() != errors.NoRunningTasksError {
		cli.PrintAndLogError(log, err, errors.StopRunningTaskError)
		return err
	}
	if stoppedTimesheet != nil {
		log.Info().Msgf("task id %d (timesheet id %d) stopped\n", stoppedTimesheet.Task.ID, stoppedTimesheet.ID)
		fmt.Println(
			color.WhiteString("Task ID %d", stoppedTimesheet.Task.ID),
			color.YellowString("stopped"),
			color.WhiteString("at %s", stoppedTimesheet.StopTime.Time.Format(constants.TimestampLayout)),
			color.BlueString(stoppedTimesheet.StopTime.Time.Sub(stoppedTimesheet.StartTime).Truncate(time.Second).String()),
		)
	}
	// Create a new timesheet for the task
	timesheetData := new(models.TimesheetData)
	timesheetData.Task = *taskData
	timesheetData.StartTime = time.Now()
	err = models.Timesheet(timesheetData).Create()
	if err != nil {
		cli.PrintAndLogError(log, err, "%s for task %s", errors.CreateTimesheetError, taskdisplay)
		return err
	}
	fmt.Println(
		color.WhiteString("Task ID %d ", taskData.ID),
		color.CyanString(taskData.Synopsis),
		color.MagentaString("(%s) ", taskData.Description),
		color.GreenString("started"),
		color.WhiteString("at %s", timesheetData.StartTime.Format(constants.TimestampLayout)),
	)
	return nil
}
