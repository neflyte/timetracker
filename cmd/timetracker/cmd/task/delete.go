package task

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/neflyte/timetracker/lib/ui/cli"
	"github.com/spf13/cobra"
)

var (
	// DeleteCmd represents the command to delete a task
	DeleteCmd = &cobra.Command{
		Use:     "delete [task id]",
		Aliases: []string{"d", "rm"},
		Short:   "Mark a task as deleted",
		Args:    cobra.ExactArgs(1),
		RunE:    deleteTask,
	}
)

func deleteTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("deleteTask")
	task := models.NewTask()
	task.Data().ID, task.Data().Synopsis = task.Resolve(args[0])
	// FIXME: check if the task is actually valid
	// FIXME: check if the task is running; abort with message that task is running unless `force` flag is used
	err := task.Delete()
	if err != nil {
		cli.PrintAndLogError(log, err, "%s; task=%#v", errors.DeleteTaskError, task.Data())
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d ", task.Data().ID), color.RedString("deleted"))
	return nil
}
