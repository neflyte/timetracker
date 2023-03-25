package cli

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
)

func StopRunningTimesheet() error {
	log := logger.GetLogger("StopRunningTimesheet")
	task := models.NewTask()
	stoppedTimesheet, err := task.StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		PrintAndLogError(log, err, tterrors.StopRunningTaskError)
		return err
	}
	if stoppedTimesheet != nil {
		log.Info().
			Msgf("task id %d (timesheet id %d) stopped\n", stoppedTimesheet.Task.ID, stoppedTimesheet.ID)
		fmt.Println(
			color.WhiteString("Task ID %d", stoppedTimesheet.Task.ID),
			color.YellowString("stopped"),
			color.WhiteString("at %s", stoppedTimesheet.StopTime.Time.Format(constants.TimestampLayout)),
			color.BlueString(stoppedTimesheet.StopTime.Time.Sub(stoppedTimesheet.StartTime).Truncate(time.Second).String()),
		)
	}
	return nil
}

func StartRunningTimesheet(task models.Task) error {
	log := logger.GetLogger("StartRunningTimesheet")
	taskdisplay := strconv.Itoa(int(task.Data().ID))
	timesheetData := new(models.TimesheetData)
	timesheetData.Task = *task.Data()
	timesheetData.StartTime = time.Now()
	err := models.Timesheet(timesheetData).Create()
	if err != nil {
		PrintAndLogError(log, err, "%s for task %s", tterrors.CreateTimesheetError, taskdisplay)
		return err
	}
	fmt.Println(
		color.WhiteString("Task ID %d ", task.Data().ID),
		color.CyanString(task.Data().Synopsis),
		color.MagentaString("(%s) ", task.Data().Description),
		color.GreenString("started"),
		color.WhiteString("at %s", timesheetData.StartTime.Format(constants.TimestampLayout)),
	)
	return nil
}
