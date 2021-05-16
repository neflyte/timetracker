package tray

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/nightlyone/lockfile"
	"os"
	"path"
	"sync"
	"time"
)

const (
	trayPidfile = "timetracker-tray.pid"
)

var (
	mStatus            *systray.MenuItem
	mAbout             *systray.MenuItem
	mQuit              *systray.MenuItem
	lockFile           lockfile.Lockfile
	pidPath            string
	wg                 sync.WaitGroup
	trayQuitChan       chan bool
	actionLoopQuitChan chan bool
)

func Run() (err error) {
	log := logger.GetLogger("tray.Run")
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = "."
		err = nil // Make sure we can't accidentally return a non-nil error in this case
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
	log.Trace().Msgf("pidPath=%s", pidPath)

	// Start the ActionLoop
	actionLoopQuitChan = make(chan bool, 1)
	go appstate.ActionLoop(actionLoopQuitChan)
	// Start the systray in a goroutine
	wg = sync.WaitGroup{}
	wg.Add(1)
	log.Trace().Msg("go systray.Run(...)")
	go systray.Run(onReady, onExit)
	// Wait for the systray to finish initializing
	log.Debug().Msg("waiting for systray")
	wg.Wait()
	log.Debug().Msg("systray initialized")
	// Start mainLoop
	trayQuitChan = make(chan bool, 1)
	log.Trace().Msg("go mainLoop(...)")
	go mainLoop(trayQuitChan)
	// Start GUI
	log.Trace().Msg("gui.StartGUI()")
	gui.StartGUI()
	log.Trace().Msg("gui has finished")
	// Shut down mainLoop
	trayQuitChan <- true
	// Shut down ActionLoop
	actionLoopQuitChan <- true
	// Shut down systray
	systray.Quit()
	return
}

func onReady() {
	var err error

	log := logger.GetLogger("tray.onReady")
	log.Trace().Msg("starting")
	defer func() {
		// Signal that we're initialized
		log.Debug().Msg("signalling true")
		wg.Done()
	}()
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
	log.Trace().Msg("setting up observables")
	appstate.ObsRunningTimesheet.ForEach(
		func(item interface{}) {
			tsd, ok := item.(*models.TimesheetData)
			if ok {
				if tsd == nil {
					// No running timesheet
					log.Trace().Msg("got nil running timesheet item")
					systray.SetTemplateIcon(icons.Check, icons.Check)
					mStatus.SetTitle("(idle)")
					return
				}
				log.Trace().Msg("got non-running timesheet object")
				systray.SetTemplateIcon(icons.Running, icons.Running)
				statusText := fmt.Sprintf(
					"%s %s",
					tsd.Task.Synopsis,
					time.Since(tsd.StartTime).Truncate(time.Second).String(),
				)
				mStatus.SetTitle(statusText)
			}
		},
		func(err error) {
			systray.SetTemplateIcon(icons.Error, icons.Error)
			mStatus.SetTitle("[error]")
		},
		func() {
			log.Trace().Msg("running timesheet observable is done")
		},
	)
	log.Trace().Msg("done")
}

func onExit() {
	funcLog := logger.GetLogger("tray.onExit")
	funcLog.Trace().Msg("started")
	err := lockFile.Unlock()
	if err != nil {
		funcLog.Err(err).Msgf("error releasing pidfile")
		return
	}
	funcLog.Debug().Msg("unlocked pidfile")
	funcLog.Trace().Msg("done")
}

func mainLoop(quitChan chan bool) {
	funcLog := logger.GetLogger("tray.mainLoop")
	// Start main loop
	funcLog.Trace().Msg("starting")
	for {
		select {
		case <-mStatus.ClickedCh:
			funcLog.Debug().Msg("status menu item selected")
			switch appstate.GetLastState() {
			case constants.TimesheetStatusRunning:
				runningTimesheet := appstate.GetRunningTimesheet()
				gui.ShowTimetrackerWindowWithConfirm(
					"Stop running task?",
					fmt.Sprintf(
						"Stop task %s (%s) ?",
						runningTimesheet.Task.Synopsis,
						time.Since(runningTimesheet.StartTime).Truncate(time.Second).String(),
					),
					func(res bool) {
						if res {
							// Stop the running task
							funcLog.Debug().Msgf("stopping task %s", runningTimesheet.Task.Synopsis)
							err := models.Task(new(models.TaskData)).StopRunningTask()
							if err != nil {
								funcLog.Err(err).Msg(errors.StopRunningTaskError)
							}
							// Get a new timesheet and update the appstate
							appstate.UpdateRunningTimesheet()
						}
					},
					false,
				)
			case constants.TimesheetStatusError:
				gui.ShowTimetrackerWindowWithError(appstate.GetStatusError())
			case constants.TimesheetStatusIdle:
				gui.ShowTimetrackerWindow()
			}
		case <-mAbout.ClickedCh:
			funcLog.Debug().Msg("about menu item selected; showing about dialog")
			gui.ShowTimetrackerWindowWithAbout()
		case <-mQuit.ClickedCh:
			funcLog.Debug().Msg("quit menu item selected; quitting app")
			gui.StopGUI()
			return
		case <-quitChan:
			funcLog.Debug().Msg("quit channel fired; exiting function")
			return
		}
	}
}
