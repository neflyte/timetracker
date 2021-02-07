package task

import (
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/spf13/cobra"
)

var (
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
	err := models.Task(new(models.TaskData)).StopRunningTask()
	if err != nil {
		utils.PrintAndLogError(errors.StopRunningTaskError, err, log)
		return err
	}
	return nil
}
