package models

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bluele/factory-go/factory"
	"github.com/gofrs/uuid"
	"github.com/neflyte/timetracker/lib/database"
	ttErrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type taskFactoryContextKey int8

const (
	testTaskSynopsis     = "Task-1"
	testTaskDescription  = "This is a task"
	testTaskSynopsis2    = "Foo Bar Baz"
	testTaskDescription2 = "quux Narf fnord"

	synopsisBar        = "Bar"
	descriptionQuxQuux = "qux quux"

	taskFactoryDatabaseKey taskFactoryContextKey = 0
)

var (
	TaskFactory = factory.NewFactory(new(TaskData)).
		Attr("Synopsis", func(_ factory.Args) (interface{}, error) {
			// Use a v4 UUID as the task synopsis
			synuuid, err := uuid.NewV4()
			if err != nil {
				return nil, err
			}
			return synuuid.String(), nil
		}).
		Attr("Description", func(args factory.Args) (interface{}, error) {
			taskData, ok := args.Instance().(*TaskData)
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

func TestUnit_TaskFactory_Create(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	ctx := context.WithValue(context.Background(), taskFactoryDatabaseKey, db)
	taskDataIntf, err := TaskFactory.CreateWithContext(ctx)
	require.Nil(t, err)
	require.NotNil(t, taskDataIntf)
	taskData, taskDataOK := taskDataIntf.(*TaskData)
	require.True(t, taskDataOK, "taskDataIntf was not *TaskData; taskDataIntf=%#v", taskDataIntf)
	require.NotEqual(t, 0, taskData.ID)
	t.Logf("created task %s", taskData.String())
}

func TestUnit_Task_CreateAndLoad_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	require.True(t, td.Data().ID > 0)

	// Load the task we just created
	td2 := NewTask()
	td2.Data().ID = td.Data().ID
	err = td2.Load(false)
	require.Nil(t, err)
	require.Equal(t, testTaskSynopsis, td2.Data().Synopsis)
	require.Equal(t, testTaskDescription, td2.Data().Description)
}

func TestUnit_Task_Load_NotFound(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	require.True(t, td.Data().ID > 0)

	// Try to load a task that does not exist
	td2 := NewTask()
	td2.Data().ID = 2
	err = td2.Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Task_Create_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task with an invalid ID
	td := NewTask()
	td.Data().ID = 1
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTaskState{Details: ttErrors.OverwriteTaskByCreateError}))
}

func TestUnit_Task_Create_EmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task with an empty synopsis
	td := NewTask()
	td.Data().Synopsis = ""
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTaskState{Details: ttErrors.EmptySynopsisTaskError}))
}

func TestUnit_Task_Load_BySynopsis_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: testTaskSynopsis, Description: testTaskDescription},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := NewTaskWithData(tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Load the Task-2 task using its synopsis
	td := NewTask()
	td.Data().Synopsis = "Task-2"
	err := td.Load(false)
	require.Nil(t, err)
	require.NotEqual(t, 0, td.Data().ID)
	require.Equal(t, "Task-2", td.Data().Synopsis)
	require.Equal(t, "Task number two", td.Data().Description)
}

func TestUnit_Task_Load_BySynopsis_NotFound(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: testTaskSynopsis, Description: testTaskDescription},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := NewTaskWithData(tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Try loading Task-4 which does not exist
	td := NewTask()
	td.Data().Synopsis = "Task-4"
	err := td.Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Task_Load_WithDeleted(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	require.True(t, td.Data().ID > 0)

	// Delete the task
	err = td.Delete()
	require.Nil(t, err)

	// Load the task we just deleted
	td2 := NewTask()
	td2.Data().ID = td.Data().ID
	err = td2.Load(true)
	require.Nil(t, err)
	require.Equal(t, testTaskSynopsis, td2.Data().Synopsis)
	require.Equal(t, testTaskDescription, td2.Data().Description)
}

func TestUnit_Task_Load_InvalidIDAndEmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Load a task with an invalid ID and an empty synopsis
	td := NewTask()
	td.Data().ID = 0
	err := td.Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTaskState{Details: ttErrors.LoadInvalidTaskError}))
}

func TestUnit_Task_Delete_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	require.True(t, td.Data().ID > 0)

	// Delete the task
	err = td.Delete()
	require.Nil(t, err)

	// Try to load the task we just deleted
	td2 := NewTask()
	td2.Data().ID = td.Data().ID
	err = td2.Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Task_Delete_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Delete a task with invalid ID
	td := NewTask()
	td.Data().ID = 0
	err := td.Delete()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTaskState{Details: ttErrors.DeleteInvalidTaskError}))
}

