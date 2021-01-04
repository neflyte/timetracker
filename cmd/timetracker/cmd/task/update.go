package task

import (
	"errors"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	UpdateCmd = &cobra.Command{
		Use:   "update [task id]",
		Short: "Update task details",
		Args:  cobra.ExactArgs(1),
		RunE:  updateTask,
	}
	updateSynopsis    string
	updateDescription string
	updateUndelete    = false
)

func init() {
	UpdateCmd.Flags().StringVarP(&updateSynopsis, "synopsis", "s", "", "Update the task's short description")
	UpdateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "Update the task's log description")
	UpdateCmd.Flags().BoolVarP(&updateUndelete, "undelete", "u", false, "Undelete the task")
}

func updateTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("updateTask")
	if updateSynopsis == "" && updateDescription == "" && !updateUndelete {
		log.Println("no updates specified; nothing to do")
		return nil
	}
	taskid, err := strconv.Atoi(args[0])
	if err != nil {
		log.Printf("error converting argument (%s) into a number: %s\n", args[0], err)
		return err
	}
	task := new(models.Task)
	// We use Unscoped() here since we might be undeleting a task
	err = database.DB.Unscoped().Find(&task, uint(taskid)).Error
	if err != nil {
		log.Printf("error reading task id %d: %s\n", uint(taskid), err)
		return err
	}
	if task.DeletedAt.Valid {
		if updateUndelete {
			task.DeletedAt.Valid = false
			err = database.DB.Unscoped().Save(&task).Error
			if err != nil {
				log.Printf("error undeleting task %d: %s", task.ID, err)
				return err
			}
			log.Printf("task id %d undeleted", task.ID)
			return nil
		}
		log.Printf("task id %d is deleted; cannot update a deleted task", task.ID)
		return errors.New("cannot update a deleted task")
	}
	if updateSynopsis != "" {
		task.Synopsis = updateSynopsis
	}
	if updateDescription != "" {
		task.Description = updateDescription
	}
	err = database.DB.Save(&task).Error
	if err != nil {
		log.Printf("error updating task %d: %s", task.ID, err)
		return err
	}
	log.Printf("task id %d updated", task.ID)
	return nil
}
