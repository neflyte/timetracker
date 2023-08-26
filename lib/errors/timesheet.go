package errors

import "fmt"

const (
	// CreateTimesheetError represents an error that occurs when creating a timesheet
	CreateTimesheetError = "error creating timesheet"

	// TooManyOpenTimesheetsError = "there are too many open timesheets; this is unexpected"
	// SearchOpenTimesheetsError  = "error searching for open timesheet"
	// UpdateTimesheetError       = "error updating timesheet"

	// OverwriteTimesheetByCreateError represents an error that occurs when a timesheet is about to be overwritten by creating it again
	OverwriteTimesheetByCreateError = "cannot overwrite a timesheet by creating it"
	// TimesheetWithoutTaskError represents an error that occurs when executing an operation on a timesheet that does not have a linked task
	TimesheetWithoutTaskError = "no task is associated with the timesheet"
	// ListTimesheetError represents an error that occurs when timesheets are listed
	ListTimesheetError = "error listing timesheets"
	// LoadInvalidTimesheetError represents an error that occurs when an attempt is made to load a timesheet with an invalid (nonexistant) ID
	LoadInvalidTimesheetError = "cannot load a timesheet that does not exist"
	// UpdateInvalidTimesheetError represents an error that occurs when an attempt is made to update a timesheet with an invalid (nonexistant) ID
	UpdateInvalidTimesheetError = "cannot update a timesheet that does not exist"
	// DeleteInvalidTimesheetError represents an error that occurs when an attempt is made to delete a timesheet with an invalid (nonexistant) ID
	DeleteInvalidTimesheetError = "cannot delete a timesheet that does not exist"
)

// ErrInvalidTimesheetState represents an error that occurs when an timesheet is in an invalid state
type ErrInvalidTimesheetState struct {
	// Details is any extra information related to the error
	Details string
}

func (e ErrInvalidTimesheetState) Error() string {
	return fmt.Sprintf("Invalid timesheet state: %s", e.Details)
}
