package constants

const (
	// TimestampLayout is the format string for use with time.Format() that outputs date + time
	TimestampLayout = `2006-01-02 15:04:05 PM`
	// TimestampDateLayout is the format string for use with time.Format() that outputs a date
	TimestampDateLayout = `2006-01-02`

	// UnicodeClock is the character that represents a running task
	UnicodeClock = "⌛"
	// UnicodeHeavyCheckmark is the character that represents no running tasks (idle)
	UnicodeHeavyCheckmark = "✔"
	// UnicodeHeavyX is the character that represents an error
	UnicodeHeavyX = "✘"
	// ActionLoopDelaySeconds is the number of seconds to delay in the ActionLoop before running the loop again
	ActionLoopDelaySeconds = 5

	// TimesheetStatusIdle represents an idle timesheet
	TimesheetStatusIdle = iota
	// TimesheetStatusRunning represents a running timesheet
	TimesheetStatusRunning
	// TimesheetStatusError represents a timesheet error
	TimesheetStatusError
)
