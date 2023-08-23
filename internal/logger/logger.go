package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
)

const (
	logFileName           = "timetracker.log"
	logDirectoryMode      = 0755
	logFileMode           = 0644
	emergencyErrorLogMode = 0600
)

var (
	// rootLogger is the application root logger instance
	rootLogger zerolog.Logger

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

// InitLogger initializes the logger system
func InitLogger(logLevel string, console bool) {
	var err error

	if loggerInitialized {
		return
	}
	configHome := getConfigHome()
	logPath = path.Join(configHome, "timetracker")
	err = os.MkdirAll(logPath, logDirectoryMode)
	if err != nil {
		EmergencyLogToFile(
			"logger_error.txt",
			fmt.Sprintf("error creating log directory %s: %s\n", logPath, err.Error()),
		)
	}
	logFilePath = path.Join(logPath, logFileName)
	logFileHandle, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, logFileMode)
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
	if len(logWriters) == 0 {
		EmergencyLogToFile(
			"logger_error.txt",
			"len(logWriters) == 0; this is unexpected",
		)
	}
	multi := zerolog.MultiLevelWriter(logWriters...)
	rootLogger = zerolog.New(multi).With().Timestamp().Logger()
	// Set global logger message level
	lvl, ok := levelMap[logLevel]
	if !ok {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	loggerInitialized = true
}

// CleanupLogger cleans up the logger system before the app exits
func CleanupLogger() {
	if loggerInitialized {
		if logFileHandle != nil {
			err := logFileHandle.Close()
			if err != nil {
				EmergencyLogToFile(
					"logger_cleanup_error.txt",
					fmt.Sprintf("error cleaning up logger: %s\n", err.Error()),
				)
			}
			logFileHandle = nil
		}
		loggerInitialized = false
	}
}

// GetLogger returns a new logger with a string field `func` set to the supplied funcName
func GetLogger(funcName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).
			With().Timestamp().Str("func", funcName).Logger().
			Level(zerolog.TraceLevel)
	}
	return rootLogger.With().Str("func", funcName).Logger()
}

// GetStructLogger returns a new logger with a string field `struct` set to the supplied structname
func GetStructLogger(structName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).
			With().Timestamp().Str("struct", structName).Logger().
			Level(zerolog.TraceLevel)
	}
	return rootLogger.With().Str("struct", structName).Logger()
}

// GetPackageLogger returns a new logger with a string field `package` set to the supplied packageName
func GetPackageLogger(packageName string) zerolog.Logger {
	if !loggerInitialized {
		return zerolog.New(&zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli}).
			With().Timestamp().Str("package", packageName).Logger().
			Level(zerolog.TraceLevel)
	}
	return rootLogger.With().Str("package", packageName).Logger()
}

// GetFuncLogger returns a new logger based on the supplied logger with a string field `func` set to the supplied funcName
func GetFuncLogger(existingLog zerolog.Logger, funcName string) zerolog.Logger {
	return existingLog.With().Str("func", funcName).Logger()
}

// getConfigHome attempts to determine the user's configuration files directory
func getConfigHome() string {
	log := GetLogger("GetConfigHome")
	// Look for XDG_CONFIG_HOME
	configHome, err := os.UserConfigDir()
	if err != nil {
		log.Err(err).
			Msg("error getting user config directory")
		configHome = ""
	}
	// Look for HOME
	if configHome == "" {
		configHome, err = os.UserHomeDir()
		if err != nil {
			log.Err(err).
				Msg("error getting user home directory")
			configHome = ""
		}
	}
	// Use CWD
	if configHome == "" {
		configHome = "."
	}
	return configHome
}

// EmergencyLogToFile logs the specified message to the specified file. To be used when
// the logger facility is not available.
func EmergencyLogToFile(filename string, message string) {
	_ = os.WriteFile(filename, []byte(message), emergencyErrorLogMode) //nolint:errcheck
}
