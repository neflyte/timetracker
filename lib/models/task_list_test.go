package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnit_TaskListFromSliceIntf_Nominal(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	task2 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis2,
		Description: testTaskDescription2,
	})
	input := []interface{}{
		task1,
		task2,
	}
	expected := TaskList{
		task1,
		task2,
	}
	actual := TaskListFromSliceIntf(input)
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskListFromSliceIntf_NilSlice(t *testing.T) {
	expected := make(TaskList, 0)
	actual := TaskListFromSliceIntf(nil)
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskListFromSliceIntf_NonTaskInSlice(t *testing.T) {
	input := []interface{}{
		"foo bar",
		12345,
	}
	expected := make(TaskList, len(input))
	actual := TaskListFromSliceIntf(input)
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskListToSliceIntf_Nominal(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	task2 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis2,
		Description: testTaskDescription2,
	})
	input := TaskList{
		task1,
		task2,
	}
	expected := []interface{}{
		task1,
		task2,
	}
	actual := input.ToSliceIntf()
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskList_Index_Nominal(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	task2 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis2,
		Description: testTaskDescription2,
	})
	task3 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis3,
		Description: testTaskDescription3,
	})
	input := TaskList{
		task1,
		task2,
	}
	expected := 1
	actual := input.Index(task2)
	assert.Equal(t, expected, actual)
	// Check the non-existent case
	assert.Equal(t, -1, input.Index(task3))
}

func TestUnit_TaskList_Index_EmptyList(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	input := TaskList{}
	expected := -1
	actual := input.Index(task1)
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskList_Index_NilTask(t *testing.T) {
	input := TaskList{}
	expected := -1
	actual := input.Index(nil)
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskList_Contains_Nominal(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	task2 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis2,
		Description: testTaskDescription2,
	})
	task3 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis3,
		Description: testTaskDescription3,
	})
	input := TaskList{
		task1,
		task2,
	}
	actual := input.Contains(task2)
	assert.True(t, actual)
	assert.False(t, input.Contains(task3))
}

func TestUnit_TaskList_Contains_NilTask(t *testing.T) {
	input := TaskList{}
	actual := input.Contains(nil)
	assert.False(t, actual)
}

func TestUnit_TaskList_Contains_EmptyList(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	input := TaskList{}
	actual := input.Contains(task1)
	assert.False(t, actual)
}

func TestUnit_TaskList_Names_Nominal(t *testing.T) {
	task1 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis,
		Description: testTaskDescription,
	})
	task2 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis2,
		Description: testTaskDescription2,
	})
	task3 := NewTaskWithData(TaskData{
		Synopsis:    testTaskSynopsis3,
		Description: testTaskDescription3,
	})
	input := TaskList{
		task1,
		task2,
		task3,
	}
	expected := []string{
		task1.DisplayString(),
		task2.DisplayString(),
		task3.DisplayString(),
	}
	actual := input.Names()
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskList_Names_EmptyList(t *testing.T) {
	input := TaskList{}
	expected := make([]string, 0)
	actual := input.Names()
	assert.Equal(t, expected, actual)
}
