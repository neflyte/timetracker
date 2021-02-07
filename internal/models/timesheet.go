package models

import (
	"database/sql"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/errors"
	"gorm.io/gorm"
	"time"
)

type TimesheetData struct {
	gorm.Model
	Task      TaskData
	TaskID    uint
	StartTime time.Time `gorm:"not null"`
	StopTime  sql.NullTime
}

func (tsd *TimesheetData) TableName() string {
	return "timesheet"
}

type Timesheet interface {
	Create() error
	Load() error
	Delete() error
	LoadAll(withDeleted bool) ([]TimesheetData, error)
	SearchOpen() ([]TimesheetData, error)
	SearchDateRange() ([]TimesheetData, error)
	Update() error
}

func (tsd *TimesheetData) Create() error {
	if tsd.ID != 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "cannot overwrite a timesheet by creating it",
		}
	}
	if tsd.Task.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "no task is associated with the timesheet",
		}
	}
	return database.DB.Create(tsd).Error
}

func (tsd *TimesheetData) Load() error {
	if tsd.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "cannot load a timesheet that does not exist",
		}
	}
	return database.DB.Joins("Task").First(tsd, tsd.ID).Error
}

func (tsd *TimesheetData) Delete() error {
	if tsd.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "cannot delete a timesheet that does not exist",
		}
	}
	err := tsd.Load()
	if err != nil {
		return err
	}
	return database.DB.Delete(tsd).Error
}

func (tsd *TimesheetData) LoadAll(withDeleted bool) ([]TimesheetData, error) {
	db := database.DB
	if withDeleted {
		db = db.Unscoped()
	}
	timesheets := make([]TimesheetData, 0)
	err := db.Joins("Task").Find(&timesheets).Error
	return timesheets, err
}

func (tsd *TimesheetData) SearchOpen() ([]TimesheetData, error) {
	timesheets := make([]TimesheetData, 0)
	args := map[string]interface{}{
		"stop_time": nil,
	}
	if tsd.Task.ID > 0 {
		args["task_id"] = tsd.Task.ID
	}
	err := database.DB.Joins("Task").Where(args).Find(&timesheets).Error
	return timesheets, err
}

func (tsd *TimesheetData) SearchDateRange() ([]TimesheetData, error) {
	timesheets := make([]TimesheetData, 0)
	if tsd.StopTime.Valid {
		err := database.DB.
			Joins("Task").
			Where("start_time >= ? AND stop_time <= ?", tsd.StartTime, tsd.StopTime.Time).
			Find(&timesheets).
			Error
		return timesheets, err
	}
	err := database.DB.
		Joins("Task").
		Where("start_time >= ?", tsd.StartTime).
		Find(&timesheets).
		Error
	return timesheets, err
}

func (tsd *TimesheetData) Update() error {
	if tsd.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "cannot update a timesheet that does not exist",
		}
	}
	if tsd.Task.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "no task is associated with the timesheet",
		}
	}
	return database.DB.Save(tsd).Error
}
