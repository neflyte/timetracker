package utils

import "github.com/rs/zerolog"

func ObservableErrorHandler(name string, log zerolog.Logger) func(error) {
	return func(err error) {
		log.Err(err).
			Msgf("error from %s", name)
	}
}

func ObservableCloseHandler(name string, log zerolog.Logger) func() {
	return func() {
		log.Debug().
			Msgf("%s closed", name)
	}
}
