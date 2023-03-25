package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/bluele/factory-go/factory"
	"github.com/gofrs/uuid"
	"github.com/neflyte/timetracker/internal/models"
	"gorm.io/gorm"
)

type taskFactoryContextKey int8

const (
	taskFactoryDatabaseKey taskFactoryContextKey = 0
)

var (
	TaskFactory = factory.NewFactory(new(models.TaskData)).
		Attr("Synopsis", func(_ factory.Args) (interface{}, error) {
			// Use a v4 UUID as the task synopsis
			synuuid, err := uuid.NewV4()
			if err != nil {
				return nil, err
			}
			return synuuid.String(), nil
		}).
		Attr("Description", func(args factory.Args) (interface{}, error) {
			taskData, ok := args.Instance().(*models.TaskData)
			if !ok {
				return nil, errors.New("args for Description was not *TaskData; this is unexpected")
			}
			return fmt.Sprintf("description for task %s", taskData.Synopsis), nil
		}).
		OnCreate(func(args factory.Args) error {
			dbIntf := args.Context().Value(taskFactoryDatabaseKey)
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

func TestUnit_StopRunningTimesheet_NoRunningTimesheet(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	err := StopRunningTimesheet()
	require.Nil(t, err)
}

func TestUnit_StopRunningTimesheet_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	ctx := context.WithValue(context.Background(), taskFactoryDatabaseKey, db)
	taskDataIntf, err := TaskFactory.CreateWithContext(ctx)
	require.Nil(t, err)
	require.NotNil(t, taskDataIntf)
	taskData, taskDataOK := taskDataIntf.(*models.TaskData)
	require.True(t, taskDataOK, "taskDataIntf was not *TaskData; taskDataIntf=%#v", taskDataIntf)
	require.NotEqual(t, 0, taskData.ID)

	// Start the task
	tsd := models.NewTimesheet()
	tsd.Data().Task = *taskData.Data()
	tsd.Data().StartTime = time.Now()
	err = tsd.Create()
	require.Nil(t, err)

	// Wait 0.5s
	<-time.After(500 * time.Millisecond)

	// Stop the timesheet
	err = StopRunningTimesheet()
	require.Nil(t, err)

	// Assert that the timesheet stopped
	err = tsd.Data().Load()
	require.Nil(t, err)
	require.True(t, tsd.Data().StopTime.Valid)
	require.True(t, time.Now().After(tsd.Data().StopTime.Time))
}

func TestUnit_StartRunningTimesheet_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	ctx := context.WithValue(context.Background(), taskFactoryDatabaseKey, db)
	taskDataIntf, err := TaskFactory.CreateWithContext(ctx)
	require.Nil(t, err)
	require.NotNil(t, taskDataIntf)
	taskData, taskDataOK := taskDataIntf.(*models.TaskData)
	require.True(t, taskDataOK, "taskDataIntf was not *TaskData; taskDataIntf=%#v", taskDataIntf)
	require.NotEqual(t, 0, taskData.ID)

	// Start the timesheet
	err = StartRunningTimesheet(models.NewTaskWithData(*taskData))
	require.Nil(t, err)

	// Assert that it started
	openTimesheets, err := models.NewTimesheet().SearchOpen()
	require.Nil(t, err)
	require.Len(t, openTimesheets, 1)
	require.Equal(t, taskData.ID, openTimesheets[0].Task.ID)
}

func TestUnit_StartRunningTimesheet_NilTask(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	err := StartRunningTimesheet(nil)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, tterrors.ErrInvalidTaskData{}))
}

func TestUnit_StartRunningTimesheet_InvalidTask(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	expectedErr := tterrors.ErrInvalidTimesheetState{
		Details: tterrors.TimesheetWithoutTaskError,
	}

	err := StartRunningTimesheet(models.NewTask())
	require.NotNil(t, err)
	require.Equal(t, expectedErr, err)
}
