package task

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/spf13/cobra"
)

var (
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
	taskData := new(models.TaskData)
	taskData.ID, taskData.Synopsis = utils.ResolveTask(args[0])
	task := models.Task(taskData)
	err := task.Delete()
	if err != nil {
		utils.PrintAndLogError(fmt.Sprintf("%s; task=%#v", errors.DeleteTaskError, taskData), err, log)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d ", taskData.ID), color.RedString("deleted"))
	return nil
}
