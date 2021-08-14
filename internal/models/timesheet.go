package models

import (
	"database/sql"
	"fmt"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/database"
	ttErrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
	fmt.Stringer
	schema.Tabler
	Data() *TimesheetData
	Create() error
	Load() error
	Update() error
	Delete() error
	LoadAll(withDeleted bool) ([]TimesheetData, error)
	SearchOpen() ([]TimesheetData, error)
	SearchDateRange(withDeleted bool) ([]TimesheetData, error)
	LastStartedTasks(limit uint) (startedTasks []TaskData, err error)
	TaskReport(startDate, endDate time.Time, withDeleted bool) (reportData TaskReport, err error)
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
		"TimesheetData{Task=%s, StartTime=%s, StopTime=%s}",
		tsd.Task.String(), startTime, stopTime,
	)
}

// Create creates a new timesheet record
func (tsd *TimesheetData) Create() error {
	if tsd.ID != 0 {
		return ttErrors.ErrInvalidTimesheetState{
			Details: ttErrors.OverwriteTimesheetByCreateError,
		}
	}
	if tsd.Task.ID == 0 {
		return ttErrors.ErrInvalidTimesheetState{
			Details: ttErrors.TimesheetWithoutTaskError,
		}
	}
	tx := database.Get().Begin()
	err := tx.Create(tsd).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// Load attempts to load a timesheet by ID
func (tsd *TimesheetData) Load() error {
	if tsd.ID == 0 {
		return ttErrors.ErrInvalidTimesheetState{
			Details: ttErrors.LoadInvalidTimesheetError,
		}
	}
	return database.Get().
		Joins("Task").
		First(tsd, tsd.ID).
		Error
}

// Delete marks a timesheet as deleted
func (tsd *TimesheetData) Delete() error {
	if tsd.ID == 0 {
		return ttErrors.ErrInvalidTimesheetState{
			Details: ttErrors.DeleteInvalidTimesheetError,
		}
	}
	err := tsd.Load()
	if err != nil {
		return err
	}
	return database.Get().
		Delete(tsd).
		Error
}

// LoadAll loads all timesheet records, optionally including deleted timesheets
func (tsd *TimesheetData) LoadAll(withDeleted bool) ([]TimesheetData, error) {
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	timesheets := make([]TimesheetData, 0)
	err := db.Joins("Task").
		Find(&timesheets).
		Error
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
	err := database.Get().
		Joins("Task").
		Where(args).
		Find(&timesheets).
		Error
	return timesheets, err
}

// SearchDateRange returns the timesheets that start on or after the StartTime and end on or before the StopTime
func (tsd *TimesheetData) SearchDateRange(withDeleted bool) ([]TimesheetData, error) {
	timesheets := make([]TimesheetData, 0)
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	if tsd.StopTime.Valid {
		err := db.Joins("Task").
			Where("start_time >= ? AND stop_time <= ?", tsd.StartTime, tsd.StopTime.Time).
			Find(&timesheets).
			Error
		return timesheets, err
	}
	err := db.Joins("Task").
		Where("start_time >= ?", tsd.StartTime).
		Find(&timesheets).
		Error
	return timesheets, err
}

// Update attempts to update the timesheet record in the database
func (tsd *TimesheetData) Update() error {
	if tsd.ID == 0 {
		return ttErrors.ErrInvalidTimesheetState{
			Details: ttErrors.UpdateInvalidTimesheetError,
		}
	}
	if tsd.Task.ID == 0 {
		return ttErrors.ErrInvalidTimesheetState{
			Details: ttErrors.TimesheetWithoutTaskError,
		}
	}
	tx := database.Get().Begin()
	err := tx.Save(tsd).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

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

// findTaskByID attempts to fina a task with the specified ID in the specified slice of tasks
func findTaskByID(tasks []TaskData, id uint) *TaskData {
	for _, task := range tasks {
		if task.ID == id {
			return &task
		}
	}
	return nil
}

// TaskReportData is a struct that contains a single entry of a Task Report
type TaskReportData struct {
	TaskID          uint
	TaskSynopsis    string
	TaskDescription string
	StartDate       sql.NullTime
	DurationSeconds int
}

// TaskReport is a type alias for a slice of TaskReportData structs
type TaskReport []TaskReportData

// NewTaskReportData returns a pointer to a new instance of the TaskReportData struct
func NewTaskReportData() *TaskReportData {
	return &TaskReportData{
		StartDate: sql.NullTime{},
	}
}

// String implements fmt.Stringer
func (trd *TaskReportData) String() string {
	trdDuration := time.Second * time.Duration(trd.DurationSeconds)
	return fmt.Sprintf(
		"[%d] %s: %s; %s -> %s",
		trd.TaskID,
		trd.TaskSynopsis,
		trd.TaskDescription,
		trd.StartDate.Time.Format(constants.TimestampDateLayout),
		trdDuration.String(),
	)
}

// Duration returns the DurationSeconds property as a time.Duration
func (trd *TaskReportData) Duration() time.Duration {
	return time.Second * time.Duration(trd.DurationSeconds)
}

func (trd *TaskReportData) Clone() TaskReportData {
	startDate := sql.NullTime{}
	_ = startDate.Scan(trd.StartDate)
	return TaskReportData{
		TaskID:          trd.TaskID,
		TaskSynopsis:    trd.TaskSynopsis,
		TaskDescription: trd.TaskDescription,
		StartDate:       startDate,
		DurationSeconds: trd.DurationSeconds,
	}
}

func (tr TaskReport) Clone() TaskReport {
	taskReport := make(TaskReport, len(tr))
	for _, taskReportData := range tr {
		taskReport = append(taskReport, taskReportData.Clone())
	}
	return taskReport
}

// TaskReport returns a list of tasks and their aggregated durations between the two supplied dates
func (tsd *TimesheetData) TaskReport(startDate, endDate time.Time, withDeleted bool) (reportData TaskReport, err error) {
	var rows *sql.Rows

	reportData = make([]TaskReportData, 0)
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	// TODO: Move the SQL statement to an appropriate constant
	rows, err = db.Raw(
		`SELECT 
	ts.task_id AS task_id,
	t.synopsis AS task_synopsis,
	t.description AS task_description,
	ts.start_time AS start_date,
	STRFTIME('%s', ts.stop_time) - STRFTIME('%s', ts.start_time) AS duration_seconds 
FROM timesheet ts JOIN task t ON ts.task_id = t.id 
WHERE ts.start_time >= ? AND ts.stop_time <= ? AND ts.stop_time IS NOT NULL 
GROUP BY ts.task_id, DATE(ts.start_time) 
ORDER BY DATE(ts.start_time) ASC`,
		startDate,
		endDate).
		Rows()
	if err != nil {
		return nil, err
	}
	defer database.CloseRows(rows)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("error moving to next row of results %w", rows.Err())
		}
		taskReportData := NewTaskReportData()
		err = database.Get().ScanRows(rows, taskReportData)
		if err != nil {
			return nil, err
		}
		reportData = append(reportData, *taskReportData)
	}
	return
}
