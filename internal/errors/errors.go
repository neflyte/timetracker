package errors

const (
	// ScanNowIntoSQLNullTimeError describes an error that occurs when scanning time.Now() into a sql.NullTime object
	ScanNowIntoSQLNullTimeError = "error scanning time.Now() into sql.NullTime"
)

// ErrScanNowIntoSQLNull represents an error that occurs when scanning time.Now() into a sql.NullTime object
type ErrScanNowIntoSQLNull struct {
	// Wrapped is a Wrapped error
	Wrapped error
}

func (e ErrScanNowIntoSQLNull) Error() string {
	return ScanNowIntoSQLNullTimeError
}

// Unwrap implements a Wrapped error
func (e ErrScanNowIntoSQLNull) Unwrap() error {
	return e.Wrapped
}
