package task

import (
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	DeleteCmd = &cobra.Command{
		Use:   "delete [task id]",
		Short: "Mark a task as deleted",
		Args:  cobra.ExactArgs(1),
		RunE:  deleteTask,
	}
)

func deleteTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("deleteTask")
	taskid, err := strconv.Atoi(args[0])
	if err != nil {
		log.Printf("error converting argument (%s) into a number: %s\n", args[0], err)
		return err
	}
	task := new(models.Task)
	err = database.DB.Delete(task, uint(taskid)).Error
	if err != nil {
		log.Printf("error deleting task id %d: %s\n", uint(taskid), err)
		return err
	}
	log.Printf("task id %d deleted", uint(taskid))
	return nil
}
