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
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)

	// Create a timesheet
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = time.Now()
	err = tsd.Create()
	require.Nil(t, err)
	require.NotEqual(t, tsd.Data().ID, 0)

	// Re-load the timesheet
	reloaded := NewTimesheet()
	reloaded.Data().ID = tsd.Data().ID
	err = reloaded.Load()
	require.Nil(t, err)
	require.Equal(t, tsd.Data().Task.ID, reloaded.Data().Task.ID)
	require.Equal(t, tsd.Data().Task.Synopsis, reloaded.Data().Task.Synopsis)
	require.Equal(t, tsd.Data().Task.Description, reloaded.Data().Task.Description)
	require.Equal(t, tsd.Data().StartTime.Format(constants.TimestampLayout), reloaded.Data().StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_Create_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := NewTimesheet()
	tsd.Data().ID = 1
	err := tsd.Create()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.OverwriteTimesheetByCreateError}))
}

func TestUnit_Timesheet_Create_InvalidTaskID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := NewTimesheet()
	tsd.Data().Task.ID = 0
	err := tsd.Create()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.TimesheetWithoutTaskError}))
}

func TestUnit_Timesheet_Load_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := NewTimesheet()
	tsd.Data().Task.ID = 0
	err := tsd.Load()
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
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)

	// Create a timesheet
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = time.Now()
	err = tsd.Create()
	require.Nil(t, err)
	require.NotEqual(t, tsd.Data().ID, 0)

	// Delete the timesheet
	err = tsd.Delete()
	require.Nil(t, err)

	// Try to re-load the timesheet
	reloaded := NewTimesheet()
	reloaded.Data().ID = tsd.Data().ID
	err = reloaded.Load()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Timesheet_Delete_InvalidID(t *testing.T) {
	tsd := NewTimesheet()
	err := tsd.Delete()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.DeleteInvalidTimesheetError}))
}

func TestUnit_Timesheet_LoadAll_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create two tasks
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	td2 := NewTask()
	td2.Data().Synopsis = "Task-2"
	td2.Data().Description = "Task number two"
	err = td2.Create()
	require.Nil(t, err)

	// Create a closed timesheet for Task-1
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = time.Now().Add(-10 * time.Minute)
	tsd.Data().StopTime.Time = time.Now()
	tsd.Data().StopTime.Valid = true
	err = tsd.Create()
	require.Nil(t, err)

	// Create an open timesheet for Task-2
	tsd2 := NewTimesheet()
	tsd2.Data().Task = *td2.Data()
	tsd2.Data().StartTime = time.Now()
	err = tsd2.Create()
	require.Nil(t, err)

	const expectedTimesheets = 2

	// Load all timesheets
	timesheets, err := NewTimesheet().LoadAll(false)
	require.Nil(t, err)

	// Verify that we loaded all of the timesheets
	require.Len(t, timesheets, expectedTimesheets)
	require.Equal(t, tsd.Data().ID, timesheets[0].ID)
	require.Equal(t, tsd.Data().Task.ID, timesheets[0].Task.ID)
	require.Equal(t, tsd.Data().Task.Synopsis, timesheets[0].Task.Synopsis)
	require.Equal(t, tsd.Data().Task.Description, timesheets[0].Task.Description)
	require.Equal(t, tsd.Data().StartTime.Format(constants.TimestampLayout), timesheets[0].StartTime.Format(constants.TimestampLayout))
	require.Equal(t, tsd.Data().StopTime.Time.Format(constants.TimestampLayout), timesheets[0].StopTime.Time.Format(constants.TimestampLayout))
	require.Equal(t, tsd2.Data().ID, timesheets[1].ID)
	require.Equal(t, tsd2.Data().Task.ID, timesheets[1].Task.ID)
	require.Equal(t, tsd2.Data().Task.Synopsis, timesheets[1].Task.Synopsis)
	require.Equal(t, tsd2.Data().Task.Description, timesheets[1].Task.Description)
	require.Equal(t, tsd2.Data().StartTime.Format(constants.TimestampLayout), timesheets[1].StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_SearchOpen_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)

	// Create an open timesheet for the task
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = time.Now()
	err = tsd.Create()
	require.Nil(t, err)

	// Search for open timesheets
	opensheets, err := NewTimesheet().SearchOpen()
	require.Nil(t, err)
	require.Len(t, opensheets, 1)
	require.Equal(t, tsd.Data().ID, opensheets[0].ID)
	require.Equal(t, tsd.Data().Task.ID, opensheets[0].Task.ID)
	require.Equal(t, tsd.Data().Task.Synopsis, opensheets[0].Task.Synopsis)
	require.Equal(t, tsd.Data().Task.Description, opensheets[0].Task.Description)
	require.Equal(t, tsd.Data().StartTime.Format(constants.TimestampLayout), opensheets[0].StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_Update_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)

	// Create an open timesheet for the task
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = time.Now()
	err = tsd.Create()
	require.Nil(t, err)

	unexpectedStartTime := tsd.Data().StartTime.Format(constants.TimestampLayout)

	// Update the start time
	tsd.Data().StartTime = tsd.Data().StartTime.Add(-15 * time.Minute)
	err = tsd.Update()
	require.Nil(t, err)

	// Reload the timesheet
	reloaded := NewTimesheet()
	reloaded.Data().ID = tsd.Data().ID
	err = reloaded.Load()
	require.Nil(t, err)
	require.NotEqual(t, unexpectedStartTime, reloaded.Data().StartTime.Format(constants.TimestampLayout))
}

func TestUnit_Timesheet_Update_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := NewTimesheet()
	tsd.Data().ID = 0
	err := tsd.Update()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.UpdateInvalidTimesheetError}))
}

func TestUnit_Timesheet_Update_InvalidTaskID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	tsd := NewTimesheet()
	tsd.Data().ID = 1
	tsd.Data().Task.ID = 0
	err := tsd.Update()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTimesheetState{Details: ttErrors.TimesheetWithoutTaskError}))
}
