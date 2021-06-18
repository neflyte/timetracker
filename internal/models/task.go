package models

import (
	"database/sql"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// TaskData is the main Task data strucure
type TaskData struct {
	gorm.Model
	// Synopsis is a short title or identifier of the task
	Synopsis string `gorm:"uniqueindex"`
	// Description is a longer description of the task
	Description string

	log zerolog.Logger `gorm:"-"`
}

// NewTaskData creates a new TaskData structure and returns a pointer to it
func NewTaskData() *TaskData {
	return &TaskData{
		log: logger.GetStructLogger("TaskData"),
	}
}

// TableName implements schema.Tabler
func (td *TaskData) TableName() string {
	return "task"
}

// Clone creates a clone of this TaskData object and returns a pointer to it
func (td *TaskData) Clone() *TaskData {
	clone := NewTaskData()
	// Clone GORM fields
	clone.ID = td.ID
	clone.CreatedAt = td.CreatedAt
	clone.UpdatedAt = td.UpdatedAt
	clone.DeletedAt = td.DeletedAt
	// Clone TaskData fields
	clone.Synopsis = td.Synopsis
	clone.Description = td.Description
	return clone
}

// Task is the main interface to task definitions
type Task interface {
	Create() error
	Load(withDeleted bool) error
	Delete() error
	LoadAll(withDeleted bool) ([]TaskData, error)
	Search(text string) ([]TaskData, error)
	Update(withDeleted bool) error
	StopRunningTask() (*TimesheetData, error)
	Clear()
	String() string
	FindTaskBySynopsis(tasks []TaskData, synopsis string) *TaskData
	Resolve(arg string) (uint, string)
}

// String implements fmt.Stringer
func (td *TaskData) String() string {
	return fmt.Sprintf("[%d] %s: %s", td.ID, td.Synopsis, td.Description)
}

// Create creates a new task
func (td *TaskData) Create() error {
	if td.ID != 0 {
		return tterrors.ErrInvalidTaskState{
			Details: tterrors.OverwriteTaskByCreateError,
		}
	}
	if td.Synopsis == "" {
		return tterrors.ErrInvalidTaskState{
			Details: tterrors.EmptySynopsisTaskError,
		}
	}
	return database.Get().Create(td).Error
}

// Load attempts to load the task specified by ID or Synopsis
func (td *TaskData) Load(withDeleted bool) error {
	if td.ID == 0 && td.Synopsis == "" {
		return tterrors.ErrInvalidTaskState{
			Details: tterrors.LoadInvalidTaskError,
		}
	}
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	if td.ID > 0 {
		return db.First(td, td.ID).Error
	}
	return db.Where("synopsis = ?", td.Synopsis).First(td).Error
}

// Delete marks the task as deleted
func (td *TaskData) Delete() error {
	if td.ID == 0 {
		return tterrors.ErrInvalidTaskState{
			Details: tterrors.DeleteInvalidTaskError,
		}
	}
	err := td.Load(false)
	if err != nil {
		return err
	}
	return database.Get().Delete(td).Error
}

// LoadAll loads all tasks in the database, optionally including deleted tasks
func (td *TaskData) LoadAll(withDeleted bool) ([]TaskData, error) {
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	tasks := make([]TaskData, 0)
	err := db.Find(&tasks).Error
	return tasks, err
}

// Search searches for a task by synopsis or description using SQL LIKE
func (td *TaskData) Search(text string) ([]TaskData, error) {
	tasks := make([]TaskData, 0)
	err := database.Get().
		Model(new(TaskData)).
		Where("synopsis LIKE ? OR description LIKE ?", text, text).
		Find(&tasks).
		Error
	return tasks, err
}

// Update writes task changes to the database
func (td *TaskData) Update(withDeleted bool) error {
	if td.ID == 0 {
		return tterrors.ErrInvalidTaskState{
			Details: tterrors.UpdateInvalidTaskError,
		}
	}
	if td.Synopsis == "" {
		return tterrors.ErrInvalidTaskState{
			Details: tterrors.UpdateEmptySynopsisTaskError,
		}
	}
	db := database.Get()
	if withDeleted {
		db = db.Unscoped()
	}
	return db.Save(td).Error
}

// StopRunningTask stops the currently running task, if any
func (td *TaskData) StopRunningTask() (timesheetData *TimesheetData, err error) {
	log := logger.GetFuncLogger(td.log, "StopRunningTask")
	timesheets, err := Timesheet(new(TimesheetData)).SearchOpen()
	if err != nil {
		log.Err(err).Msg("error searching for open timesheets")
		return nil, err
	}
	// No running tasks, return nil
	if len(timesheets) == 0 {
		return nil, tterrors.ErrNoRunningTask{}
	}
	timesheetData = &timesheets[0]
	stoptime := new(sql.NullTime)
	err = stoptime.Scan(time.Now())
	if err != nil {
		log.Err(err).Msg(tterrors.ScanNowIntoSQLNullTimeError)
		return nil, tterrors.ErrScanNowIntoSQLNull{Wrapped: err}
	}
	timesheetData.StopTime = *stoptime
	err = Timesheet(timesheetData).Update()
	if err != nil {
		log.Err(err).Msg("error updating running timesheet")
		return nil, err
	}
	return
}

// Clear resets the state of this object to the default, newly-initialized state
func (td *TaskData) Clear() {
	td.ID = 0
	td.Synopsis = ""
	td.Description = ""
	td.CreatedAt = time.Now()
	td.DeletedAt.Time = time.Now()
	td.DeletedAt.Valid = false
	td.UpdatedAt = time.Now()
}

// FindTaskBySynopsis returns a task with a matching synopsis from a slice of tasks
func (td *TaskData) FindTaskBySynopsis(tasks []TaskData, synopsis string) *TaskData {
	for _, task := range tasks {
		if task.Synopsis == synopsis {
			return &task
		}
	}
	return nil
}

// Resolve takes a string argument and produces either a taskid (uint) or a synopsis (string)
func (td *TaskData) Resolve(arg string) (taskid uint, tasksynopsis string) {
	log := logger.GetFuncLogger(td.log, "Resolve")
	if arg == "" {
		return 0, ""
	}
	log.Trace().Msgf("arg=%s", arg)
	id, err := strconv.Atoi(arg)
	if err != nil {
		log.Trace().Msgf("error converting arg to number: %s; returning arg", err)
		return 0, arg
	}
	log.Trace().Msgf("returning %d", uint(id))
	return uint(id), ""
}