func TestUnit_Task_LoadAll_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: testTaskSynopsis, Description: testTaskDescription},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := NewTaskWithData(tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Load all tasks
	loadedTasks, err := NewTask().LoadAll(false)
	require.Nil(t, err)
	require.Len(t, loadedTasks, len(tasks))
	for idx, loadedTask := range loadedTasks {
		require.Equal(t, tasks[idx].Synopsis, loadedTask.Synopsis)
		require.Equal(t, tasks[idx].Description, loadedTask.Description)
	}
}

func TestUnit_Task_LoadAll_WithDeleted(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: testTaskSynopsis, Description: testTaskDescription},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		task := NewTaskWithData(tasks[idx])
		err := task.Create()
		require.Nil(t, err)
		tasks[idx] = *task.Data()
		require.NotEqual(t, uint(0), tasks[idx].ID)
	}

	// Delete Task-2
	err := NewTaskWithData(tasks[1]).Delete()
	require.Nil(t, err)

	// Load all tasks with deleted
	loadedTasks, err := NewTask().LoadAll(true)
	require.Nil(t, err)
	require.Len(t, loadedTasks, len(tasks))
	for _, task := range tasks {
		loadedTask := task.FindTaskBySynopsis(loadedTasks, task.Synopsis)
		require.NotNil(t, loadedTask)
		require.Equal(t, task.Synopsis, loadedTask.Synopsis)
		require.Equal(t, task.Description, loadedTask.Description)
	}
}

func TestUnit_Task_Clone(t *testing.T) {
	// Create test data
	timeNow := time.Now()
	deletedAtTime := gorm.DeletedAt{
		Time:  timeNow,
		Valid: true,
	}
	td := NewTask()
	td.Data().ID = uint(42)
	td.Data().CreatedAt = timeNow
	td.Data().UpdatedAt = timeNow
	td.Data().DeletedAt = deletedAtTime
	td.Data().Synopsis = testTaskSynopsis2
	td.Data().Description = testTaskDescription
	// Clone
	clone := td.Clone()
	require.Equal(t, uint(42), clone.Data().ID)
	require.Equal(t, timeNow, clone.Data().CreatedAt)
	require.Equal(t, timeNow, clone.Data().UpdatedAt)
	require.Equal(t, deletedAtTime, clone.Data().DeletedAt)
	require.Equal(t, testTaskSynopsis2, clone.Data().Synopsis)
	require.Equal(t, testTaskDescription, clone.Data().Description)
}

func TestUnit_Task_Clear(t *testing.T) {
	// Create test data
	td := NewTask()
	td.Data().ID = uint(42)
	td.Data().Synopsis = testTaskSynopsis2
	td.Data().Description = testTaskDescription

	// Clear the task
	td.Clear()
	require.Equal(t, uint(0), td.Data().ID)
	require.Equal(t, "", td.Data().Synopsis)
	require.Equal(t, "", td.Data().Description)
}

func TestUnit_Task_Search_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	require.NotEqual(t, 0, td.Data().ID)

	// Test synopsis
	tasks, err := NewTask().Search("%-1")
	require.Nil(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 1)
	require.Equal(t, testTaskSynopsis, tasks[0].Synopsis)
	require.Equal(t, testTaskDescription, tasks[0].Description)

	// Test description
	tasks, err = NewTask().Search("This is a%")
	require.Nil(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 1)
	require.Equal(t, testTaskSynopsis, tasks[0].Synopsis)
	require.Equal(t, testTaskDescription, tasks[0].Description)
}

func TestUnit_Task_Search_NotFound(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)

	// Test synopsis
	tasks, err := NewTask().Search("Ba%")
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	// Test description
	tasks, err = NewTask().Search("%qux%")
	require.Nil(t, err)
	require.Len(t, tasks, 0)
}

func TestUnit_Task_SearchBySynopsis_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = testTaskDescription
	err := td.Create()
	require.Nil(t, err)
	td = NewTask()
	td.Data().Synopsis = testTaskSynopsis2
	td.Data().Description = testTaskDescription2
	err = td.Create()
	require.Nil(t, err)

	// Search for something that exists
	tasks, err := NewTask().SearchBySynopsis(testTaskSynopsis2)
	require.Nil(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, testTaskSynopsis2, tasks[0].Synopsis)
	require.Equal(t, testTaskDescription2, tasks[0].Description)

	// Search for something that doesn't exist
	tasks, err = NewTask().SearchBySynopsis("gotta gotta get up to get down")
	require.Nil(t, err)
	require.Len(t, tasks, 0)
}

