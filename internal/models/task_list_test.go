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
	expected := make(TaskList, 0)
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
	actual := TaskListToSliceIntf(input)
	assert.Equal(t, expected, actual)
}

func TestUnit_TaskListToSliceIntf_NilTaskList(t *testing.T) {
	expected := make([]interface{}, 0)
	actual := TaskListToSliceIntf(nil)
	assert.Equal(t, expected, actual)
}
