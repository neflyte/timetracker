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
)

type ErrInvalidTaskState struct {
	Details string
}

func (e ErrInvalidTaskState) Error() string {
	return fmt.Sprintf("Invalid task state: %s", e.Details)
}
