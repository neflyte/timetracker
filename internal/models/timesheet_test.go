package models

import (
	"errors"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/database"
	ttErrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestUnit_Timesheet_CreateAndLoad_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTaskData()
	td.Synopsis = testTaskSynopsis
	td.Description = testTaskDescription
	err := Task(td).Create()
	require.Nil(t, err)

	// Create a timesheet
	tsd := new(TimesheetData)
	tsd.Task = *td
	tsd.StartTime = time.Now()
	err = Timesheet(tsd).Create()
	require.Nil(t, err)
	require.NotEqual(t, tsd.ID, 0)

	// Re-load the timesheet
	reloaded := new(TimesheetData)
	reloaded.ID = tsd.ID
	err = Timesheet(reloaded).Load()
	require.Nil(t, err)
	require.Equal(t, tsd.Task.ID, reloaded.Task.ID)
	require.Equal(t, tsd.Task.Synopsis, reloaded.Task.Synopsis)
	require.Equal(t, tsd.Task.Description, reloaded.Task.Description)
	require.Equal(t, tsd.StartTime.Format(constants.TimestampLayout), reloaded.StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_Create_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := new(TimesheetData)
	tsd.ID = 1
	err := Timesheet(tsd).Create()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.OverwriteTimesheetByCreateError}))
}

func TestUnit_Timesheet_Create_InvalidTaskID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := new(TimesheetData)
	tsd.Task.ID = 0
	err := Timesheet(tsd).Create()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.TimesheetWithoutTaskError}))
}

func TestUnit_Timesheet_Load_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := new(TimesheetData)
	tsd.Task.ID = 0
	err := Timesheet(tsd).Load()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{
		Details: ttErrors.LoadInvalidTimesheetError,
	}))
}

func TestUnit_Timesheet_Delete_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTaskData()
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)

	// Create a timesheet
	tsd := new(TimesheetData)
	tsd.Task = *td
	tsd.StartTime = time.Now()
	err = Timesheet(tsd).Create()
	require.Nil(t, err)
	require.NotEqual(t, tsd.ID, 0)

	// Delete the timesheet
	err = Timesheet(tsd).Delete()
	require.Nil(t, err)

	// Try to re-load the timesheet
	reloaded := new(TimesheetData)
	reloaded.ID = tsd.ID
	err = Timesheet(reloaded).Load()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Timesheet_Delete_InvalidID(t *testing.T) {
	tsd := new(TimesheetData)
	err := Timesheet(tsd).Delete()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.DeleteInvalidTimesheetError}))
}

func TestUnit_Timesheet_LoadAll_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create two tasks
	td := NewTaskData()
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)
	td2 := NewTaskData()
	td2.Synopsis = "Task-2"
	td2.Description = "Task number two"
	err = Task(td2).Create()
	require.Nil(t, err)

	// Create a closed timesheet for Task-1
	tsd := new(TimesheetData)
	tsd.Task = *td
	tsd.StartTime = time.Now().Add(-10 * time.Minute)
	tsd.StopTime.Time = time.Now()
	tsd.StopTime.Valid = true
	err = Timesheet(tsd).Create()
	require.Nil(t, err)

	// Create an open timesheet for Task-2
	tsd2 := new(TimesheetData)
	tsd2.Task = *td2
	tsd2.StartTime = time.Now()
	err = Timesheet(tsd2).Create()
	require.Nil(t, err)

	const expectedTimesheets = 2

	// Load all timesheets
	timesheets, err := Timesheet(new(TimesheetData)).LoadAll(false)
	require.Nil(t, err)

	// Verify that we loaded all of the timesheets
	require.Len(t, timesheets, expectedTimesheets)
	require.Equal(t, tsd.ID, timesheets[0].ID)
	require.Equal(t, tsd.Task.ID, timesheets[0].Task.ID)
	require.Equal(t, tsd.Task.Synopsis, timesheets[0].Task.Synopsis)
	require.Equal(t, tsd.Task.Description, timesheets[0].Task.Description)
	require.Equal(t, tsd.StartTime.Format(constants.TimestampLayout), timesheets[0].StartTime.Format(constants.TimestampLayout))
	require.Equal(t, tsd.StopTime.Time.Format(constants.TimestampLayout), timesheets[0].StopTime.Time.Format(constants.TimestampLayout))
	require.Equal(t, tsd2.ID, timesheets[1].ID)
	require.Equal(t, tsd2.Task.ID, timesheets[1].Task.ID)
	require.Equal(t, tsd2.Task.Synopsis, timesheets[1].Task.Synopsis)
	require.Equal(t, tsd2.Task.Description, timesheets[1].Task.Description)
	require.Equal(t, tsd2.StartTime.Format(constants.TimestampLayout), timesheets[1].StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_SearchOpen_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTaskData()
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)

	// Create an open timesheet for the task
	tsd := new(TimesheetData)
	tsd.Task = *td
	tsd.StartTime = time.Now()
	err = Timesheet(tsd).Create()
	require.Nil(t, err)

	// Search for open timesheets
	opensheets, err := Timesheet(new(TimesheetData)).SearchOpen()
	require.Nil(t, err)
	require.Len(t, opensheets, 1)
	require.Equal(t, tsd.ID, opensheets[0].ID)
	require.Equal(t, tsd.Task.ID, opensheets[0].Task.ID)
	require.Equal(t, tsd.Task.Synopsis, opensheets[0].Task.Synopsis)
	require.Equal(t, tsd.Task.Description, opensheets[0].Task.Description)
	require.Equal(t, tsd.StartTime.Format(constants.TimestampLayout), opensheets[0].StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_Update_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTaskData()
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)

	// Create an open timesheet for the task
	tsd := new(TimesheetData)
	tsd.Task = *td
	tsd.StartTime = time.Now()
	err = Timesheet(tsd).Create()
	require.Nil(t, err)

	unexpectedStartTime := tsd.StartTime.Format(constants.TimestampLayout)

	// Update the start time
	tsd.StartTime = tsd.StartTime.Add(-15 * time.Minute)
	err = Timesheet(tsd).Update()
	require.Nil(t, err)

	// Reload the timesheet
	reloaded := new(TimesheetData)
	reloaded.ID = tsd.ID
	err = Timesheet(reloaded).Load()
	require.Nil(t, err)
	require.NotEqual(t, unexpectedStartTime, reloaded.StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_Update_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := new(TimesheetData)
	tsd.ID = 0
	err := Timesheet(tsd).Update()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.UpdateInvalidTimesheetError}))
}

func TestUnit_Timesheet_Update_InvalidTaskID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := new(TimesheetData)
	tsd.ID = 1
	tsd.Task.ID = 0
	err := Timesheet(tsd).Update()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.TimesheetWithoutTaskError}))
}
