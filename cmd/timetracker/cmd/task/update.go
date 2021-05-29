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
	// UpdateCmd represents the command that updates an existing task
	UpdateCmd = &cobra.Command{
		Use:     "update [task id]",
		Aliases: []string{"u"},
		Short:   "Update task details",
		Args:    cobra.ExactArgs(1),
		RunE:    updateTask,
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
		fmt.Println(color.WhiteString("no updates specified; nothing to do"))
		log.Info().Msg("no updates specified; nothing to do")
		return nil
	}
	taskData := models.NewTaskData()
	taskData.ID, _ = taskData.Resolve(args[0])
	err := models.Task(taskData).Load(true)
	if err != nil {
		cli.PrintAndLogError(log, err, errors.LoadTaskError)
		return err
	}
	if taskData.DeletedAt.Valid {
		if updateUndelete {
			taskData.DeletedAt.Valid = false
			err = models.Task(taskData).Update(true)
			if err != nil {
				cli.PrintAndLogError(log, err, errors.UndeleteTaskError)
				return err
			}
			fmt.Println(color.WhiteString("Task ID %d", taskData.ID), color.GreenString("undeleted"))
			log.Info().Msgf("task id %d undeleted", taskData.ID)
			return nil
		}
		err = fmt.Errorf("task id %d is deleted", taskData.ID)
		cli.PrintAndLogError(log, err, errors.UpdateDeletedTaskError)
		return err
	}
	if updateSynopsis != "" {
		taskData.Synopsis = updateSynopsis
	}
	if updateDescription != "" {
		taskData.Description = updateDescription
	}
	err = models.Task(taskData).Update(false)
	if err != nil {
		cli.PrintAndLogError(log, err, errors.UpdateTaskError)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d", taskData.ID), color.GreenString("updated"))
	return nil
}
