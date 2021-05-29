package models

import (
	"errors"
	"github.com/neflyte/timetracker/internal/database"
	ttErrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestUnit_Task_CreateAndLoad_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := new(TaskData)
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)
	require.True(t, td.ID > 0)

	// Load the task we just created
	td2 := new(TaskData)
	td2.ID = td.ID
	err = Task(td2).Load(false)
	require.Nil(t, err)
	require.Equal(t, "Task-1", td2.Synopsis)
	require.Equal(t, "This is a task", td2.Description)
}

func TestUnit_Task_Load_NotFound(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := new(TaskData)
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)
	require.True(t, td.ID > 0)

	// Try to load a task that does not exist
	td2 := new(TaskData)
	td2.ID = 2
	err = Task(td2).Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Task_Create_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task with an invalid ID
	td := new(TaskData)
	td.ID = 1
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.NotNil(t, err)
	_, ok := err.(ttErrors.ErrInvalidTaskState)
	require.True(t, ok)
}

func TestUnit_Task_Create_EmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task with an empty synopsis
	td := new(TaskData)
	td.Synopsis = ""
	td.Description = "This is a task"
	err := Task(td).Create()
	require.NotNil(t, err)
	_, ok := err.(ttErrors.ErrInvalidTaskState)
	require.True(t, ok)
}

func TestUnit_Task_Load_BySynopsis_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: "Task-1", Description: "Task number one"},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := Task(&tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Load the Task-2 task using its synopsis
	td := new(TaskData)
	td.Synopsis = "Task-2"
	err := Task(td).Load(false)
	require.Nil(t, err)
	require.NotEqual(t, 0, td.ID)
	require.Equal(t, "Task-2", td.Synopsis)
	require.Equal(t, "Task number two", td.Description)
}

func TestUnit_Task_Load_BySynopsis_NotFound(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: "Task-1", Description: "Task number one"},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := Task(&tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Try loading Task-4 which does not exist
	td := new(TaskData)
	td.Synopsis = "Task-4"
	err := Task(td).Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Task_Load_WithDeleted(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := new(TaskData)
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)
	require.True(t, td.ID > 0)

	// Delete the task
	err = Task(td).Delete()
	require.Nil(t, err)

	// Load the task we just deleted
	td2 := new(TaskData)
	td2.ID = td.ID
	err = Task(td2).Load(true)
	require.Nil(t, err)
	require.Equal(t, "Task-1", td2.Synopsis)
	require.Equal(t, "This is a task", td2.Description)
}

func TestUnit_Task_Load_InvalidIDAndEmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Load a task with an invalid ID and an empty synopsis
	td := new(TaskData)
	td.ID = 0
	err := Task(td).Load(false)
	require.NotNil(t, err)
	_, ok := err.(ttErrors.ErrInvalidTaskState)
	require.True(t, ok)
}

func TestUnit_Task_Delete_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := new(TaskData)
	td.Synopsis = "Task-1"
	td.Description = "This is a task"
	err := Task(td).Create()
	require.Nil(t, err)
	require.True(t, td.ID > 0)

	// Delete the task
	err = Task(td).Delete()
	require.Nil(t, err)

	// Try to load the task we just deleted
	td2 := new(TaskData)
	td2.ID = td.ID
	err = Task(td2).Load(false)
	require.NotNil(t, err)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestUnit_Task_Delete_InvalidIDAndEmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Delete a task with invalid ID and empty synopsis
	td := new(TaskData)
	td.ID = 0
	td.Synopsis = ""
	err := Task(td).Delete()
	require.NotNil(t, err)
	_, ok := err.(ttErrors.ErrInvalidTaskState)
	require.True(t, ok)
}

func TestUnit_Task_LoadAll_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create some tasks
	tasks := []TaskData{
		{Synopsis: "Task-1", Description: "Task number one"},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := Task(&tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Load all tasks
	loadedTasks, err := Task(new(TaskData)).LoadAll(false)
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
		{Synopsis: "Task-1", Description: "Task number one"},
		{Synopsis: "Task-2", Description: "Task number two"},
		{Synopsis: "Task-3", Description: "Task number three"},
	}
	for idx := range tasks {
		err := Task(&tasks[idx]).Create()
		require.Nil(t, err)
	}

	// Delete Task-2
	err := Task(&tasks[1]).Delete()
	require.Nil(t, err)

	// Load all tasks with deleted
	loadedTasks, err := Task(new(TaskData)).LoadAll(true)
	require.Nil(t, err)
	require.Len(t, loadedTasks, len(tasks))
	for _, task := range tasks {
		loadedTask := task.FindTaskBySynopsis(loadedTasks, task.Synopsis)
		require.NotNil(t, loadedTask)
		require.Equal(t, task.Synopsis, loadedTask.Synopsis)
		require.Equal(t, task.Description, loadedTask.Description)
	}
}

