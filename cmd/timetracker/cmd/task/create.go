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
	CreateCmd = &cobra.Command{
		Use:     "create [synopsis] [description]",
		Aliases: []string{"c"},
		Short:   "Create a task",
		Args:    cobra.MaximumNArgs(2),
		RunE:    createTask,
	}
	taskSynopsis    string
	taskDescription string
)

func init() {
	CreateCmd.Flags().StringVarP(&taskSynopsis, "synopsis", "s", "", "A short description of the task")
	CreateCmd.Flags().StringVarP(&taskDescription, "description", "d", "", "A long description of the task")
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
	fmt.Println(color.WhiteString("Task ID %d ", taskData.ID), color.GreenString("created"))
	return nil
}
