package cmd

import (
	"fmt"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"os"
	"path"
)

const (
	defaultDatabaseFileName = "timetracker.db"
)

var (
	// AppVersion is the application version number
	AppVersion = ""

	rootCmd = &cobra.Command{
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
	rootCmd.AddCommand(taskCmd, timesheetCmd, statusCmd, trayCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf("timetracker %s", AppVersion))
}

func Execute() {
	log := logger.GetLogger("Execute")
	if err := rootCmd.Execute(); err != nil {
		log.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func initDatabase() {
	log := logger.GetLogger("initDatabase")
	configFile := configFileName
	if configFile == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			log.Printf("error getting user config dir: %s\n", err)
			userConfigDir = "."
		} else {
			userConfigDir = path.Join(userConfigDir, "timetracker")
			// Make sure this directory exists...
			err := os.MkdirAll(userConfigDir, 0755)
			if err != nil {
				log.Fatal().Msgf("error creating configuration directory: %s\n", err)
			}
		}
		configFile = path.Join(userConfigDir, defaultDatabaseFileName)
	}
	log.Printf("configFile=%s\n", configFile)
	db, err := database.Open(configFile)
	if err != nil {
		log.Fatal().Msgf("error opening database at %s: %s\n", configFile, err)
	}
	log.Printf("database opened")
	database.DB = db
	err = db.AutoMigrate(new(models.TaskData), new(models.TimesheetData))
	if err != nil {
		cleanUp(nil, nil)
		log.Fatal().Msgf("error auto-migrating database schema: %s\n", err)
	}
	log.Printf("schema migrated (if necessary)")
}

func initLogger() {
	logger.InitLogger(logLevel, consoleLogging)
}

func cleanUp(_ *cobra.Command, _ []string) {
	database.Close(database.DB)
	database.DB = nil
	logger.CleanupLogger()
}
