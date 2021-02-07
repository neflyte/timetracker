package task

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/spf13/cobra"
	"time"
)

var (
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
	taskData := new(models.TaskData)
	taskData.ID, taskData.Synopsis = utils.ResolveTask(args[0])
	taskdisplay := taskData.Synopsis
	if taskData.ID > 0 {
		taskdisplay = fmt.Sprintf("%d", taskData.ID)
	}
	log.Debug().Msgf("taskdisplay=%s", taskdisplay)
	// Load the task to make sure it exists
	err = models.Task(taskData).Load(false)
	if err != nil {
		utils.PrintAndLogError(errors.LoadTaskError, err, log)
		return err
	}
	// Stop any running task
	err = models.Task(taskData).StopRunningTask()
	if err != nil {
		utils.PrintAndLogError(errors.StopRunningTaskError, err, log)
		return err
	}
	// Create a new timesheet for the task
	timesheetData := new(models.TimesheetData)
	timesheetData.Task = *taskData
	timesheetData.StartTime = time.Now()
	err = models.Timesheet(timesheetData).Create()
	if err != nil {
		utils.PrintAndLogError(fmt.Sprintf("%s for task %s", errors.CreateTimesheetError, taskdisplay), err, log)
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
