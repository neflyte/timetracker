package utils

import "github.com/rs/zerolog"

// ObservableErrorHandler returns an error handler that logs to the Error log level
func ObservableErrorHandler(name string, log zerolog.Logger) func(error) {
	return func(err error) {
		log.Err(err).
			Msgf("error from %s", name)
	}
}

// ObservableCloseHandler returns a handler that logs a closed status to the Debug log level
func ObservableCloseHandler(name string, log zerolog.Logger) func() {
	return func() {
		log.Debug().
			Msgf("%s closed", name)
	}
}