func TestUnit_Task_Search_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := new(TaskData)
	td.Synopsis = "Foo"
	td.Description = "blah blah blah"
	err := Task(td).Create()
	require.Nil(t, err)

	// Test synopsis
	tasks, err := Task(new(TaskData)).Search("Fo%")
	require.Nil(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 1)
	require.Equal(t, "Foo", tasks[0].Synopsis)
	require.Equal(t, "blah blah blah", tasks[0].Description)

	// Test description
	tasks, err = Task(new(TaskData)).Search("%blah")
	require.Nil(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 1)
	require.Equal(t, "Foo", tasks[0].Synopsis)
	require.Equal(t, "blah blah blah", tasks[0].Description)
}

func TestUnit_Task_Search_NotFound(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := new(TaskData)
	td.Synopsis = "Foo"
	td.Description = "blah blah blah"
	err := Task(td).Create()
	require.Nil(t, err)

	// Test synopsis
	tasks, err := Task(new(TaskData)).Search("Ba%")
	require.Nil(t, err)
	require.Len(t, tasks, 0)

	// Test description
	tasks, err = Task(new(TaskData)).Search("%qux%")
	require.Nil(t, err)
	require.Len(t, tasks, 0)
}

func TestUnit_Update_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := new(TaskData)
	td.Synopsis = "Foo"
	td.Description = "blah blah blah"
	err := Task(td).Create()
	require.Nil(t, err)

	// Update test data
	td.Synopsis = "Bar"
	td.Description = "qux quux"
	err = Task(td).Update(false)
	require.Nil(t, err)

	// Reload data
	reloaded := new(TaskData)
	reloaded.ID = td.ID
	err = Task(reloaded).Load(false)
	require.Nil(t, err)
	require.Equal(t, "Bar", reloaded.Synopsis)
	require.Equal(t, "qux quux", reloaded.Description)
}

func TestUnit_Update_InvalidID(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Update test data
	td := new(TaskData)
	td.ID = 0
	err := Task(td).Update(false)
	require.NotNil(t, err)
	_, ok := err.(ttErrors.ErrInvalidTaskState)
	require.True(t, ok)
}

func TestUnit_Update_EmptySynopsis(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Update test data
	td := new(TaskData)
	td.Synopsis = ""
	err := Task(td).Update(false)
	require.NotNil(t, err)
	_, ok := err.(ttErrors.ErrInvalidTaskState)
	require.True(t, ok)
}

/*func TestUnit_Update_Deleted(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create test data
	td := new(TaskData)
	td.Synopsis = "Foo"
	td.Description = "blah blah blah"
	err := Task(td).Create()
	require.Nil(t, err)

	// Delete test data
	err = Task(td).Delete()
	require.Nil(t, err)

	// Reload the data
	err = Task(td).Load(true)
	require.Nil(t, err)

	// Update test data without deleted flag
	td.Synopsis = "Bar"
	td.Description = "qux quux"
	err = Task(td).Update(false)
	require.Nil(t, err)

	// Reload data to ensure it didn't update
	reloaded := new(TaskData)
	reloaded.ID = td.ID
	err = Task(reloaded).Load(true)
	require.Nil(t, err)
	require.NotEqual(t, "Bar", reloaded.Synopsis)
	require.NotEqual(t, "qux quux", reloaded.Description)

	// Update test data WITH deleted flag
	td.Synopsis = "Bar"
	td.Description = "qux quux"
	err = Task(td).Update(true)
	require.Nil(t, err)

	// Reload data to ensure it did update
	reloaded = new(TaskData)
	reloaded.ID = td.ID
	err = Task(reloaded).Load(true)
	require.Nil(t, err)
	require.Equal(t, "Bar", reloaded.Synopsis)
	require.Equal(t, "qux quux", reloaded.Description)
}*/

// StopRunningTask
func TestUnit_StopRunningTask_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.Set(db)

	// Create a task
	td := new(TaskData)
	td.Synopsis = "Task-1"
	td.Description = "Task number one"
	err := Task(td).Create()
	require.Nil(t, err)

	// Start the task
	tsd := new(TimesheetData)
	tsd.Task = *td
	tsd.StartTime = time.Now()
	err = Timesheet(tsd).Create()
	require.Nil(t, err)

	// Wait 0.5s
	<-time.After(500 * time.Millisecond)

	// Stop the running task
	stopped, err := Task(new(TaskData)).StopRunningTask()
	require.Nil(t, err)
	require.NotNil(t, stopped)
}
