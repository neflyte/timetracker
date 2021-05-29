package models

import (
	"database/sql"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/errors"
	"gorm.io/gorm"
	"time"
)

// TimesheetData is the main timesheet data structure
type TimesheetData struct {
	gorm.Model
	// Task is the task object linked to this Timesheet
	Task TaskData
	// TaskID is the database ID of the linked task object
	TaskID uint
	// StartTime is the time that the task was started at
	StartTime time.Time `gorm:"not null"`
	// StopTime is the time that the task was stopped at; if it is NULL, that means the task is still running
	StopTime sql.NullTime
}

// TableName implements schema.Tabler
func (tsd *TimesheetData) TableName() string {
	return "timesheet"
}

// Timesheet is the main timesheet function interface
type Timesheet interface {
	Create() error
	Load() error
	Delete() error
	LoadAll(withDeleted bool) ([]TimesheetData, error)
	SearchOpen() ([]TimesheetData, error)
	SearchDateRange() ([]TimesheetData, error)
	Update() error
	String() string
}

// String implements fmt.Stringer
func (tsd *TimesheetData) String() string {
	startTime := tsd.StartTime.String()
	stopTime := "(running)"
	if tsd.StopTime.Valid {
		stopTime = tsd.StopTime.Time.String()
	}
	return fmt.Sprintf(
		"TimesheetData[ Task=%s, StartTime=%s, StopTime=%s ]",
		tsd.Task.String(), startTime, stopTime,
	)
}

// Create creates a new timesheet record
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
	return database.Get().Create(tsd).Error
}

// Load attempts to load a timesheet by ID
func (tsd *TimesheetData) Load() error {
	if tsd.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: "cannot load a timesheet that does not exist",
		}
	}
	return database.Get().Joins("Task").First(tsd, tsd.ID).Error
}

// Delete marks a timesheet as deleted
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
	return database.Get().Delete(tsd).Error
}

// LoadAll loads all timesheet records, optionally including deleted timesheets
func (tsd *TimesheetData) LoadAll(withDeleted bool) ([]TimesheetData, error) {
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	timesheets := make([]TimesheetData, 0)
	err := db.Joins("Task").Find(&timesheets).Error
	return timesheets, err
}

// SearchOpen returns all open timesheets; there should be only one
func (tsd *TimesheetData) SearchOpen() ([]TimesheetData, error) {
	timesheets := make([]TimesheetData, 0)
	args := map[string]interface{}{
		"stop_time": nil,
	}
	if tsd.Task.ID > 0 {
		args["task_id"] = tsd.Task.ID
	}
	err := database.Get().Joins("Task").Where(args).Find(&timesheets).Error
	return timesheets, err
}

// SearchDateRange returns the timesheets that start on or after the StartTime and end on or before the StopTime
func (tsd *TimesheetData) SearchDateRange() ([]TimesheetData, error) {
	timesheets := make([]TimesheetData, 0)
	if tsd.StopTime.Valid {
		err := database.Get().
			Joins("Task").
			Where("start_time >= ? AND stop_time <= ?", tsd.StartTime, tsd.StopTime.Time).
			Find(&timesheets).
			Error
		return timesheets, err
	}
	err := database.Get().
		Joins("Task").
		Where("start_time >= ?", tsd.StartTime).
		Find(&timesheets).
		Error
	return timesheets, err
}

// Update attempts to update the timesheet record in the database
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
	return database.Get().Save(tsd).Error
}
