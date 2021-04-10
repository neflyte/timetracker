package tray

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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
	trayPidfile            = "timetracker-tray.pid"
	statusLoopDelaySeconds = 5
)

var (
	mStatus      *systray.MenuItem
	mAbout       *systray.MenuItem
	mQuit        *systray.MenuItem
	lockFile     lockfile.Lockfile
	pidPath      string
	wg           sync.WaitGroup
	FyneApp      fyne.App
	trayQuitChan chan bool
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
	// Set up fyne
	log.Trace().Msg("setting up FyneApp")
	FyneApp = app.New()
	// Create the main timetracker window
	log.Trace().Msg("creating timetracker window")
	ttw := gui.TimetrackerWindow(&FyneApp)
	if ttw != nil {
		log.Trace().Msg("set ttw as master, then hide it")
		// Set the window as the master and hide it
		(*ttw).SetMaster()
		(*ttw).Hide()
	}
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
	go mainLoop(trayQuitChan, FyneApp)
	// Start window closer
	log.Trace().Msg("starting window closer")
	err = gui.StartWindowCloser()
	if err != nil {
		log.Err(err).Msg("error starting window closer")
	}
	// Start fyne
	log.Trace().Msg("FyneApp.Run()")
	FyneApp.Run()
	log.Trace().Msg("FyneApp finished")
	// Stop window closer
	err = gui.StopWindowCloser()
	if err != nil {
		log.Err(err).Msg("error stopping window closer")
	}
	// Shut down mainLoop
	trayQuitChan <- true
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
	log.Trace().Msg("done")
}

func onExit() {
	log := logger.GetLogger("tray.onExit")
	log.Trace().Msg("started")
	err := lockFile.Unlock()
	if err != nil {
		log.Err(err).Msgf("error releasing pidfile")
		return
	}
	log.Debug().Msg("unlocked pidfile")
	log.Trace().Msg("done")
}

func mainLoop(quitChan chan bool, app fyne.App) {
	log := logger.GetLogger("tray.mainLoop")
	log.Trace().Msg("starting")
	statusLoopQuitChan := make(chan bool, 1)
	// Start the status loop in a goroutine
	log.Trace().Msg("go statusLoop(...)")
	go statusLoop(statusLoopQuitChan)
	// Start main loop
	for {
		select {
		case <-mStatus.ClickedCh:
			log.Debug().Msg("status menu item selected")
			switch appstate.GetLastState() {
			case constants.TimesheetStatusRunning:
				runningTimesheet := appstate.GetRunningTimesheet()
				gui.NewConfirmDialogWindow(
					app,
					"Stop running task?",
					fmt.Sprintf(
						"Stop task %s (%s)",
						runningTimesheet.Task.Synopsis,
						time.Since(runningTimesheet.StartTime).Truncate(time.Second).String(),
					),
					nil,
					func(b bool) {
						if b {
							// Stop the running task
							log.Debug().Msgf("stopping task %s", runningTimesheet.Task.Synopsis)
							err := models.Task(new(models.TaskData)).StopRunningTask()
							if err != nil {
								log.Err(err).Msg(errors.StopRunningTaskError)
							}
						}
					},
				).Show()
			case constants.TimesheetStatusError:
				gui.NewErrorDialogWindow(
					app,
					"timetracker Error",
					appstate.GetStatusError(),
					nil,
					nil,
				).Show()
			case constants.TimesheetStatusIdle:
				// Show the main GUI window
				gui.ShowTimetrackerWindow(nil)
			}
		case <-mAbout.ClickedCh:
			log.Debug().Msg("about menu item selected")
			gui.NewErrorDialogWindow(
				app,
				"About Box",
				fmt.Errorf("this is actually an about box"),
				nil,
				nil,
			).Show()
		case <-mQuit.ClickedCh:
			log.Debug().Msg("quit menu item selected; quitting app")
			statusLoopQuitChan <- true
			app.Quit()
			return
		case <-quitChan:
			log.Debug().Msg("quit channel fired; exiting function")
			statusLoopQuitChan <- true
			return
		}
	}
}

func statusLoop(quitChan chan bool) {
	var err error
	var timesheets []models.TimesheetData

	log := logger.GetLogger("tray.statusLoop")
	log.Trace().Msg("starting")
	appstate.SetStatusError(nil)
	appstate.SetLastState(constants.TimesheetStatusIdle) // Start with idle state
	appstate.SetRunningTimesheet(nil)                    // Start with no running timesheet
	tsd := new(models.TimesheetData)
	for {
		// log.Debug().Msg("getting running timesheet")
		timesheets, err = models.Timesheet(tsd).SearchOpen()
		appstate.SetStatusError(err)
		if err != nil {
			// log.Trace().Msg("set nil running timesheet")
			appstate.SetRunningTimesheet(nil) // Reset running timesheet
			log.Err(err).Msg("error getting running timesheet")
			if appstate.GetLastState() != constants.TimesheetStatusError {
				// Show error icon
				systray.SetTemplateIcon(icons.Error, icons.Error)
				// Update status menu to show `[error]`
				mStatus.SetTitle("[error]")
				appstate.SetLastState(constants.TimesheetStatusError)
			}
		} else {
			if len(timesheets) == 0 {
				// No running task
				// log.Debug().Msg("no running task")
				if appstate.GetLastState() != constants.TimesheetStatusIdle {
					appstate.SetRunningTimesheet(nil) // Reset running timesheet
					// Show check icon
					systray.SetTemplateIcon(icons.Check, icons.Check)
					// Update status menu to show `(idle)`
					mStatus.SetTitle("(idle)")
					appstate.SetLastState(constants.TimesheetStatusIdle)
				}
			} else {
				// Running task...
				// log.Debug().Msgf("running task: %#v", timesheets[0])
				appstate.SetRunningTimesheet(&timesheets[0])
				// Update status menu item to show task ID and duration
				statusText := fmt.Sprintf(
					"%s %s",
					timesheets[0].Task.Synopsis,
					time.Since(timesheets[0].StartTime).Truncate(time.Second).String(),
				)
				mStatus.SetTitle(statusText)
				if appstate.GetLastState() != constants.TimesheetStatusRunning {
					// Show running icon
					systray.SetTemplateIcon(icons.Running, icons.Running)
					appstate.SetLastState(constants.TimesheetStatusRunning)
				}
			}
		}
		// Delay
		// log.Debug().Msgf("delaying %d seconds until next timesheet check", statusLoopDelaySeconds)
		select {
		case <-quitChan:
			log.Debug().Msg("quit channel fired; exiting function")
			return
		case <-time.After(statusLoopDelaySeconds * time.Second):
			break
		}
	}
}
