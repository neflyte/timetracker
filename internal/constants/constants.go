package constants

const (
	TimestampLayout       = `2006-01-02 15:04:05 PM` // TimestampLayout is the format string for use with time.Format() that outputs date + time
	TimestampDateLayout   = `2006-01-02`             // TimestampDateLayout is the format string for use with time.Format() that outputs a date
	UnicodeClock          = "⌛"
	UnicodeHeavyCheckmark = "✔"
	UnicodeHeavyX         = "✘"
)

const (
	TimesheetStatusIdle = iota
	TimesheetStatusRunning
	TimesheetStatusError
)
