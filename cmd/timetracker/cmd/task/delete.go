package task

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
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
	taskData := models.NewTaskData()
	taskData.ID, taskData.Synopsis = taskData.Resolve(args[0])
	err := models.Task(taskData).Delete()
	if err != nil {
		// fmt.Sprintf("%s; task=%#v", errors.DeleteTaskError, taskData)
		cli.PrintAndLogError(log, err, "%s; task=%#v", errors.DeleteTaskError, taskData)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d ", taskData.ID), color.RedString("deleted"))
	return nil
}
