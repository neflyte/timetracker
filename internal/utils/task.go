package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/ttacon/chalk"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func GetRunningTimesheet() (*models.Timesheet, error) {
	log := logger.GetLogger("GetRunningTask")
	timesheet := new(models.Timesheet)
	result := database.DB.Joins("Task").Where("stop_time IS NULL").First(&timesheet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Info().Msg("no running tasks")
			return nil, nil
		}
		fmt.Println(chalk.Red, "Error reading running task:", chalk.White, chalk.Dim.TextStyle(result.Error.Error()))
		log.Err(result.Error).Msg("error selecting running task")
		return nil, result.Error
	}
	log.Debug().Msgf("task id %d is running\n", timesheet.Task.ID)
	return timesheet, nil
}

func StopRunningTask() error {
	log := logger.GetLogger("StopRunningTask")
	timesheet, err := GetRunningTimesheet()
	if err != nil {
		fmt.Println(chalk.Red, "Error reading running task:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("error reading running task")
		return err
	}
	// No running tasks, return nil
	if timesheet == nil {
		return nil
	}
	stoptime := new(sql.NullTime)
	err = stoptime.Scan(time.Now())
	if err != nil {
		log.Err(err).Msg("error scanning time.Now() into sql.NullTime")
		return err
	}
	timesheet.StopTime = *stoptime
	err = database.DB.Save(&timesheet).Error
	if err != nil {
		log.Err(err).Msgf("error updating timesheet id %d for task id %d", timesheet.ID, timesheet.Task.ID)
		return err
	}
	log.Info().Msgf("task id %d (timesheet id %d) stopped\n", timesheet.Task.ID, timesheet.ID)
	fmt.Println(
		chalk.White, chalk.Dim.TextStyle("Task"), strconv.Itoa(int(timesheet.Task.ID)),
		chalk.Yellow, "stopped",
		chalk.White, chalk.Dim.TextStyle("at"),
		timesheet.StopTime.Time.Format(constants.TimestampLayout),
		chalk.Blue, timesheet.StopTime.Time.Sub(timesheet.StartTime).Truncate(time.Second).String(),
	)
	return nil
}

func ResolveTask(arg string) (taskid int, tasksynopsis string) {
	log := logger.GetLogger("ResolveTask")
	if arg == "" {
		return -1, ""
	}
	log.Trace().Msgf("arg=%s", arg)
	id, err := strconv.Atoi(arg)
	if err != nil {
		log.Trace().Msgf("error converting arg to number: %s; returning arg", err)
		return -1, arg
	}
	log.Trace().Msgf("returning %d", id)
	return id, ""
}
