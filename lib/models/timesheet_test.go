package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bluele/factory-go/factory"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/database"
	ttErrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type timesheetFactoryKey int8

const (
	timesheetFactoryTaskIDsKey  timesheetFactoryKey = 0
	timesheetFactoryDatabaseKey timesheetFactoryKey = 1
)

var (
	TimesheetFactory = factory.NewFactory(new(TimesheetData)).
		Attr("TaskID", func(args factory.Args) (interface{}, error) {
			// Pick a random task ID from the list
			availableTaskIDsIntf := args.Context().Value(timesheetFactoryTaskIDsKey)
			if availableTaskIDsIntf == nil {
				return nil, errors.New("got nil task IDs list from context")
			}
			availableTaskIDs, castOK := availableTaskIDsIntf.([]uint)
			if !castOK {
				return nil, errors.New("task IDs list is not a []uint; this is unexpected")
			}
			if len(availableTaskIDs) == 0 {
				return nil, errors.New("task IDs list is empty; this is unexpected")
			}
			randomIndex := rnd.Intn(len(availableTaskIDs) - 1)
			return availableTaskIDs[randomIndex], nil
		}).
		Attr("StartTime", func(_ factory.Args) (interface{}, error) {
			// time.Now() minus random hours from 0-3 and random minutes from 1-59
			randHours := rnd.Intn(3)
			randMinutes := rnd.Intn(59)
			if randMinutes == 0 {
				randMinutes = 1
			}
			timesheetStartTime := time.Now().
				Truncate(time.Second).
				Add(-time.Hour * time.Duration(randHours)).
				Add(-time.Minute * time.Duration(randMinutes))
			return timesheetStartTime, nil
		}).
		Attr("StopTime", func(args factory.Args) (interface{}, error) {
			// StartTime plus random hours from 0-3 and random minutes from 1-59
			tsdPtr, castOK := args.Instance().(*TimesheetData)
			if !castOK {
				return nil, errors.New("arg was not *TimesheetData; this is unexpected")
			}
			randHours := rnd.Intn(3)
			randMinutes := rnd.Intn(59)
			if randMinutes == 0 {
				randMinutes = 1
			}
			timesheetStopTime := tsdPtr.StartTime.
				Add(time.Hour * time.Duration(randHours)).
				Add(time.Minute * time.Duration(randMinutes))
			stopTimeNulltime := sql.NullTime{
				Time:  timesheetStopTime,
				Valid: true,
			}
			return stopTimeNulltime, nil
		}).
		OnCreate(func(args factory.Args) error {
			dbIntf := args.Context().Value(timesheetFactoryDatabaseKey)
			if dbIntf == nil {
				return errors.New("db in context is nil; this is unexpected")
			}
			db, dbOK := dbIntf.(*gorm.DB)
			if !dbOK {
				return errors.New("db in context was not *gorm.DB; this is unexpected")
			}
			tx := db.Begin()
			err := tx.Create(args.Instance()).Error
			if err != nil {
				tx.Rollback()
				return err
			}
			tx.Commit()
			return nil
		})
)

func TestUnit_Timesheet_String(t *testing.T) {
	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	// Create a timesheet
	timeNow := time.Now().Round(time.Second)
	timeNowString := timeNow.String()
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = timeNow
	// Set up expectations
	expectedString := fmt.Sprintf(
		"TimesheetData{Task=%s, StartTime=%s, StopTime=(running)}",
		td.String(),
		timeNowString,
	)
	// Get String value
	actualString := tsd.String()
	require.Equal(t, expectedString, actualString)

	// Set a stop time
	tsd.Data().StopTime.Time = timeNow
	tsd.Data().StopTime.Valid = true

	// Set up expectations
	expectedString = fmt.Sprintf(
		"TimesheetData{Task=%s, StartTime=%s, StopTime=%s}",
		td.String(),
		timeNowString,
		timeNowString,
	)
	// Get String value
	actualString = tsd.String()
	require.Equal(t, expectedString, actualString)
}

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

func TestUnit_Timesheet_RunningTimesheet_Nominal(t *testing.T) {
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

	// Get running timesheet
	runningTS, err := tsd.RunningTimesheet()
	require.Nil(t, err)
	require.NotNil(t, runningTS)
	opensheet := runningTS.Data()
	require.Equal(t, tsd.Data().ID, opensheet.ID)
	require.Equal(t, tsd.Data().Task.ID, opensheet.Task.ID)
	require.Equal(t, tsd.Data().Task.Synopsis, opensheet.Task.Synopsis)
	require.Equal(t, tsd.Data().Task.Description, opensheet.Task.Description)
	require.Equal(t, tsd.Data().StartTime.Format(constants.TimestampLayout), opensheet.StartTime.Format(constants.TimestampLayout))
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

func TestUnit_Timesheet_TaskReport(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create 10 tasks
	taskFactoryCtx := context.WithValue(context.Background(), taskFactoryDatabaseKey, db)
	const numTasks = 10
	createdTasks := make([]*TaskData, numTasks)
	createdTaskIDs := make([]uint, numTasks)
	for x := 0; x < numTasks; x++ {
		taskIntf, err := TaskFactory.CreateWithContext(taskFactoryCtx)
		require.Nil(t, err)
		require.NotNil(t, taskIntf)
		task, castOK := taskIntf.(*TaskData)
		require.True(t, castOK)
		createdTasks[x] = task
		createdTaskIDs[x] = task.ID
	}

	// Create some timesheets for random tasks
	timesheetFactoryCtx := context.WithValue(context.Background(), timesheetFactoryDatabaseKey, db)
	timesheetFactoryCtx = context.WithValue(timesheetFactoryCtx, timesheetFactoryTaskIDsKey, createdTaskIDs)
	const numTimesheets = 100
	createdTimesheets := make([]*TimesheetData, numTimesheets)
	for x := 0; x < numTimesheets; x++ {
		timesheetIntf, err := TimesheetFactory.CreateWithContext(timesheetFactoryCtx)
		require.Nil(t, err)
		require.NotNil(t, timesheetIntf)
		timesheet, castOK := timesheetIntf.(*TimesheetData)
		require.True(t, castOK)
		// Reload the timesheet so we get the task
		ts := NewTimesheetWithData(*timesheet)
		err = ts.Load()
		require.Nil(t, err)
		createdTimesheets[x] = ts.Data()
	}

	// Get a task report for today
	timeNow := time.Now().Truncate(time.Second)
	ts := NewTimesheet()
	reportTasks, err := ts.TaskReport(
		timeNow.AddDate(0, 0, -1),
		timeNow.AddDate(0, 0, 1),
		false,
	)
	require.Nil(t, err)
	require.NotNil(t, reportTasks)
	require.NotEqual(t, 0, len(reportTasks))
	t.Logf("start date\tsynopsis\tduration")
	for _, reportTask := range reportTasks {
		t.Logf(
			"%s\t%s\t%s",
			reportTask.StartDate.Time.Format(constants.TimestampDateLayout),
			reportTask.TaskSynopsis,
			reportTask.Duration().String(),
		)
	}
}
