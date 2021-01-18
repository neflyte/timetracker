package task

import (
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
	StartCmd = &cobra.Command{
		Use:   "start [task id/synopsis]",
		Short: "Start a task",
		Args:  cobra.ExactArgs(1),
		RunE:  startTask,
	}
)

func startTask(_ *cobra.Command, args []string) error {
	var err error
	log := logger.GetLogger("startTask")
	taskid, tasksyn := utils.ResolveTask(args[0])
	taskdisplay := tasksyn
	if taskid > -1 {
		taskdisplay = strconv.Itoa(taskid)
	}
	task := new(models.Task)
	if taskid > -1 {
		err = database.DB.Find(&task, uint(taskid)).Error
	} else {
		err = database.DB.Where("synopsis = ?", tasksyn).First(&task).Error
	}
	if err != nil {
		fmt.Println(chalk.Red, "Error reading task", taskdisplay, chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error reading task %s", taskdisplay)
		return err
	}
	err = utils.StopRunningTask()
	if err != nil {
		fmt.Println(chalk.Red, "Error stopping running task:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("error stopping running task")
		return err
	}
	timesheet := new(models.Timesheet)
	timesheet.TaskID = task.ID
	timesheet.StartTime = time.Now()
	err = database.DB.Create(&timesheet).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error creating timesheet for task", taskdisplay, chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error creating timesheet for task %s", taskdisplay)
		return err
	}
	fmt.Println(
		chalk.White, chalk.Dim.TextStyle("Task"), taskdisplay,
		chalk.Blue, "started",
		chalk.White, chalk.Dim.TextStyle("at"),
		timesheet.StartTime.Format(`2006-01-02 15:04:05 PM`),
	)
	return nil
}
