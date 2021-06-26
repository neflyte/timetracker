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
	"strconv"
	"time"
)

const (
	createCmdMaxNumberArgs = 2
)

var (
	// CreateCmd is the definition of the create command
	CreateCmd = &cobra.Command{
		Use:     "create [synopsis] [description]",
		Aliases: []string{"c"},
		Short:   "Create a task",
		Args:    cobra.MaximumNArgs(createCmdMaxNumberArgs),
		RunE:    createTask,
	}
	taskSynopsis         string
	taskDescription      string
	taskStartAfterCreate bool
)

func init() {
	CreateCmd.Flags().StringVarP(&taskSynopsis, "synopsis", "s", "", "A short description of the task")
	CreateCmd.Flags().StringVarP(&taskDescription, "description", "d", "", "A long description of the task")
	CreateCmd.Flags().BoolVar(&taskStartAfterCreate, "start", false, "start the task after creating it")
}

func createTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("createTask")
	task := models.NewTask()
	if len(args) > 0 {
		task.Data().Synopsis = args[0]
	}
	if taskSynopsis != "" {
		task.Data().Synopsis = taskSynopsis
	}
	if len(args) > 1 {
		task.Data().Description = args[1]
	}
	if taskDescription != "" {
		task.Data().Description = taskDescription
	}
	err := task.Create()
	if err != nil {
		cli.PrintAndLogError(log, err, errors.CreateTaskError)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d", task.Data().ID), color.GreenString("created"))
	if taskStartAfterCreate {
		// TODO: Move this code to a common spot
		taskdisplay := strconv.Itoa(int(task.Data().ID))
		var stoppedTimesheet *models.TimesheetData
		stoppedTimesheet, err = task.StopRunningTask()
		if err != nil {
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
		timesheetData := new(models.TimesheetData)
		timesheetData.Task = *task.Data()
		timesheetData.StartTime = time.Now()
		err = models.Timesheet(timesheetData).Create()
		if err != nil {
			cli.PrintAndLogError(log, err, "%s for task %s", errors.CreateTimesheetError, taskdisplay)
			return err
		}
		fmt.Println(
			color.WhiteString("Task ID %d ", task.Data().ID),
			color.CyanString(task.Data().Synopsis),
			color.MagentaString("(%s) ", task.Data().Description),
			color.GreenString("started"),
			color.WhiteString("at %s", timesheetData.StartTime.Format(constants.TimestampLayout)),
		)
	}
	return nil
}
