package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/neflyte/timetracker/cmd/timetracker-tray/cmd"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/startup"
	"github.com/neflyte/timetracker/lib/ui/tray"
	"github.com/neflyte/timetracker/lib/ui/tray/autostart"
)

var (
	configFileName   string
	logLevel         string
	showVersion      bool
	console          bool
	enableAutostart  bool
	disableAutostart bool
)

func init() {
	flag.StringVar(&configFileName, "config", "", "Specify the full path and filename of the database to use")
	flag.StringVar(&logLevel, "logLevel", constants.DefaultLogLevel, "Specify the logging level")
	flag.BoolVar(&showVersion, "version", false, "Display the program version")
	flag.BoolVar(&console, "console", false, "Log to the console")
	flag.BoolVar(&enableAutostart, "enableAutostart", false, "Start Timetracker Tray at login")
	flag.BoolVar(&disableAutostart, "disableAutostart", false, "Do not start Timetracker Tray at login")
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

	doHandleAutostart()

	startup.SetDatabaseFileName(configFileName)
	startup.InitDatabase()
	defer startup.CleanupDatabase()
	tray.Run(nil)
}

func doHandleAutostart() {
	if !enableAutostart && !disableAutostart {
		return
	}
	log := logger.GetLogger("doHandleAutostart")
	if enableAutostart && disableAutostart {
		log.Error().
			Msg("Cannot specify both -enableAutostart and -disableAutostart at the same time")
		os.Exit(1)
	}
	if enableAutostart && !autostart.IsEnabled() {
		err := autostart.Enable()
		if err != nil {
			log.Err(err).
				Msg("Cannot enable autostart")
			os.Exit(1)
		}
	}
	if disableAutostart && autostart.IsEnabled() {
		err := autostart.Disable()
		if err != nil {
			log.Err(err).
				Msg("Cannot disable autostart")
			os.Exit(1)
		}
	}
	os.Exit(0)
}
