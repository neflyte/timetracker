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
	// RootLogger is the application root logger instance
	RootLogger zerolog.Logger

	logFileHandle     *os.File
	logFilePath       string
	logPath           string
	loggerInitialized = false

	levelMap = map[string]zerolog.Level{
		"fatal": zerolog.FatalLevel,
		"error": zerolog.ErrorLevel,
		"warn":  zerolog.WarnLevel,
		"info":  zerolog.InfoLevel,
		"debug": zerolog.DebugLevel,
		"trace": zerolog.TraceLevel,
	}
)

func init() {
	// Override zerolog's default time field format so it doesn't truncate nanoseconds
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

func InitLogger(logLevel string, console bool) {
	var err error

	if !loggerInitialized {
		return
	}
	configHome := GetConfigHome()
	logPath = path.Join(configHome, "timetracker")
	err = os.MkdirAll(logPath, 0755)
	if err != nil {
		fmt.Printf("")
	}
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
		logWriters = append(logWriters, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli})
	}
	// Create a new root logger
	if len(logWriters) > 1 {
		multi := zerolog.MultiLevelWriter(logWriters...)
		RootLogger = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		RootLogger = zerolog.New(logWriters[0]).With().Timestamp().Logger()
	}
	// Set global logger message level
	lvl, ok := levelMap[logLevel]
	if !ok {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	loggerInitialized = true
}

func CleanupLogger() {
	if loggerInitialized {
		if logFileHandle != nil {
			err := logFileHandle.Close()
			if err != nil {
				fmt.Printf("*  error cleaning up logger: %s\n", err)
			}
			logFileHandle = nil
		}
		loggerInitialized = false
	}
}

func GetLogger(funcName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).
			With().Timestamp().Str("func", funcName).Logger().
			Level(zerolog.TraceLevel)
	}
	return RootLogger.With().Str("func", funcName).Logger()
}

func GetStructLogger(structName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).
			With().Timestamp().Str("struct", structName).Logger().
			Level(zerolog.TraceLevel)
	}
	return RootLogger.With().Str("struct", structName).Logger()
}

func GetPackageLogger(packageName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).
			With().Timestamp().Str("package", packageName).Logger().
			Level(zerolog.TraceLevel)
	}
	return RootLogger.With().Str("package", packageName).Logger()
}

func GetFuncLogger(existingLog zerolog.Logger, funcName string) zerolog.Logger {
	return existingLog.With().Str("func", funcName).Logger()
}

func GetConfigHome() string {
	log := GetLogger("GetConfigHome")
	// Look for XDG_CONFIG_HOME
	configHome, err := os.UserConfigDir()
	if err != nil {
		log.Err(err).Msg("error getting user config directory")
		configHome = ""
	}
	// Look for HOME
	if configHome == "" {
		configHome, err = os.UserHomeDir()
		if err != nil {
			log.Err(err).Msg("error getting user home directory")
			configHome = ""
		}
	}
	// Use CWD
	if configHome == "" {
		configHome = "."
	}
	return configHome
}
