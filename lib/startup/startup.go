package startup

import (
	"os"
	"path"

	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/database"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
)

var (
	startupLogger    = logger.GetPackageLogger("startup")
	databaseFileName = constants.DefaultDatabaseFileName
	logLevel         = constants.DefaultLogLevel
	logConsole       = false
)

// SetDatabaseFileName sets the file name of the database file
func SetDatabaseFileName(databaseFile string) {
	databaseFileName = databaseFile
}

// SetLogLevel sets the logger level
func SetLogLevel(level string) {
	logLevel = level
}

// SetConsole sets a flag that logs messages to the console
func SetConsole(logToConsole bool) {
	logConsole = logToConsole
}

// InitDatabase initializes the database system
func InitDatabase() {
	log := logger.GetFuncLogger(startupLogger, "InitDatabase")
	configFile := databaseFileName
	if configFile == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			log.Err(err).
				Msg("error getting user config dir")
			userConfigDir = "."
		} else {
			userConfigDir = path.Join(userConfigDir, "timetracker")
			// Make sure this directory exists...
			mkdirErr := os.MkdirAll(userConfigDir, constants.ConfigDirectoryMode)
			if mkdirErr != nil {
				log.Fatal().
					Err(mkdirErr).
					Msg("error creating configuration directory")
				return
			}
		}
		configFile = path.Join(userConfigDir, constants.DefaultDatabaseFileName)
	}
	log.Debug().
		Str("configFile", configFile).
		Msg("resolved config file")
	db, err := database.Open(configFile)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("database", configFile).
			Msg("error opening database")
	}
	log.Debug().Msg("database opened")
	database.Set(db)
	err = db.AutoMigrate(new(models.TaskData), new(models.TimesheetData))
	if err != nil {
		CleanupDatabase()
		log.Fatal().
			Err(err).
			Msg("error auto-migrating database schema")
		return
	}
	log.Debug().Msg("schema migrated (if necessary)")
}

// CleanupDatabase tears down the database system
func CleanupDatabase() {
	database.Close(database.Get())
	database.Set(nil)
}

// InitLogger initialized the logger system
func InitLogger() {
	logger.InitLogger(logLevel, logConsole)
	startupLogger = logger.GetPackageLogger("startup")
}

// CleanupLogger tears down the logger system
func CleanupLogger() {
	logger.CleanupLogger()
}
