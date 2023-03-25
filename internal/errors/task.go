package errors

import "fmt"

const (
	// CreateTaskError represents an error that occurs when creating a new task
	CreateTaskError = "error creating new task"
	// DeleteTaskError represents an error that occurs when deleting a task
	DeleteTaskError = "error deleting task"
	// ListTaskError represents an error that occurs when listing tasks
	ListTaskError = "error listing tasks"
	// SearchTaskError represents an error that occurs when searching for tasks
	SearchTaskError = "error searching for tasks"
	// LoadTaskError represents an error that occurs when loading a task
	LoadTaskError = "error loading task"
	// StopRunningTaskError represents an error that occurs when stopping the running task
	StopRunningTaskError = "error stopping the running task"
	// UpdateDeletedTaskError represents an error that occurs when updating a deleted task
	UpdateDeletedTaskError = "cannot update a deleted task"
	// UndeleteTaskError represents an error that occurs when undeleting a task
	UndeleteTaskError = "error undeleting task"
	// UpdateTaskError represents an error that occurs when updating a task
	UpdateTaskError = "error updating task"
	// NoRunningTasksError represents an error that occurs when a running task was expected but not found
	NoRunningTasksError = "a task is not running"
	// OverwriteTaskByCreateError represents an error that occurs when a task is about to be overwritten by creating it again
	OverwriteTaskByCreateError = "cannot overwrite a task by creating it"
	// LoadInvalidTaskError represents an error that occurs when an attempt is made to load a task with an invalid (nonexistant) ID
	LoadInvalidTaskError = "cannot load a task that does not exist"
	// UpdateInvalidTaskError represents an error that occurs when an attempt is made to update a task with an invalid (nonexistant) ID
	UpdateInvalidTaskError = "cannot update a task that does not exist"
	// DeleteInvalidTaskError represents an error that occurs when an attempt is made to delete a task with an invalid (nonexistant) ID
	DeleteInvalidTaskError = "cannot delete a task that does not exist"
	// EmptySynopsisTaskError represents an error that occurs when a task synopsis was expected but not found
	EmptySynopsisTaskError = "cannot create a task with an empty synopsis"
	// UpdateEmptySynopsisTaskError represents an error that occurs when an attempt is made to update an existing task to have an empty synopsis
	UpdateEmptySynopsisTaskError = "cannot update a task to have an empty synopsis"
	// InvalidTaskDataError represents an error that occurs when a task is found to have invalid data
	InvalidTaskDataError = "the task is invalid"
)

// ErrInvalidTaskState represents an error that occurs when a task is in an invalid state
type ErrInvalidTaskState struct {
	Details string
}

func (e ErrInvalidTaskState) Error() string {
	return fmt.Sprintf("Invalid task state: %s", e.Details)
}

// ErrNoRunningTask represents an error that occurs when a running task was expected but not found
type ErrNoRunningTask struct{}

func (e ErrNoRunningTask) Error() string {
	return NoRunningTasksError
}

type ErrInvalidTaskData struct{}

func (e ErrInvalidTaskData) Error() string {
	return InvalidTaskDataError
}