func TestUnit_Task_Resolve_Nominal(t *testing.T) {
	td := NewTask()

	// Empty argument
	id, syn := td.Resolve("")
	require.Equal(t, uint(0), id)
	require.Equal(t, "", syn)

	// Text string
	id, syn = td.Resolve(testTaskSynopsis)
	require.Equal(t, uint(0), id)
	require.Equal(t, testTaskSynopsis, syn)

	// Numeric string
	id, syn = td.Resolve("42")
	require.Equal(t, uint(42), id)
	require.Equal(t, syn, "")
}

func TestUnit_FindTaskBySynopsis_TaskNotExists(t *testing.T) {
	tasks := NewTask().FindTaskBySynopsis(make([]TaskData, 0), testTaskSynopsis)
	require.Nil(t, tasks)
}

func TestUnit_Update_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := NewTask()
	td.Data().Synopsis = "Foo"
	td.Data().Description = "blah blah blah"
	err := td.Create()
	require.Nil(t, err)

	// Update test data
	td.Data().Synopsis = synopsisBar
	td.Data().Description = descriptionQuxQuux
	err = td.Update(false)
	require.Nil(t, err)

	// Reload data
	reloaded := NewTask()
	reloaded.Data().ID = td.Data().ID
	err = reloaded.Load(false)
	require.Nil(t, err)
	require.Equal(t, synopsisBar, reloaded.Data().Synopsis)
	require.Equal(t, descriptionQuxQuux, reloaded.Data().Description)
}

func TestUnit_Update_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Update test data
	td := NewTask()
	td.Data().ID = 0
	err := td.Update(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTaskState{Details: ttErrors.UpdateInvalidTaskError}))
}

func TestUnit_Update_EmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Update test data
	td := NewTask()
	td.Data().ID = 1
	td.Data().Synopsis = ""
	err := td.Update(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrInvalidTaskState{Details: ttErrors.UpdateEmptySynopsisTaskError}))
}

func TestUnit_Update_Deleted(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := NewTask()
	td.Data().Synopsis = "Foo"
	td.Data().Description = "blah blah blah"
	err := td.Create()
	require.Nil(t, err)

	// Delete test data
	err = td.Delete()
	require.Nil(t, err)

	// Reload the data
	err = td.Load(true)
	require.Nil(t, err)

	// Update test data without deleted flag
	td.Data().Synopsis = synopsisBar
	td.Data().Description = descriptionQuxQuux
	err = td.Update(false)
	require.NotNil(t, err)

	// Reload data to ensure it didn't update
	reloaded := NewTask()
	reloaded.Data().ID = td.Data().ID
	err = reloaded.Load(true)
	require.Nil(t, err)
	require.NotEqual(t, synopsisBar, reloaded.Data().Synopsis)
	require.NotEqual(t, descriptionQuxQuux, reloaded.Data().Description)

	// Update test data WITH deleted flag
	td.Data().Synopsis = synopsisBar
	td.Data().Description = descriptionQuxQuux
	err = td.Update(true)
	require.Nil(t, err)

	// Reload data to ensure it did update
	reloaded = NewTask()
	reloaded.Data().ID = td.Data().ID
	err = reloaded.Load(true)
	require.Nil(t, err)
	require.Equal(t, synopsisBar, reloaded.Data().Synopsis)
	require.Equal(t, descriptionQuxQuux, reloaded.Data().Description)
}

// StopRunningTask
func TestUnit_StopRunningTask_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = "Task number one"
	err := td.Create()
	require.Nil(t, err)

	// Start the task
	tsd := NewTimesheet()
	tsd.Data().Task = *td.Data()
	tsd.Data().StartTime = time.Now()
	err = tsd.Create()
	require.Nil(t, err)

	// Wait 0.5s
	<-time.After(500 * time.Millisecond)

	// Stop the running task
	stopped, err := NewTask().StopRunningTask()
	require.Nil(t, err)
	require.NotNil(t, stopped)
}

func TestUnit_StopRunningTask_NoRunningTask(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := NewTask()
	td.Data().Synopsis = testTaskSynopsis
	td.Data().Description = "Task number one"
	err := td.Create()
	require.Nil(t, err)

	// Stop the running task of which there are none
	stopped, err := NewTask().StopRunningTask()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, ttErrors.ErrNoRunningTask{}))
	require.Nil(t, stopped)
}
