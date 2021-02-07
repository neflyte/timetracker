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
	taskData := new(models.TaskData)
	taskData.ID, _ = utils.ResolveTask(args[0])
	err := models.Task(taskData).Load(true)
	if err != nil {
		utils.PrintAndLogError(errors.LoadTaskError, err, log)
		return err
	}
	if taskData.DeletedAt.Valid {
		if updateUndelete {
			taskData.DeletedAt.Valid = false
			err = models.Task(taskData).Update(true)
			if err != nil {
				utils.PrintAndLogError(errors.UndeleteTaskError, err, log)
				return err
			}
			fmt.Println(color.WhiteString("Task ID %d", taskData.ID), color.GreenString("undeleted"))
			log.Info().Msgf("task id %d undeleted", taskData.ID)
			return nil
		}
		err = fmt.Errorf("task id %d is deleted", taskData.ID)
		utils.PrintAndLogError(errors.UpdateDeletedTaskError, err, log)
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
		utils.PrintAndLogError(errors.UpdateTaskError, err, log)
		return err
	}
	fmt.Println(color.WhiteString("Task ID %d", taskData.ID), color.GreenString("updated"))
	return nil
}
