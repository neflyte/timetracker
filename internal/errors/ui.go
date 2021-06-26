package errors

import "fmt"

// ErrPidfileNotFound represents an error that occurs when a Process ID (PID) file cannot be found
type ErrPidfileNotFound struct {
	PidfilePath string
}

func (e *ErrPidfileNotFound) Error() string {
	return fmt.Sprintf("Timetracker pidfile was not found; PidfilePath=%s", e.PidfilePath)
}
