package task

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/neflyte/timetracker/lib/constants"
	tterrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/neflyte/timetracker/lib/ui/cli"
	"github.com/spf13/cobra"
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
	taskData := models.NewTask()
	taskData.Data().ID, taskData.Data().Synopsis = taskData.Resolve(args[0])
	// Load the task to make sure it exists
	err = taskData.Load(false)
	if err != nil {
		cli.PrintAndLogError(log, err, tterrors.LoadTaskError)
		return err
	}
	// Stop any running task
	err = cli.StopRunningTimesheet()
	if err != nil {
		return err
	}
	// Create a new timesheet for the task
	taskdisplay := taskData.Data().Synopsis
	if taskData.Data().ID > 0 {
		taskdisplay = fmt.Sprintf("%d", taskData.Data().ID)
	}
	log.Debug().Msgf("taskdisplay=%s", taskdisplay)
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *taskData.Data()
	timesheet.Data().StartTime = time.Now()
	err = timesheet.Create()
	if err != nil {
		cli.PrintAndLogError(log, err, "%s for task %s", tterrors.CreateTimesheetError, taskdisplay)
		return err
	}
	fmt.Println(
		color.WhiteString("Task ID %d ", taskData.Data().ID),
		color.CyanString(taskData.Data().Synopsis),
		color.MagentaString("(%s) ", taskData.Data().Description),
		color.GreenString("started"),
		color.WhiteString("at %s", timesheet.Data().StartTime.Format(constants.TimestampLayout)),
	)
	return nil
}
