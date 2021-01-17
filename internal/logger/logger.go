package logger

import (
	"github.com/rs/zerolog"
)

var (
	rootLogger        zerolog.Logger
	loggerInitialized = false
)

func InitLogger() {
	if !loggerInitialized {
		rootLogger = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
		loggerInitialized = true
	}
}

func GetLogger(funcName string) zerolog.Logger {
	if !loggerInitialized {
		InitLogger()
	}
	return rootLogger.With().Str("func", funcName).Logger()
}
