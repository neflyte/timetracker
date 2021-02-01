package task

import (
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"strconv"
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
	task := new(models.Task)
	if len(args) == 0 && taskSynopsis == "" {
		// Need an argument!
		return errors.New("need at least a synopsis to create a new task")
	}
	if len(args) > 0 {
		task.Synopsis = args[0]
	}
	if taskSynopsis != "" {
		task.Synopsis = taskSynopsis
	}
	if len(args) > 1 {
		task.Description = args[1]
	}
	if taskDescription != "" {
		task.Description = taskDescription
	}
	err := database.DB.Create(task).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error creating new task:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("error creating new task")
		return err
	}
	fmt.Println(chalk.White, chalk.Dim.TextStyle("Task ID"), strconv.Itoa(int(task.ID)), chalk.Green, "created")
	return nil
}
