package task

import (
	"fmt"
	"github.com/neflyte/timetracker/internal/constants"
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
		Use:     "start [task id/synopsis]",
		Aliases: []string{"s"},
		Short:   "Start a task",
		Args:    cobra.ExactArgs(1),
		RunE:    startTask,
	}
)

func startTask(_ *cobra.Command, args []string) (err error) {
	log := logger.GetLogger("startTask")
	taskid, tasksyn := utils.ResolveTask(args[0])
	taskdisplay := tasksyn
	if taskid > -1 {
		taskdisplay = strconv.Itoa(taskid)
	}
	log.Debug().Msgf("taskdisplay=%s", taskdisplay)
	task := new(models.Task)
	if taskid > -1 {
		err = database.DB.Find(&task, uint(taskid)).Error
	} else {
		err = database.DB.Where(&models.Task{Synopsis: tasksyn}).First(&task).Error
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
	timesheet := &models.Timesheet{Task: *task, StartTime: time.Now()}
	err = database.DB.Create(&timesheet).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error creating timesheet for task", taskdisplay, chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msgf("error creating timesheet for task %s", taskdisplay)
		return err
	}
	fmt.Println(
		chalk.White, "Task", taskdisplay,
		chalk.Cyan, task.Synopsis,
		chalk.Magenta, chalk.Dim.TextStyle("("+task.Description+")"),
		chalk.Blue, "started",
		chalk.White, "at",
		timesheet.StartTime.Format(constants.TimestampLayout),
	)
	return nil
}
