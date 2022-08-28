package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
)

const (
	defaultDatabaseFileName = "timetracker.db"
	configDirectoryMode     = 0755
)

var (
	// AppVersion is the application version number; it must always be exported
	AppVersion = "dev"

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
	rootCmd.AddCommand(taskCmd, timesheetCmd, statusCmd, trayCmd, guiCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf("timetracker %s\n", AppVersion))
}

// Execute is the main entry point for the CLI
func Execute() {
	log := logger.GetLogger("Execute")
	// On macOS, if no CLI parameters were specified then default to starting the GUI
	if runtime.GOOS == "darwin" && len(os.Args) < 2 {
		rootCmd.SetArgs([]string{"gui"})
	}
	err := rootCmd.Execute()
	if err != nil {
		log.Err(err).Msg("error executing root command")
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
			mkdirErr := os.MkdirAll(userConfigDir, configDirectoryMode)
			if mkdirErr != nil {
				log.Fatal().Msgf("error creating configuration directory: %s", mkdirErr.Error())
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
	database.Set(db)
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
	database.Close(database.Get())
	database.Set(nil)
	logger.CleanupLogger()
}
