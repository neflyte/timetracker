package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/neflyte/timetracker/internal/database"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TaskData is the main Task data structure
type TaskData struct {
	gorm.Model
	// Synopsis is a short title or identifier of the task
	Synopsis string `gorm:"uniqueindex"`
	// Description is a longer description of the task
	Description string

	// log is the struct logger
	log zerolog.Logger `gorm:"-"`
}

// NewTask creates a new TaskData structure and returns a Task interface to it
func NewTask() Task {
	return &TaskData{
		log: logger.GetStructLogger("TaskData"),
	}
}

// NewTaskWithData returns a new Task interfaced based on the supplied TaskData struct
func NewTaskWithData(data TaskData) Task {
	data.log = logger.GetStructLogger("TaskData")
	return &data
}

// TableName implements schema.Tabler
func (td *TaskData) TableName() string {
	return "task"
}

// Clone creates a clone of this TaskData object and returns the clone
func (td *TaskData) Clone() Task {
	clone := NewTask()
	// Clone GORM fields
	clone.Data().ID = td.ID
	clone.Data().CreatedAt = td.CreatedAt
	clone.Data().UpdatedAt = td.UpdatedAt
	clone.Data().DeletedAt = td.DeletedAt
	// Clone TaskData fields
	clone.Data().Synopsis = td.Synopsis
	clone.Data().Description = td.Description
	return clone
}

// TaskDatas is a helper type for a slice of TaskData structs
type TaskDatas []TaskData

// AsTaskList returns the slice of TaskData structs as a slice of Task interfaces
func (td TaskDatas) AsTaskList() TaskList {
	taskList := make(TaskList, len(td))
	for idx := range td {
		taskList[idx] = NewTaskWithData(td[idx])
	}
	return taskList
}

// Task is the main interface to task definitions
type Task interface {
	fmt.Stringer
	schema.Tabler
	Data() *TaskData
	Create() error
	Load(withDeleted bool) error
	Update(withDeleted bool) error
	Delete() error
	Clear()
	Clone() Task
	LoadAll(withDeleted bool) ([]TaskData, error)
	Search(text string) ([]TaskData, error)
	SearchBySynopsis(synopsis string) ([]TaskData, error)
	StopRunningTask() (*TimesheetData, error)
	FindTaskBySynopsis(tasks []TaskData, synopsis string) *TaskData
	Resolve(arg string) (uint, string)
}

// Data returns the underlying struct of the interface
func (td *TaskData) Data() *TaskData {
	return td
}

// String implements fmt.Stringer
func (td *TaskData) String() string {
	return fmt.Sprintf("%s (#%d)", td.Synopsis, td.ID)
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
	tx := database.Get().Begin()
	err := tx.Create(td).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
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
	tx := database.Get().Begin()
	err = tx.Delete(td).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
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

// SearchBySynopsis searches for a task by synopsis only using SQL equals
func (td *TaskData) SearchBySynopsis(synopsis string) ([]TaskData, error) {
	tasks := make([]TaskData, 0)
	err := database.Get().
		Model(new(TaskData)).
		Where("synopsis = ?", synopsis).
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
	} else if td.DeletedAt.Valid {
		return errors.New("cannot update a deleted task")
	}
	tx := db.Begin()
	err := tx.Save(td).Error
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// StopRunningTask stops the currently running task, if any
func (td *TaskData) StopRunningTask() (timesheetData *TimesheetData, err error) {
	log := logger.GetFuncLogger(td.log, "StopRunningTask")
	timesheets, err := NewTimesheet().SearchOpen()
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
	timesheet := NewTimesheetWithData(*timesheetData)
	err = timesheet.Update()
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
