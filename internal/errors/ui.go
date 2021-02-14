package errors

import "fmt"

type ErrPidfileNotFound struct {
	PidfilePath string
}

func (e *ErrPidfileNotFound) Error() string {
	return fmt.Sprintf("Timetracker pidfile was not found; PidfilePath=%s", e.PidfilePath)
}
