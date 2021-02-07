package task

import (
	"database/sql"
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
	StopCmd = &cobra.Command{
		Use:     "stop [task id/synopsis]",
		Aliases: []string{"st"},
		Short:   "Stop a running task",
		Args:    cobra.MaximumNArgs(1),
		RunE:    stopTask,
	}
	stopRunningTask = false
)

func init() {
	StopCmd.Flags().BoolVarP(&stopRunningTask, "running", "r", false, "stop the running task, if any")
}

func stopTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("stopTask")
	// Stop the currently-running task (if any)
	if stopRunningTask {
		err := models.Task(new(models.TaskData)).StopRunningTask()
		if err != nil {
			utils.PrintAndLogError(errors.StopRunningTaskError, err, log)
			return err
		}
		return nil
	}
	// Stop the specified task if it exists
	if len(args) == 0 {
		// FIXME: No arg, can't do anything
		return nil
	}
	taskData := new(models.TaskData)
	taskData.ID, taskData.Synopsis = utils.ResolveTask(args[0])
	err := models.Task(taskData).Load(false)
	if err != nil {
		utils.PrintAndLogError(errors.LoadTaskError, err, log)
		return err
	}
	timesheetData := new(models.TimesheetData)
	// result := database.DB.Where("task_id = ? AND stop_time IS NULL",taskData.ID).Find(&timesheetData)
	openTimesheets, err := models.Timesheet(timesheetData).SearchOpen()
	if err != nil {
		utils.PrintAndLogError(errors.SearchOpenTimesheetsError, err, log)
		return err
	}
	// Sanity check: there should be at most one element in the openTimesheets slice
	if len(openTimesheets) > 1 {
		err = fmt.Errorf("%s", errors.TooManyOpenTimesheetsError)
		utils.PrintAndLogError("", err, log)
		return err
	}
	if len(openTimesheets) == 1 {
		timesheetData = &openTimesheets[0]
	}
	stoptime := new(sql.NullTime)
	err = stoptime.Scan(time.Now())
	if err != nil {
		utils.PrintAndLogError(errors.ScanNowIntoSQLNullTimeError, err, log)
		return err
	}
	timesheetData.StopTime = *stoptime
	err = models.Timesheet(timesheetData).Update()
	// err = database.DB.Save(&timesheetData).Error
	if err != nil {
		utils.PrintAndLogError(errors.UpdateTimesheetError, err, log)
		return err
	}
	fmt.Println(
		color.WhiteString("Task %d ", timesheetData.Task.ID),
		color.YellowString("stopped"),
		color.WhiteString(" at %s ", timesheetData.StopTime.Time.Format(constants.TimestampLayout)),
		color.BlueString(timesheetData.StopTime.Time.Sub(timesheetData.StartTime).Truncate(time.Second).String()),
	)
	return nil
}
