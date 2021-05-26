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
	taskData := new(models.TaskData)
	if len(args) > 0 {
		taskData.Synopsis = args[0]
	}
	if taskSynopsis != "" {
		taskData.Synopsis = taskSynopsis
	}
	if len(args) > 1 {
		taskData.Description = args[1]
	}
	if taskDescription != "" {
		taskData.Description = taskDescription
	}
	task := models.Task(taskData)
	err := task.Create()
	if err != nil {
		utils.PrintAndLogError(errors.CreateTaskError, err, log)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d", taskData.ID), color.GreenString("created"))
	if taskStartAfterCreate {
		// TODO: Move this code to a common spot
		taskdisplay := strconv.Itoa(int(taskData.ID))
		err = task.StopRunningTask()
		if err != nil {
			utils.PrintAndLogError(errors.StopRunningTaskError, err, log)
			return err
		}
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
	}
	return nil
}
