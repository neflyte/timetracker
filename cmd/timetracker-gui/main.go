package main

import (
	"flag"
	"fmt"

	"github.com/neflyte/timetracker/cmd/timetracker-gui/cmd"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/startup"
)

const (
	guiPidfile = "timetracker-gui.pid"
)

var (
	configFileName                       string
	logLevel                             string
	showVersion                          bool
	consoleLogging                       bool
	guiCmdOptionStopRunningTask          bool
	guiCmdOptionShowCreateAndStartDialog bool
	guiCmdOptionShowManageWindow         bool
	guiCmdOptionShowAboutWindow          bool
)

func init() {
	flag.StringVar(&configFileName, "config", "", "Specify the full path and filename of the database to use")
	flag.StringVar(&logLevel, "logLevel", constants.DefaultLogLevel, "Specify the logging level")
	flag.BoolVar(&consoleLogging, "console", false, "Also log messages to the console")
	flag.BoolVar(&showVersion, "version", false, "Display the program version")
	// GUI flags
	flag.BoolVar(&guiCmdOptionStopRunningTask, "stop-running-task", false, "Stops the running task, if any")
	flag.BoolVar(&guiCmdOptionShowCreateAndStartDialog, "create-and-start", false, "Shows the Create and Start New Task dialog")
	flag.BoolVar(&guiCmdOptionShowManageWindow, "manage", false, "Shows the Manage Window")
	flag.BoolVar(&guiCmdOptionShowAboutWindow, "about", false, "Shows the About Window")
}

func main() {
	flag.Parse()
	if showVersion {
		fmt.Printf("timetracker-gui %s\n", cmd.AppVersion)
		return
	}
	startup.SetLogLevel(logLevel)
	startup.SetConsole(consoleLogging)
	startup.InitLogger()
	defer startup.CleanupLogger()
	startup.SetDatabaseFileName(configFileName)
	startup.InitDatabase()
	defer startup.CleanupDatabase()
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
