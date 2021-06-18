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
	// StopCmd represents the command that stops the running task
	StopCmd = &cobra.Command{
		Use:     "stop",
		Aliases: []string{"st"},
		Short:   "Stop the running task",
		Args:    cobra.ExactArgs(0),
		RunE:    stopTask,
	}
)

func stopTask(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("stopTask")
	timesheetData, err := models.Task(models.NewTaskData()).StopRunningTask()
	if err != nil {
		cli.PrintAndLogError(log, err, errors.StopRunningTaskError)
		return err
	}
	log.Info().Msgf("task id %d (timesheet id %d) stopped\n", timesheetData.Task.ID, timesheetData.ID)
	fmt.Println(
		color.WhiteString("Task ID %d", timesheetData.Task.ID),
		color.YellowString("stopped"),
		color.WhiteString("at %s", timesheetData.StopTime.Time.Format(constants.TimestampLayout)),
		color.BlueString(timesheetData.StopTime.Time.Sub(timesheetData.StartTime).Truncate(time.Second).String()),
	)
	return nil
}
