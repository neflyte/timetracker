package task

import (
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
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
		fmt.Println(chalk.Red, "Invalid task ID:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("invalid task id")
		return err
	}
	task := new(models.Task)
	err = database.DB.Delete(task, uint(taskid)).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error deleting task ", taskid, ":", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error deleting task %d", taskid)
		return err
	}
	fmt.Println(chalk.White, chalk.Dim.TextStyle("Task ID"), strconv.Itoa(taskid), chalk.Red, "deleted")
	return nil
}
