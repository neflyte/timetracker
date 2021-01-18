package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/ttacon/chalk"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func StopRunningTask() error {
	log := logger.GetLogger("StopRunningTask")
	timesheet := new(models.Timesheet)
	result := database.DB.Where("stop_time IS NULL").First(&timesheet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Info().Msg("no running tasks")
			return nil
		}
		fmt.Println(chalk.Red, "Error reading running task:", chalk.White, chalk.Dim.TextStyle(result.Error.Error()))
		log.Err(result.Error).Msg("error selecting running task")
		return result.Error
	}
	log.Debug().Msgf("task id %d is running\n", timesheet.TaskID)
	stoptime := new(sql.NullTime)
	err := stoptime.Scan(time.Now())
	if err != nil {
		log.Err(err).Msg("error scanning time.Now() into sql.NullTime")
		return err
	}
	timesheet.StopTime = *stoptime
	err = database.DB.Save(&timesheet).Error
	if err != nil {
		log.Err(err).Msgf("error updating timesheet id %d for task id %d", timesheet.ID, timesheet.TaskID)
		return err
	}
	log.Info().Msgf("task id %d (timesheet id %d) stopped\n", timesheet.TaskID, timesheet.ID)
	fmt.Println(
		chalk.White, chalk.Dim.TextStyle("Task"), strconv.Itoa(int(timesheet.TaskID)),
		chalk.Yellow, "stopped",
		chalk.White, chalk.Dim.TextStyle("at"),
		timesheet.StopTime.Time.Format(`2006-01-02 15:04:05 PM`),
	)
	return nil
}

func ResolveTask(arg string) (taskid int, tasksynopsis string) {
	taskid = -1
	if arg == "" {
		return
	}
	taskid, err := strconv.Atoi(arg)
	if err == nil {
		return
	}
	return taskid, arg
}
