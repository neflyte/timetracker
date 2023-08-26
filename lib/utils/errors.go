package utils

import "errors"

var (
	ErrStalePidfile = errors.New("stale pidfile detected")
)

type ErrCheckPidfile struct {
	wrapped error
}

func (e ErrCheckPidfile) Unwrap() error {
	return e.wrapped
}

func (e ErrCheckPidfile) Error() string {
	return "error checking pidfile"
}
