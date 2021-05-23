package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path"
	"time"
)

const (
	logFileName = "timetracker.log"
)

var (
	RootLogger        zerolog.Logger // RootLogger is the application root logger instance
	logFileHandle     *os.File
	logFilePath       string
	logPath           string
	loggerInitialized = false
)

func InitLogger(logLevel string, console bool) {
	var err error

	if !loggerInitialized {
		configHome := GetConfigHome()
		logPath = path.Join(configHome, "timetracker")
		_ = os.MkdirAll(logPath, 0755)
		logFilePath = path.Join(logPath, logFileName)
		logFileHandle, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			logFileHandle = nil
		}
		// Set up the log writers
		logWriters := make([]io.Writer, 0)
		if logFileHandle != nil {
			logWriters = append(logWriters, logFileHandle)
		}
		if console || logFileHandle == nil {
			logWriters = append(logWriters, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
		}
		// Create a new root logger
		if len(logWriters) > 1 {
			multi := zerolog.MultiLevelWriter(logWriters...)
			RootLogger = zerolog.New(multi).With().Timestamp().Logger()
		} else {
			RootLogger = zerolog.New(logWriters[0]).With().Timestamp().Logger()
		}
		// Set global logger message level
		lvl := zerolog.InfoLevel
		switch logLevel {
		case "fatal":
			lvl = zerolog.FatalLevel
		case "error":
			lvl = zerolog.ErrorLevel
		case "warn":
			lvl = zerolog.WarnLevel
		case "info":
			lvl = zerolog.InfoLevel
		case "debug":
			lvl = zerolog.DebugLevel
		case "trace":
			lvl = zerolog.TraceLevel
		}
		zerolog.SetGlobalLevel(lvl)
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
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr}).
			With().Timestamp().Str("func", funcName).Logger().
			Level(zerolog.TraceLevel)
	}
	return RootLogger.With().Str("func", funcName).Logger()
}

func GetStructLogger(structName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr}).
			With().Timestamp().Str("struct", structName).Logger().
			Level(zerolog.TraceLevel)
	}
	return RootLogger.With().Str("struct", structName).Logger()
}

func GetConfigHome() string {
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
	return configHome
}
