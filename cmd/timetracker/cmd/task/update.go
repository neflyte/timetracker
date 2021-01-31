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
		fmt.Println(chalk.White, chalk.Dim.TextStyle("no updates specified; nothing to do"))
		log.Info().Msg("no updates specified; nothing to do")
		return nil
	}
	taskid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(chalk.Red, "Error converting argument", args[0], "into a number", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error converting argument (%s) into a number", args[0])
		return err
	}
	task := new(models.Task)
	// We use Unscoped() here since we might be undeleting a task
	err = database.DB.Unscoped().Find(&task, uint(taskid)).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error reading task", strconv.Itoa(taskid), chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error reading task id %d", taskid)
		return err
	}
	if task.DeletedAt.Valid {
		if updateUndelete {
			task.DeletedAt.Valid = false
			err = database.DB.Unscoped().Save(&task).Error
			if err != nil {
				fmt.Println(chalk.Red, "Error undeleting task", strconv.Itoa(taskid), chalk.White, chalk.Dim.TextStyle(err.Error()))
				log.Err(err).Msgf("error undeleting task id %d", taskid)
				return err
			}
			fmt.Println(chalk.White, chalk.Dim.TextStyle("Task ID"), strconv.Itoa(int(task.ID)), chalk.Green, "undeleted")
			log.Info().Msgf("task id %d undeleted", task.ID)
			return nil
		}
		err = errors.New("cannot update a deleted task")
		fmt.Println(chalk.Red, "Task ID", strconv.Itoa(int(task.ID)), "is deleted")
		log.Err(err).Msgf("task id %d is deleted", task.ID)
		return err
	}
	if updateSynopsis != "" {
		task.Synopsis = updateSynopsis
	}
	if updateDescription != "" {
		task.Description = updateDescription
	}
	err = database.DB.Save(&task).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error updating task", strconv.Itoa(taskid), chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error updating task id %d", taskid)
		return err
	}
	fmt.Println(chalk.White, chalk.Dim.TextStyle("Task ID"), strconv.Itoa(int(task.ID)), chalk.Green, "updated")
	return nil
}
