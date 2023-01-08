package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
)

const (
	defaultDatabaseFileName = "timetracker.db"
	configDirectoryMode     = 0755
	minimumArgumentCount    = 2 // minimumArgumentCount is the minimum number of arguments needed to assume the GUI is not being requested
)

var (
	AppVersion = "dev" // AppVersion is the application version number; it must always be exported
	rootCmd    = &cobra.Command{
		Version:           AppVersion,
		Use:               "timetracker",
		Short:             "A simple time tracker",
		Long:              "A simple time tracker for various tasks with basic reporting",
		PersistentPostRun: cleanUp,
	}
	configFileName string
	logLevel       string
	consoleLogging bool
)

func init() {
	cobra.OnInitialize(initLogger, initDatabase)
	rootCmd.PersistentFlags().StringVarP(&configFileName, "config", "c", "", "Specify the full path and filename of the database to use")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "l", "info", "Specify the logging level")
	rootCmd.PersistentFlags().BoolVar(&consoleLogging, "console", false, "Log messages to the console as well as the log file")
	rootCmd.AddCommand(taskCmd, timesheetCmd, statusCmd, trayCmd, guiCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf("timetracker %s\n", AppVersion))
}

// Execute is the main entry point for the CLI
func Execute() {
	// If there were no CLI parameters supplied...
	if len(os.Args) < minimumArgumentCount {
		forceRunGUI()
	}
	if runtime.GOOS == "windows" && os.Args[1] != "gui" {
		isAttached, err := cli.AttachToParentConsole()
		if err != nil {
			logger.EmergencyLogToFile("attachconsole_failed.txt", err.Error())
			return
		}
		if !isAttached {
			logger.EmergencyLogToFile("attachconsole_notattached.txt", "not attached")
		}
	}
	err := rootCmd.Execute()
	if err != nil {
		logger.EmergencyLogToFile("rootcmd_error.txt", err.Error())
		os.Exit(1)
	}
}

// forceRunGUI forces the GUI to be started on select platforms
func forceRunGUI() {
	switch runtime.GOOS {
	case "darwin":
		// On macOS, default to starting the GUI
		rootCmd.SetArgs([]string{"gui"})
	case "windows":
		// On windows, start the GUI directly without cobra
		initLogger()
		initDatabase()
		err := preDoGUI(nil, nil)
		if err != nil {
			logger.EmergencyLogToFile("execute_error.txt", err.Error())
			cleanUp(nil, nil)
			os.Exit(1)
		}
		err = doGUI(nil, nil)
		if err != nil {
			logger.EmergencyLogToFile("execute_error.txt", err.Error())
			cleanUp(nil, nil)
			os.Exit(1)
		}
		err = postDoGUI(nil, nil)
		if err != nil {
			logger.EmergencyLogToFile("execute_error.txt", err.Error())
			cleanUp(nil, nil)
			os.Exit(1)
		}
		cleanUp(nil, nil)
		return
	}
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
		cleanUp(nil, nil)
		log.Fatal().
			Err(err).
			Msg("error auto-migrating database schema")
		return
	}
	log.Debug().Msg("schema migrated (if necessary)")
}

func initLogger() {
	logger.InitLogger(logLevel, consoleLogging)
}

func cleanUp(_ *cobra.Command, _ []string) {
	database.Close(database.Get())
	database.Set(nil)
	logger.CleanupLogger()
}
