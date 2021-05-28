package tray

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"os"
	"os/exec"
	"time"
)

const (
	guiOptionStopRunningTask  = "--stop-running-task"
	guiOptionShowManageWindow = "--manage"
	guiOptionShowAboutWindow  = "--about"
)

var (
	mStatus *systray.MenuItem
	mManage *systray.MenuItem
	// mLastStarted       *systray.MenuItem
	// lastStartedItems   []*systray.MenuItem
	mAbout             *systray.MenuItem
	mQuit              *systray.MenuItem
	trayQuitChan       chan bool
	actionLoopQuitChan chan bool
	trayLogger         = logger.GetPackageLogger("tray")
)

func Run() {
	log := logger.GetFuncLogger(trayLogger, "Run")
	// Start the ActionLoop
	actionLoopQuitChan = make(chan bool, 1)
	log.Trace().Msg("go appstate.ActionLoop(...)")
	go appstate.ActionLoop(actionLoopQuitChan)
	// Start the systray
	log.Trace().Msg("systray.Run(...)")
	systray.Run(onReady, onExit)
}

func onReady() {
	log := logger.GetFuncLogger(trayLogger, "onReady")
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
	// Start mainLoop
	trayQuitChan = make(chan bool, 1)
	log.Trace().Msg("go mainLoop(...)")
	go mainLoop(trayQuitChan)
	log.Trace().Msg("done")
}

func onExit() {
	// Shut down mainLoop
	trayQuitChan <- true
	// Shut down ActionLoop
	actionLoopQuitChan <- true
}

func updateStatus(tsd *models.TimesheetData) {
	log := logger.GetFuncLogger(trayLogger, "updateStatus")
	if tsd == nil {
		// No running timesheet
		log.Trace().Msg("got nil running timesheet item")
		systray.SetTemplateIcon(icons.Check, icons.Check)
		mStatus.SetTitle("Start new task")
		mStatus.SetTooltip("Display a task selector and start a task")
		return
	}
	log.Trace().Msgf("got running timesheet object: %s", tsd.String())
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
			handleStatusClick()
		case <-mManage.ClickedCh:
			err := launchGUI(guiOptionShowManageWindow)
			if err != nil {
				log.Err(err).Msg("error launching gui to show manage window")
			}
		case <-mAbout.ClickedCh:
			err := launchGUI(guiOptionShowAboutWindow)
			if err != nil {
				log.Err(err).Msg("error launching gui to show about window")
			}
		case <-mQuit.ClickedCh:
			log.Trace().Msg("quit option clicked; calling systray.Quit() and exiting function")
			systray.Quit()
			return
		case <-quitChan:
			log.Trace().Msg("quit channel fired; exiting function")
			return
		}
	}
}

func launchGUI(options ...string) (err error) {
	log := logger.GetFuncLogger(trayLogger, "launchGUI")
	log.Debug().Msgf("options=%#v", options)
	timetrackerExecutable, err := os.Executable()
	if err != nil {
		log.Err(err).Msg("error getting path and name of this program")
		return err
	}
	guiOptions := []string{"gui"}
	guiOptions = append(guiOptions, options...)
	guiCmd := exec.Command(timetrackerExecutable, guiOptions...)
	err = guiCmd.Start()
	if err != nil {
		log.Err(err).Msgf("error launching gui with Cmd %s", guiCmd.String())
		return err
	}
	log.Debug().Msg("gui launched successfully")
	return nil
}

func handleStatusClick() {
	log := logger.GetFuncLogger(trayLogger, "handleStatusClick")
	switch appstate.GetLastState() {
	case constants.TimesheetStatusRunning:
		err := launchGUI(guiOptionStopRunningTask)
		if err != nil {
			log.Err(err).Msg("error launching gui to stop running task")
		}
	case constants.TimesheetStatusError:
		log.Error().Msg("IMPLEMENTATION MISSING")
	case constants.TimesheetStatusIdle:
		err := launchGUI()
		if err != nil {
			log.Err(err).Msg("error launching gui")
		}
	}
}
