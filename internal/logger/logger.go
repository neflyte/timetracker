package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"path"
)

const (
	logFileName = "timetracker.log"
)

var (
	rootLogger        zerolog.Logger
	logFileHandle     *os.File
	loggerInitialized = false
)

func InitLogger() {
	if !loggerInitialized {
		// Look for XDG_CONFIG_HOME
		configHome, err := os.UserConfigDir()
		if err != nil {
			configHome = ""
		}
		// Look for HOME
		if configHome == "" {
			configHome, err = os.UserHomeDir()
			if err != nil {
				configHome = ""
			}
		}
		// Use CWD
		if configHome == "" {
			configHome = "."
		}
		logPath := path.Join(configHome, "timetracker")
		// fmt.Printf("logPath=%s\n", logPath)
		_ = os.MkdirAll(logPath, 0755)
		logFilePath := path.Join(logPath, logFileName)
		// fmt.Printf("logFilePath=%s\n", logFilePath)
		logFileHandle, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logFileHandle = nil
		}
		// TODO: allow for different log levels
		if logFileHandle != nil {
			rootLogger = zerolog.New(logFileHandle).With().Timestamp().Logger().Level(zerolog.DebugLevel)
		} else {
			rootLogger = zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(zerolog.DebugLevel)
		}
		loggerInitialized = true
	}
}

func CleanupLogger() {
	if loggerInitialized {
		if logFileHandle != nil {
			err := logFileHandle.Close()
			if err != nil {
				fmt.Printf("error cleaning up logger: %s\n", err)
			}
			logFileHandle = nil
		}
		loggerInitialized = false
	}
}

func GetLogger(funcName string) zerolog.Logger {
	if !loggerInitialized {
		InitLogger()
	}
	return rootLogger.With().Str("func", funcName).Logger()
}
