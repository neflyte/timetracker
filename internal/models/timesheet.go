package models

import (
	"database/sql"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"time"
)

const (
	lastStartedTasksDefaultLimit = 5
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

	// log is the struct logger
	log zerolog.Logger `gorm:"-"`
}

// NewTimesheet returns an new, initialized Timesheet interface
func NewTimesheet() Timesheet {
	return &TimesheetData{
		Task:     TaskData{},
		StopTime: sql.NullTime{},
		log:      logger.GetStructLogger("TimesheetData"),
	}
}

// NewTimesheetWithData returns a new Timesheet interface based on the supplied TimesheetData struct
func NewTimesheetWithData(data TimesheetData) Timesheet {
	data.log = logger.GetStructLogger("TimesheetData")
	return &data
}

// TableName implements schema.Tabler
func (tsd *TimesheetData) TableName() string {
	return "timesheet"
}

// Timesheet is the main timesheet function interface
type Timesheet interface {
	Data() *TimesheetData
	Create() error
	Load() error
	Delete() error
	LoadAll(withDeleted bool) ([]TimesheetData, error)
	SearchOpen() ([]TimesheetData, error)
	SearchDateRange() ([]TimesheetData, error)
	Update() error
	String() string
	LastStartedTasks(limit uint) (startedTasks []TaskData, err error)
}

// Data returns the struct underlying the interface
func (tsd *TimesheetData) Data() *TimesheetData {
	return tsd
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
			Details: errors.OverwriteTimesheetByCreateError,
		}
	}
	if tsd.Task.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: errors.TimesheetWithoutTaskError,
		}
	}
	return database.Get().Create(tsd).Error
}

// Load attempts to load a timesheet by ID
func (tsd *TimesheetData) Load() error {
	if tsd.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: errors.LoadInvalidTimesheetError,
		}
	}
	return database.Get().Joins("Task").First(tsd, tsd.ID).Error
}

// Delete marks a timesheet as deleted
func (tsd *TimesheetData) Delete() error {
	if tsd.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: errors.DeleteInvalidTimesheetError,
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
			Details: errors.UpdateInvalidTimesheetError,
		}
	}
	if tsd.Task.ID == 0 {
		return errors.ErrInvalidTimesheetState{
			Details: errors.TimesheetWithoutTaskError,
		}
	}
	return database.Get().Save(tsd).Error
}

/*
SELECT task_id
FROM (
     SELECT task_id, start_time
     FROM timesheet
     ORDER BY start_time DESC
)
GROUP BY task_id
ORDER BY start_time DESC
LIMIT 5;
*/

// LastStartedTasks returns a list of most-recently started tasks. The size of the list is limited
// by the limit parameter. If a limit of zero is specified, the default value is used.
func (tsd *TimesheetData) LastStartedTasks(limit uint) (startedTasks []TaskData, err error) {
	startedTasks = make([]TaskData, 0)
	taskLimit := uint(lastStartedTasksDefaultLimit)
	if limit > 0 {
		taskLimit = limit
	}
	taskIDs := make([]uint, 0)
	subquery := database.Get().
		Model(tsd).
		Select("task_id", "start_time").
		Order("start_time DESC")
	err = database.Get().
		Table("(?) as data", subquery).
		Select("task_id").
		Group("task_id").
		Order("start_time DESC").
		Limit(int(taskLimit)).
		Find(&taskIDs).
		Error
	if err != nil {
		return
	}
	unorderedStartedTasks := make([]TaskData, 0)
	err = database.Get().
		Model(new(TaskData)).
		Find(&unorderedStartedTasks, taskIDs).
		Error
	if err != nil {
		return
	}
	for _, taskID := range taskIDs {
		task := findTaskByID(unorderedStartedTasks, taskID)
		if task != nil {
			startedTasks = append(startedTasks, *task)
		}
	}
	return
}

func findTaskByID(tasks []TaskData, id uint) *TaskData {
	for _, task := range tasks {
		if task.ID == id {
			return &task
		}
	}
	return nil
}
