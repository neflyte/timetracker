package tray

import (
	"github.com/getlantern/systray"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/nightlyone/lockfile"
	"os"
	"path"
)

const (
	trayPidfile = "timetracker-tray.pid"
)

var (
	// mStatus *systray.MenuItem
	mQuit    *systray.MenuItem
	lockFile lockfile.Lockfile
	pidPath  string
)

func Run() (err error) {
	log := logger.GetLogger("tray.Run")
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = "."
	} else {
		userConfigDir = path.Join(userConfigDir, "timetracker")
	}
	if userConfigDir != "." {
		err = os.MkdirAll(userConfigDir, 0755)
		if err != nil {
			log.Err(err).Msgf("error creating directories for pidfile; userConfigDir=%s", userConfigDir)
			return
		}
	}
	pidPath = path.Join(userConfigDir, trayPidfile)
	systray.Run(onReady, onExit)
	return nil
}

func onReady() {
	var err error

	log := logger.GetLogger("tray.onReady")
	lockFile, err = lockfile.New(pidPath)
	if err != nil {
		log.Err(err).Msgf("error creating pidfile; pidPath=%s", pidPath)
		return
	}
	log.Debug().Msgf("locked pidfile %s", pidPath)
	systray.SetTitle("Timetracker")
	systray.SetTooltip("Timetracker")
	systray.SetTemplateIcon(icons.Check, icons.Check)
	/* mStatus */ _ = systray.AddMenuItem("(idle)", "Timetracker task status")
	systray.AddSeparator()
	mQuit = systray.AddMenuItem("Quit", "Quit the Timetracker tray app")
	// Start the main loop
	mainLoop()
}

func onExit() {
	log := logger.GetLogger("tray.onExit")
	err := lockFile.Unlock()
	if err != nil {
		log.Err(err).Msgf("error releasing pidfile")
		return
	}
	log.Debug().Msg("unlocked pidfile")
}

func mainLoop() {
	for {
		select {
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}
