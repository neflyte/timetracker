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
	mStatus *systray.MenuItem
	mManage *systray.MenuItem
	// mLastStarted       *systray.MenuItem
	// lastStartedItems   []*systray.MenuItem
	mAbout             *systray.MenuItem
	mQuit              *systray.MenuItem
	lockFile           lockfile.Lockfile
	pidPath            string
	wg                 sync.WaitGroup
	trayQuitChan       chan bool
	actionLoopQuitChan chan bool
	trayLogger         = logger.GetPackageLogger("tray")
)

func Run() (err error) {
	log := logger.GetFuncLogger(trayLogger, "Run")
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

	log := logger.GetFuncLogger(trayLogger, "onReady")
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
	mStatus = systray.AddMenuItem("Start new task", "Display a task selector and start a task")
	mManage = systray.AddMenuItem("Manage tasks", "Display the Manage Tasks window to add, change, or remove tasks")
	// TODO: List the top 5 last-started tasks as easy-start options
	// systray.AddSeparator()
	// mLastStarted = systray.AddMenuItem("Recent tasks", "Select a recently started task to start it again")
	// lastStartedItems = make([]*systray.MenuItem, 0)
	systray.AddSeparator()
	mAbout = systray.AddMenuItem("About Timetracker", "About the Timetracker app")
	mQuit = systray.AddMenuItem("Quit", "Quit the Timetracker tray app")
	log.Trace().Msg("setting up observables")
	appstate.Observables()[appstate.KeyRunningTimesheet].ForEach(
		func(item interface{}) {
			tsd, ok := item.(*models.TimesheetData)
			if ok {
				updateStatus(tsd)
			}
		},
		func(err error) {
			systray.SetTemplateIcon(icons.Error, icons.Error)
			mStatus.SetTitle("An error occurred; click for details")
		},
		func() {
			log.Debug().Msg("running timesheet observable is done")
		},
	)
	log.Trace().Msg("priming status")
	updateStatus(appstate.GetRunningTimesheet())
	log.Trace().Msg("done")
}

func onExit() {
	log := logger.GetFuncLogger(trayLogger, "onExit")
	err := lockFile.Unlock()
	if err != nil {
		log.Err(err).Msgf("error releasing pidfile")
		return
	}
	log.Debug().Msg("unlocked pidfile")
}

func updateStatus(tsd *models.TimesheetData) {
	log := logger.GetFuncLogger(trayLogger, "updateStatus")
	log.Debug().Msg("called")
	if tsd == nil {
		// No running timesheet
		log.Debug().Msg("got nil running timesheet item")
		systray.SetTemplateIcon(icons.Check, icons.Check)
		mStatus.SetTitle("Start new task")
		mStatus.SetTooltip("Display a task selector and start a task")
		return
	}
	log.Debug().Msgf("got running timesheet object: %s", tsd.String())
	systray.SetTemplateIcon(icons.Running, icons.Running)
	statusText := fmt.Sprintf(
		"Stop task %s (%s)",
		tsd.Task.Synopsis,
		time.Since(tsd.StartTime).Truncate(time.Second).String(),
	)
	mStatus.SetTitle(statusText)
	mStatus.SetTooltip("Stop the running task")
}

func mainLoop(quitChan chan bool) {
	log := logger.GetFuncLogger(trayLogger, "mainLoop")
	// Start main loop
	log.Trace().Msg("starting")
	for {
		select {
		case <-mStatus.ClickedCh:
			switch appstate.GetLastState() {
			case constants.TimesheetStatusRunning:
				runningTimesheet := appstate.GetRunningTimesheet()
				taskMessage := fmt.Sprintf(
					"Stop task %s (%s) ?",
					runningTimesheet.Task.Synopsis,
					time.Since(runningTimesheet.StartTime).Truncate(time.Second).String(),
				)
				gui.ShowTimetrackerWindowWithConfirm(
					"Stop running task?",
					taskMessage,
					stopTaskConfirmCallback,
					false,
				)
			case constants.TimesheetStatusError:
				gui.ShowTimetrackerWindowWithError(appstate.GetStatusError())
			case constants.TimesheetStatusIdle:
				gui.ShowTimetrackerWindow()
			}
		case <-mManage.ClickedCh:
			gui.ShowTimetrackerWindowWithManageWindow()
		case <-mAbout.ClickedCh:
			gui.ShowTimetrackerWindowWithAbout()
		case <-mQuit.ClickedCh:
			gui.StopGUI()
			return
		case <-quitChan:
			log.Trace().Msg("quit channel fired; exiting function")
			return
		}
	}
}

func stopTaskConfirmCallback(res bool) {
	log := logger.GetFuncLogger(trayLogger, "stopTaskConfirmCallback")
	if res {
		// Stop the running task
		log.Debug().Msg("stopping the running task")
		err := models.Task(new(models.TaskData)).StopRunningTask()
		if err != nil {
			log.Err(err).Msg(errors.StopRunningTaskError)
		}
		// Get a new timesheet and update the appstate
		appstate.UpdateRunningTimesheet()
	}
}
