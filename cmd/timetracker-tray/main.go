package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/neflyte/timetracker/cmd/timetracker-tray/cmd"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/startup"
	"github.com/neflyte/timetracker/lib/ui/tray"
	"github.com/neflyte/timetracker/lib/utils"
)

var (
	configFileName       string
	logLevel             string
	showVersion          bool
	console              bool
	darwinEnableLaunchd  bool
	darwinDisableLaunchd bool
)

func init() {
	flag.StringVar(&configFileName, "config", "", "Specify the full path and filename of the database to use")
	flag.StringVar(&logLevel, "logLevel", constants.DefaultLogLevel, "Specify the logging level")
	flag.BoolVar(&showVersion, "version", false, "Display the program version")
	flag.BoolVar(&console, "console", false, "Log to the console")
	// macOS-only flags to manage launchd
	if runtime.GOOS == "darwin" {
		flag.BoolVar(&darwinEnableLaunchd, "enableLaunchd", false, "Start Timetracker Tray at login")
		flag.BoolVar(&darwinDisableLaunchd, "disableLaunchd", false, "Do not start Timetracker Tray at login")
	}
}

func main() {
	flag.Parse()
	if showVersion {
		fmt.Printf("timetracker-tray %s\n", cmd.AppVersion)
		return
	}

	startup.SetLogLevel(logLevel)
	startup.SetConsole(console)
	startup.InitLogger()
	defer startup.CleanupLogger()

	if runtime.GOOS == "darwin" {
		doDarwinLaunchdCommands(darwinEnableLaunchd, darwinDisableLaunchd)
	}

	startup.SetDatabaseFileName(configFileName)
	startup.InitDatabase()
	defer startup.CleanupDatabase()
	tray.Run(nil)
}

func doDarwinLaunchdCommands(enable bool, disable bool) {
	log := logger.GetLogger("doDarwinLaunchdCommands")
	if enable && disable {
		log.Fatal().
			Msg("Cannot specify both -enableLaunchd and -disableLaunchd at the same time")
	}
	if enable {
		err := utils.EnableLaunchd()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Cannot start Timetracker Tray at login")
		}
		os.Exit(0)
	}
	if disable {
		err := utils.DisableLaunchd()
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Cannot disable Timetracker Tray at login")
		}
		os.Exit(0)
	}
}
