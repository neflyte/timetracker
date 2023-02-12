package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
)

const (
	defaultDatabaseFileName = "timetracker.db"
	configDirectoryMode     = 0755
	trayPidfile             = "timetracker-tray.pid"
)

var (
	AppVersion     = "dev" // AppVersion is the application version number; it must always be exported
	configFileName string
	logLevel       string
	showVersion    bool
)

func init() {
	flag.StringVar(&configFileName, "config", "", "Specify the full path and filename of the database to use")
	flag.StringVar(&logLevel, "logLevel", "INFO", "Specify the logging level")
	flag.BoolVar(&showVersion, "version", false, "Display the program version")
}

func initDatabase() {
	log := logger.GetLogger("initDatabase")
	configFile := configFileName
	if configFile == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			log.Err(err).
				Msg("error getting user config dir")
			userConfigDir = "."
		} else {
			userConfigDir = path.Join(userConfigDir, "timetracker")
			// Make sure this directory exists...
			mkdirErr := os.MkdirAll(userConfigDir, configDirectoryMode)
			if mkdirErr != nil {
				log.Fatal().
					Err(mkdirErr).
					Msg("error creating configuration directory")
				return
			}
		}
		configFile = path.Join(userConfigDir, defaultDatabaseFileName)
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
		cleanUp()
		log.Fatal().
			Err(err).
			Msg("error auto-migrating database schema")
		return
	}
	log.Debug().Msg("schema migrated (if necessary)")
}

func initLogger() {
	logger.InitLogger(logLevel, false)
}

func cleanUp() {
	database.Close(database.Get())
	database.Set(nil)
	logger.CleanupLogger()
}

func main() {
	flag.Parse()
	if showVersion {
		fmt.Printf("timetracker-tray %s\n", AppVersion)
		return
	}
	initLogger()
	initDatabase()
	defer cleanUp()
	log := logger.GetLogger("main")
	err := preDoTray()
	if err != nil {
		log.Err(err).
			Msg("error setting up tray entry")
		return
	}
	defer func() {
		err = postDoTray()
		if err != nil {
			log.Err(err).
				Msg("error tearing down tray entry")
		}
	}()
	doTray()
}
