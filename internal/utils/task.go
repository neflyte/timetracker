package utils

import (
	"database/sql"
	"errors"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"gorm.io/gorm"
	"time"
)

func StopRunningTask() error {
	log := logger.GetLogger("StopRunningTask")
	timesheet := new(models.Timesheet)
	result := database.DB.Where("stop_time = ?", nil).First(&timesheet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println("no running tasks")
			return nil
		}
		log.Printf("error selecting running task: %s\n", result.Error)
		return result.Error
	}
	log.Printf("task id %d is running\n", timesheet.TaskID)
	stoptime := new(sql.NullTime)
	err := stoptime.Scan(time.Now())
	if err != nil {
		log.Printf("error scanning time.Now() into sql.NullTime: %s\n", err)
		return err
	}
	timesheet.StopTime = *stoptime
	err = database.DB.Save(&timesheet).Error
	if err != nil {
		log.Printf("error updating timesheet id %d for task id %d: %s\n", timesheet.ID, timesheet.TaskID, err)
		return err
	}
	log.Printf("task id %d (timesheet id %d) stopped\n", timesheet.TaskID, timesheet.ID)
	return nil
}
