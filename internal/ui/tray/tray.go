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
	"runtime"
	"time"
)

const (
	guiOptionStopRunningTask    = "--stop-running-task"
	guiOptionCreateAndStartTask = "--create-and-start"
	guiOptionShowManageWindow   = "--manage"
	guiOptionShowAboutWindow    = "--about"

	statusStartTaskTitle       = "Start new task"
	statusStartTaskDescription = "Display a task selector and start a task"
	statusStopTaskDescription  = "Stop the running task"
)

var (
	mStatus         *systray.MenuItem
	mManage         *systray.MenuItem
	mCreateAndStart *systray.MenuItem
	// mLastStarted       *systray.MenuItem
	// lastStartedItems   []*systray.MenuItem
	mAbout             *systray.MenuItem
	mQuit              *systray.MenuItem
	trayQuitChan       chan bool
	actionLoopQuitChan chan bool
	trayLogger         = logger.GetPackageLogger("tray")
)

// Run starts the systray app
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
	setTrayTitle("Timetracker")
	systray.SetTooltip("Timetracker")
	systray.SetIcon(icons.Check)
	mStatus = systray.AddMenuItem(statusStartTaskTitle, statusStartTaskDescription)
	mCreateAndStart = systray.AddMenuItem("Create and Start new task", "Display a dialog to input new task details and then start the task")
	mManage = systray.AddMenuItem("Manage tasks", "Display the Manage Tasks window to add, change, or remove tasks")
	// TODO: List the last 5 started tasks (deduplicated) as easy-start options
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
			log.Err(err).Msg("error from running timesheet observable")
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
		systray.SetIcon(icons.Check)
		mStatus.SetTitle(statusStartTaskTitle)
		mStatus.SetTooltip(statusStartTaskDescription)
		return
	}
	log.Trace().Msgf("got running timesheet object: %s", tsd.String())
	systray.SetIcon(icons.Running)
	statusText := fmt.Sprintf(
		"Stop task %s (%s)",
		tsd.Task.Synopsis,
		time.Since(tsd.StartTime).Truncate(time.Second).String(),
	)
	mStatus.SetTitle(statusText)
	mStatus.SetTooltip(statusStopTaskDescription)
}

// FIXME: figure out if there is a way to safely reduce complexity in mainLoop below to remove the nolint directive

func mainLoop(quitChan chan bool) { //nolint:cyclop
	log := logger.GetFuncLogger(trayLogger, "mainLoop")
	// Start main loop
	log.Trace().Msg("starting")
	for {
		select {
		case <-mStatus.ClickedCh:
			handleStatusClick()
		case <-mManage.ClickedCh:
			launchGUI(guiOptionShowManageWindow)
		case <-mAbout.ClickedCh:
			launchGUI(guiOptionShowAboutWindow)
		case <-mCreateAndStart.ClickedCh:
			launchGUI(guiOptionCreateAndStartTask)
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

// launchGUI launches this executable again with gui-specific parameters
func launchGUI(options ...string) {
	log := logger.GetFuncLogger(trayLogger, "launchGUI")
	log.Debug().Msgf("options=%#v", options)
	timetrackerExecutable, err := os.Executable()
	if err != nil {
		log.Err(err).Msg("error getting path and name of this program")
		return
	}
	guiOptions := []string{"gui"}
	guiOptions = append(guiOptions, options...)
	guiCmd := exec.Command(timetrackerExecutable, guiOptions...)
	err = guiCmd.Start()
	if err != nil {
		log.Err(err).Msgf("error launching gui with Cmd %s", guiCmd.String())
		return
	}
	log.Debug().Msg("gui launched successfully")
}

func handleStatusClick() {
	log := logger.GetFuncLogger(trayLogger, "handleStatusClick")
	switch appstate.GetLastState() {
	case constants.TimesheetStatusRunning:
		launchGUI(guiOptionStopRunningTask)
	case constants.TimesheetStatusError:
		log.Error().Msg("IMPLEMENTATION MISSING")
	case constants.TimesheetStatusIdle:
		launchGUI()
	}
}

func setTrayTitle(title string) {
	// Only set the title if we're not on macOS so we don't see the app name beside the icon in the menu bar
	if runtime.GOOS != "darwin" {
		systray.SetTitle(title)
	}
}
