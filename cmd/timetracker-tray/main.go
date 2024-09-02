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

	if enableAutostart || disableAutostart {
		doHandleAutostart()
		return
	}

	startup.SetDatabaseFileName(configFileName)
	startup.InitDatabase()
	defer startup.CleanupDatabase()
	tray.Run(nil)
}

func doHandleAutostart() {
	log := logger.GetLogger("doHandleAutostart").
		With().
		Bool("enableAutostart", enableAutostart).
		Bool("disableAutostart", disableAutostart).
		Logger()
	log.Debug().Msg("handle autostart")
	if enableAutostart && disableAutostart {
		log.Error().
			Msg("Cannot specify both -enableAutostart and -disableAutostart at the same time")
		os.Exit(1)
	}
	isEnabled := autostart.IsEnabled()
	log = log.With().Bool("isEnabled", isEnabled).Logger()
	if enableAutostart && !isEnabled {
		log.Debug().
			Msg("enable autostart")
		err := autostart.Enable()
		if err != nil {
			log.Err(err).
				Msg("Cannot enable autostart")
			os.Exit(1)
		}
		log.Debug().
			Msg("autostart enabled")
	}
	if disableAutostart && isEnabled {
		log.Debug().
			Msg("disable autostart")
		err := autostart.Disable()
		if err != nil {
			log.Err(err).
				Msg("Cannot disable autostart")
			os.Exit(1)
		}
		log.Debug().
			Msg("autostart disabled")
	}
}
