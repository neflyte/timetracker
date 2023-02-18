package cmd

import (
	"fmt"
	"os"

	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/startup"
	"github.com/spf13/cobra"
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
	cobra.OnInitialize(initialize)
	rootCmd.PersistentFlags().StringVarP(&configFileName, "config", "c", "", "Specify the full path and filename of the database to use")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "l", "info", "Specify the logging level")
	rootCmd.PersistentFlags().BoolVar(&consoleLogging, "console", false, "Log messages to the console as well as the log file")
	rootCmd.AddCommand(taskCmd, timesheetCmd, statusCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf("timetracker %s\n", AppVersion))
}

// Execute is the main entry point for the CLI
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.EmergencyLogToFile("rootcmd_error.txt", err.Error())
		os.Exit(1)
	}
}

func initialize() {
	startup.SetLogLevel(logLevel)
	startup.SetConsole(consoleLogging)
	startup.InitLogger()
	startup.SetDatabaseFileName(configFileName)
	startup.InitDatabase()
}

func cleanUp(_ *cobra.Command, _ []string) {
	startup.CleanupDatabase()
	startup.CleanupLogger()
}
