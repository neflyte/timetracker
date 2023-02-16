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
	guiPidfile              = "timetracker-gui.pid"
)

var (
	AppVersion                           = "dev" // AppVersion is the application version number; it must always be exported
	configFileName                       string
	logLevel                             string
	showVersion                          bool
	guiCmdOptionStopRunningTask          bool
	guiCmdOptionShowCreateAndStartDialog bool
	guiCmdOptionShowManageWindow         bool
	guiCmdOptionShowAboutWindow          bool
)

func init() {
	flag.StringVar(&configFileName, "config", "", "Specify the full path and filename of the database to use")
	flag.StringVar(&logLevel, "logLevel", "info", "Specify the logging level")
	flag.BoolVar(&showVersion, "version", false, "Display the program version")
	// GUI flags
	flag.BoolVar(&guiCmdOptionStopRunningTask, "stop-running-task", false, "Stops the running task, if any")
	flag.BoolVar(&guiCmdOptionShowCreateAndStartDialog, "create-and-start", false, "Shows the Create and Start New Task dialog")
	flag.BoolVar(&guiCmdOptionShowManageWindow, "manage", false, "Shows the Manage Window")
	flag.BoolVar(&guiCmdOptionShowAboutWindow, "about", false, "Shows the About Window")
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
		fmt.Printf("timetracker-gui %s\n", AppVersion)
		return
	}
	initLogger()
	initDatabase()
	defer cleanUp()
	log := logger.GetLogger("main")
	err := preDoGUI()
	if err != nil {
		log.Err(err).
			Msg("error setting up GUI")
		return
	}
	defer func() {
		err = postDoGUI()
		if err != nil {
			log.Err(err).
				Msg("error tearing down GUI")
		}
	}()
	doGUI()
}
