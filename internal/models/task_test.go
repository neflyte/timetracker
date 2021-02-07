package models

import (
	"github.com/neflyte/timetracker/internal/database"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnit_Task_Search_Nominal(t *testing.T) {
	db := MustOpenTestDB(t)
	defer CloseTestDB(t, db)
	database.DB = db

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
