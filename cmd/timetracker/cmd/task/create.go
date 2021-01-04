package task

import (
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
)

var (
	CreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		RunE:  createTask,
	}
	taskSynopsis    string
	taskDescription string
)

func init() {
	CreateCmd.Flags().StringVarP(&taskSynopsis, "synopsis", "s", "", "A short description of the task")
	CreateCmd.Flags().StringVarP(&taskDescription, "description", "d", "", "A long description of the task")
	_ = CreateCmd.MarkFlagRequired("synopsis")
}

func createTask(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("createTask")
	task := new(models.Task)
	task.Synopsis = taskSynopsis
	task.Description = taskDescription
	err := database.DB.Create(task).Error
	if err != nil {
		log.Printf("error creating new task: %s\n", err)
		return err
	}
	log.Printf("created task ID %d\n", task.ID)
	return nil
}
