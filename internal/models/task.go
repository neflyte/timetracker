package models

import (
	"database/sql"
	"fmt"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/utils"
	"gorm.io/gorm"
	"time"
)

type TaskData struct {
	gorm.Model
	Synopsis    string `gorm:"uniqueindex"`
	Description string
}

func (td *TaskData) TableName() string {
	return "task"
}

type Task interface {
	Create() error
	Load(withDeleted bool) error
	Delete() error
	LoadAll(withDeleted bool) ([]TaskData, error)
	Search(text string) ([]TaskData, error)
	Update(withDeleted bool) error
	StopRunningTask() error
	Clear()
}

func (td *TaskData) Create() error {
	if td.ID != 0 {
		return errors.ErrInvalidTaskState{
			Details: "cannot overwrite a task by creating it",
		}
	}
	if td.Synopsis == "" {
		return errors.ErrInvalidTaskState{
			Details: "cannot create a task with an empty synopsis",
		}
	}
	return database.DB.Create(td).Error
}

func (td *TaskData) Load(withDeleted bool) error {
	if td.ID <= 0 && td.Synopsis == "" {
		return errors.ErrInvalidTaskState{
			Details: "cannot load a task that does not exist",
		}
	}
	db := database.DB
	if withDeleted {
		db = db.Unscoped()
	}
	return db.First(td).Error
}

func (td *TaskData) Delete() error {
	if td.ID <= 0 && td.Synopsis == "" {
		return errors.ErrInvalidTaskState{
			Details: "cannot delete a task that does not exist",
		}
	}
	err := td.Load(false)
	if err != nil {
		return err
	}
	return database.DB.Delete(td).Error
}

func (td *TaskData) LoadAll(withDeleted bool) ([]TaskData, error) {
	db := database.DB
	if withDeleted {
		db = db.Unscoped()
	}
	tasks := make([]TaskData, 0)
	err := db.Find(&tasks).Error
	return tasks, err
}

func (td *TaskData) Search(text string) ([]TaskData, error) {
	tasks := make([]TaskData, 0)
	err := database.DB.
		Model(new(TaskData)).
		Where("synopsis LIKE ? OR description LIKE ?", text, text).
		Find(&tasks).
		Error
	return tasks, err
}

func (td *TaskData) Update(withDeleted bool) error {
	if td.ID == 0 {
		return errors.ErrInvalidTaskState{
			Details: "cannot update a task that does not exist",
		}
	}
	if td.Synopsis == "" {
		return errors.ErrInvalidTaskState{
			Details: "cannot update a task to have an empty synopsis",
		}
	}
	db := database.DB
	if withDeleted {
		db = db.Unscoped()
	}
	return db.Save(td).Error
}

func (td *TaskData) StopRunningTask() error {
	log := logger.GetLogger("StopRunningTask")
	timesheets, err := Timesheet(new(TimesheetData)).SearchOpen()
	if err != nil {
		utils.PrintAndLogError(errors.LoadTaskError, err, log)
		return err
	}
	// No running tasks, return nil
	if len(timesheets) == 0 {
		return nil
	}
	timesheetData := &timesheets[0]
	stoptime := new(sql.NullTime)
	err = stoptime.Scan(time.Now())
	if err != nil {
		log.Err(err).Msg("error scanning time.Now() into sql.NullTime")
		return err
	}
	timesheetData.StopTime = *stoptime
	err = Timesheet(timesheetData).Update()
	if err != nil {
		utils.PrintAndLogError(errors.StopRunningTaskError, err, log)
		return err
	}
	log.Info().Msgf("task id %d (timesheet id %d) stopped\n", timesheetData.Task.ID, timesheetData.ID)
	fmt.Println(
		color.WhiteString("Task ID %d ", timesheetData.Task.ID),
		color.YellowString("stopped"),
		color.WhiteString(" at %s ", timesheetData.StopTime.Time.Format(constants.TimestampLayout)),
		color.BlueString(timesheetData.StopTime.Time.Sub(timesheetData.StartTime).Truncate(time.Second).String()),
	)
	return nil
}

func (td *TaskData) Clear() {
	td.ID = 0
	td.Synopsis = ""
	td.Description = ""
	td.CreatedAt = time.Now()
	td.DeletedAt.Time = time.Now()
	td.DeletedAt.Valid = false
	td.UpdatedAt = time.Now()
}

func FindTaskBySynopsis(tasks []TaskData, synopsis string) *TaskData {
	for _, task := range tasks {
		if task.Synopsis == synopsis {
			return &task
		}
	}
	return nil
}
