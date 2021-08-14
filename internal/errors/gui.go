package errors

import "fmt"

type InvalidTaskReportStartDate struct {
	StartDate string
	Wrapped   error
}

func (e InvalidTaskReportStartDate) Error() string {
	details := "(none)"
	if e.Wrapped != nil {
		details = e.Wrapped.Error()
	}
	return fmt.Sprintf("the start date %s is not valid; details: %s", e.StartDate, details)
}

func (e InvalidTaskReportStartDate) Unwrap() error {
	return e.Wrapped
}

type InvalidTaskReportEndDate struct {
	EndDate string
	Wrapped error
}

func (e InvalidTaskReportEndDate) Error() string {
	details := "(none)"
	if e.Wrapped != nil {
		details = e.Wrapped.Error()
	}
	return fmt.Sprintf("the end date %s is not valid; details: %s", e.EndDate, details)
}

func (e InvalidTaskReportEndDate) Unwrap() error {
	return e.Wrapped
}
