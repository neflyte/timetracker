package main

import (
	"flag"
	"fmt"

	"github.com/neflyte/timetracker/cmd/timetracker-gui/cmd"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/startup"
	"github.com/neflyte/timetracker/lib/ui/gui"
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
	guiCmdOptionShowReportWindow         bool
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
	flag.BoolVar(&guiCmdOptionShowReportWindow, "report", false, "Shows the Report Window")
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
	// TODO: validate GUI parameters; throw error if mutually exclusive parameters are specified
	doGUI()
}

func doGUI() {
	app := gui.InitGUI(cmd.AppVersion)
	switch {
	case guiCmdOptionStopRunningTask:
		gui.ShowTimetrackerWindowAndStopRunningTask()
	case guiCmdOptionShowManageWindow:
		gui.ShowTimetrackerWindowWithManageWindow()
	case guiCmdOptionShowAboutWindow:
		gui.ShowTimetrackerWindowWithAbout()
	case guiCmdOptionShowCreateAndStartDialog:
		gui.ShowTimetrackerWindowAndShowCreateAndStartDialog()
	case guiCmdOptionShowReportWindow:
		gui.ShowTimetrackerWindowWithReportWindow()
	default:
		gui.ShowTimetrackerWindow()
	}
	// Start the GUI
	gui.StartGUI(app)
}
