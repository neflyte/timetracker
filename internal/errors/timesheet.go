package errors

import "fmt"

const (
	CreateTimesheetError       = "error creating timesheet"
	TooManyOpenTimesheetsError = "there are too many open timesheets; this is unexpected"
	SearchOpenTimesheetsError  = "error searching for open timesheet"
	UpdateTimesheetError       = "error updating timesheet"
	ListTimesheetError         = "error listing timesheets"
)

type ErrInvalidTimesheetState struct {
	Details string
}

func (e ErrInvalidTimesheetState) Error() string {
	return fmt.Sprintf("Invalid timesheet state: %s", e.Details)
}
