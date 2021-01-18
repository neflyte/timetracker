package task

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"strconv"
	"time"
)

var (
	StopCmd = &cobra.Command{
		Use:   "stop [task id/synopsis]",
		Short: "Stop a running task",
		Args:  cobra.ExactArgs(1),
		RunE:  stopTask,
	}
)

func stopTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("stopTask")
	taskid, tasksyn := utils.ResolveTask(args[0])
	timesheet := new(models.Timesheet)
	if taskid == -1 && tasksyn != "" {
		task := new(models.Task)
		err := database.DB.Where("synopsis = ?", tasksyn).Find(&task).Error
		if err != nil {
			fmt.Println(chalk.Red, "Error reading task", tasksyn, ":", chalk.White, chalk.Dim.TextStyle(err.Error()))
			log.Err(err).Msgf("error reading task %s", tasksyn)
			return err
		}
		taskid = int(task.ID)
	}
	if taskid == -1 && tasksyn == "" {
		err := errors.New("no task id or synopsis specified")
		fmt.Println(chalk.Red, "Unable to stop task:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("unable to stop task")
		return err
	}
	result := database.DB.Where("task_id = ? AND stop_time IS NULL", uint(taskid)).Find(&timesheet)
	if result.Error != nil {
		fmt.Println(chalk.Red, "Error looking for started task:", chalk.White, chalk.Dim.TextStyle(result.Error.Error()))
		log.Err(result.Error).Msg("error looking for started task")
		return result.Error
	}
	stoptime := new(sql.NullTime)
	err := stoptime.Scan(time.Now())
	if err != nil {
		fmt.Println(chalk.Red, "Error scanning time.Now() into sql.NullTime:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("error scanning time.Now() into sql.NullTime")
		return err
	}
	timesheet.StopTime = *stoptime
	err = database.DB.Save(&timesheet).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error stopping task", strconv.Itoa(int(timesheet.TaskID)), chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error stopping task id %d (timesheet id %d)", timesheet.TaskID, timesheet.ID)
		return err
	}
	fmt.Println(
		chalk.White, chalk.Dim.TextStyle("Task"), strconv.Itoa(int(timesheet.TaskID)),
		chalk.Yellow, "stopped",
		chalk.White, chalk.Dim.TextStyle("at"),
		timesheet.StopTime.Time.Format(`2006-01-02 15:04:05 PM`),
	)
	return nil
}
