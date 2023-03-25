package task

import (
	"fmt"

	"github.com/fatih/color"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
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
	task := models.NewTask()
	if len(args) > 0 {
		task.Data().Synopsis = args[0]
	}
	if taskSynopsis != "" {
		task.Data().Synopsis = taskSynopsis
	}
	if len(args) > 1 {
		task.Data().Description = args[1]
	}
	if taskDescription != "" {
		task.Data().Description = taskDescription
	}
	err := task.Create()
	if err != nil {
		cli.PrintAndLogError(log, err, tterrors.CreateTaskError)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d", task.Data().ID), color.GreenString("created")) // i18n
	if taskStartAfterCreate {
		err = cli.StopRunningTimesheet()
		if err != nil {
			return err
		}
		err = cli.StartRunningTimesheet(task)
		if err != nil {
			return err
		}
	}
	return nil
}
