package main

import (
	"flag"
	"fmt"

	"github.com/neflyte/timetracker/cmd/timetracker-tray/cmd"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/startup"
)

const (
	trayPidfile = "timetracker-tray.pid"
)

var (
	configFileName string
	logLevel       string
	showVersion    bool
	console        bool
)

func init() {
	flag.StringVar(&configFileName, "config", "", "Specify the full path and filename of the database to use")
	flag.StringVar(&logLevel, "logLevel", constants.DefaultLogLevel, "Specify the logging level")
	flag.BoolVar(&showVersion, "version", false, "Display the program version")
	flag.BoolVar(&console, "console", false, "Log to the console")
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
	startup.SetDatabaseFileName(configFileName)
	startup.InitDatabase()
	defer startup.CleanupDatabase()
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
