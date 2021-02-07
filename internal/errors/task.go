package errors

import "fmt"

const (
	CreateTaskError        = "error creating new task"
	DeleteTaskError        = "error deleting task"
	ListTaskError          = "error listing tasks"
	SearchTaskError        = "error searching for tasks"
	LoadTaskError          = "error loading task"
	StopRunningTaskError   = "error stopping the running task"
	UpdateDeletedTaskError = "cannot update a deleted task"
	UndeleteTaskError      = "error undeleting task"
	UpdateTaskError        = "error updating task"
)

type ErrInvalidTaskState struct {
	Details string
}

func (e ErrInvalidTaskState) Error() string {
	return fmt.Sprintf("Invalid task state: %s", e.Details)
}
