package tray

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/nightlyone/lockfile"
	"os"
	"path"
	"time"
)

const (
	TimesheetStatusIdle = iota
	TimesheetStatusRunning
	TimesheetStatusError

	trayPidfile            = "timetracker-tray.pid"
	statusLoopDelaySeconds = 5
)

var (
	mStatus  *systray.MenuItem
	mAbout   *systray.MenuItem
	mQuit    *systray.MenuItem
	lockFile lockfile.Lockfile
	pidPath  string
)

func Run() (err error) {
	log := logger.GetLogger("tray.Run")
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = "."
		err = nil // Make sure we can't accidentally return a non-nil error in this casse
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
	return
}

func onReady() {
	var err error

	log := logger.GetLogger("tray.onReady")
	lockFile, err = lockfile.New(pidPath)
	if err != nil {
		log.Err(err).Msgf("error creating pidfile; pidPath=%s", pidPath)
		return
	}
	err = lockFile.TryLock()
	if err != nil {
		log.Err(err).Msgf("error locking pidfile; pidPath=%s", pidPath)
		systray.Quit()
		return
	}
	log.Debug().Msgf("locked pidfile %s", pidPath)
	systray.SetTitle("Timetracker")
	systray.SetTooltip("Timetracker")
	systray.SetTemplateIcon(icons.Check, icons.Check)
	mStatus = systray.AddMenuItem("(idle)", "Timetracker task status")
	systray.AddSeparator()
	mAbout = systray.AddMenuItem("About Timetracker", "About the Timetracker app")
	mQuit = systray.AddMenuItem("Quit", "Quit the Timetracker tray app")
	// Start the status loop in a goroutine
	go statusLoop()
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
	log := logger.GetLogger("tray.mainLoop")
	log.Debug().Msg("starting main loop")
	for {
		select {
		case <-mStatus.ClickedCh:
			log.Debug().Msg("status menu item selected")
		case <-mAbout.ClickedCh:
			log.Debug().Msg("about menu item selected")
		case <-mQuit.ClickedCh:
			log.Debug().Msg("quit menu item selected; quitting systray")
			systray.Quit()
			return
		}
	}
}

func statusLoop() {
	var err error
	var timesheets []models.TimesheetData

	log := logger.GetLogger("tray.statusLoop")
	log.Debug().Msg("starting status loop")
	lastState := TimesheetStatusIdle // Start with idle state
	tsd := new(models.TimesheetData)
	for {
		log.Debug().Msg("getting running timesheet")
		timesheets, err = models.Timesheet(tsd).SearchOpen()
		if err != nil {
			log.Err(err).Msg("error getting running timesheet")
			if lastState != TimesheetStatusError {
				// Show error icon
				systray.SetTemplateIcon(icons.Error, icons.Error)
				// Update status menu to show `[error]`
				mStatus.SetTitle("[error]")
				lastState = TimesheetStatusError
			}
		} else {
			if len(timesheets) == 0 {
				// No running task
				log.Debug().Msg("no running task")
				if lastState != TimesheetStatusIdle {
					// Show check icon
					systray.SetTemplateIcon(icons.Check, icons.Check)
					// Update status menu to show `(idle)`
					mStatus.SetTitle("(idle)")
					lastState = TimesheetStatusIdle
				}
			} else {
				// Running task...
				log.Debug().Msgf("running task: %#v", timesheets[0])
				// Update status menu item to show task ID and duration
				statusText := fmt.Sprintf(
					"%s %s",
					timesheets[0].Task.Synopsis,
					time.Since(timesheets[0].StartTime).Truncate(time.Second).String(),
				)
				mStatus.SetTitle(statusText)
				if lastState != TimesheetStatusRunning {
					// Show running icon
					systray.SetTemplateIcon(icons.Running, icons.Running)
					lastState = TimesheetStatusRunning
				}
			}
		}
		// Delay
		log.Debug().Msgf("delaying %d seconds until next timesheet check", statusLoopDelaySeconds)
		<-time.After(statusLoopDelaySeconds * time.Second)
	}
}
