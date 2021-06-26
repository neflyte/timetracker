package tray

import (
	"errors"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/constants"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/spf13/viper"
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

	recentlyStartedTasks = 5
)

var (
	mStatus                    *systray.MenuItem
	mManage                    *systray.MenuItem
	mCreateAndStart            *systray.MenuItem
	mTrayOptions               *systray.MenuItem
	mTrayOptionConfirmStopTask *systray.MenuItem
	mLastStarted               *systray.MenuItem
	lastStartedItems           [recentlyStartedTasks]*systray.MenuItem
	lastStartedItemSynopses    [recentlyStartedTasks]string
	mAbout                     *systray.MenuItem
	mQuit                      *systray.MenuItem
	trayQuitChan               chan bool
	actionLoopQuitChan         chan bool
	trayLogger                 = logger.GetPackageLogger("tray")
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
	log.Debug().Msg("loading config")
	err := readConfig()
	if err != nil {
		log.Panic().Msgf("error reading app config: %s", err.Error())
		return
	}
	setTrayTitle("Timetracker")
	systray.SetTooltip("Timetracker")
	systray.SetIcon(icons.Check)
	mStatus = systray.AddMenuItem(statusStartTaskTitle, statusStartTaskDescription)
	mCreateAndStart = systray.AddMenuItem("Create and Start new task", "Display a dialog to input new task details and then start the task")
	mManage = systray.AddMenuItem("Manage tasks", "Display the Manage Tasks window to add, change, or remove tasks")
	systray.AddSeparator()
	mTrayOptions = systray.AddMenuItem("Tray options", "Set system tray icon options")
	mTrayOptionConfirmStopTask = mTrayOptions.AddSubMenuItemCheckbox(
		"Confirm when stopping a task",
		"Prompt for confirmation when stopping a running task",
		viper.GetBool(keyStopTaskConfirm),
	)
	// List the top 5 last-started tasks as easy-start options
	systray.AddSeparator()
	mLastStarted = systray.AddMenuItem("Recent tasks", "Select a recently started task to start it again")
	for x := 0; x < recentlyStartedTasks; x++ {
		lastStartedItems[x] = mLastStarted.AddSubMenuItem("--", "")
		lastStartedItemSynopses[x] = ""
	}
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
	log := logger.GetFuncLogger(trayLogger, "onExit")
	// Save config
	err := writeConfig()
	if err != nil {
		log.Err(err).Msg("error saving app config")
	}
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
	} else {
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
	// Last 5 started tasks
	lastStartedTasks, err := new(models.TimesheetData).LastStartedTasks(recentlyStartedTasks)
	if err != nil {
		log.Err(err).Msg("error loading recently-started tasks")
	} else {
		log.Debug().Msgf("len(lastStartedTasks)=%d", len(lastStartedTasks))
		// Hide entries we don't need
		if len(lastStartedTasks) < recentlyStartedTasks {
			log.Debug().Msgf("len(lastStartedTasks) < recentlyStartedTasks; %d < %d", len(lastStartedTasks), recentlyStartedTasks)
			for x := recentlyStartedTasks - 1; x > len(lastStartedItems)-1; x-- {
				log.Debug().Msgf("hiding item at index %d; item=%s", x, lastStartedItems[x].String())
				lastStartedItems[x].SetTitle("--")
				lastStartedItems[x].SetTooltip("")
				lastStartedItems[x].Hide()
				lastStartedItemSynopses[x] = ""
			}
		}
		// Fill in the entries we have
		for x := 0; x < len(lastStartedTasks); x++ {
			log.Debug().Msgf("showing item at index %d; synopsis=%s", x, lastStartedTasks[x].Synopsis)
			lastStartedItems[x].Show()
			lastStartedItems[x].SetTitle(lastStartedTasks[x].Synopsis)
			lastStartedItems[x].SetTooltip(fmt.Sprintf("Start task %s", lastStartedTasks[x].Synopsis))
			lastStartedItemSynopses[x] = lastStartedTasks[x].Synopsis
		}
	}
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
		case <-mTrayOptionConfirmStopTask.ClickedCh:
			shouldConfirmStopTask := viper.GetBool(keyStopTaskConfirm)
			newConfirmValue := !shouldConfirmStopTask
			viper.Set(keyStopTaskConfirm, newConfirmValue)
			if newConfirmValue {
				mTrayOptionConfirmStopTask.Check()
			} else {
				mTrayOptionConfirmStopTask.Uncheck()
			}
		// BEGIN Last started tasks
		case <-lastStartedItems[0].ClickedCh:
			handleLastStartedClick(0)
		case <-lastStartedItems[1].ClickedCh:
			handleLastStartedClick(1)
		case <-lastStartedItems[2].ClickedCh:
			handleLastStartedClick(2) //nolint:gomnd
		case <-lastStartedItems[3].ClickedCh:
			handleLastStartedClick(3) //nolint:gomnd
		case <-lastStartedItems[4].ClickedCh:
			handleLastStartedClick(4) //nolint:gomnd
		// END Last started tasks
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
		shouldConfirmStopTask := viper.GetBool(keyStopTaskConfirm)
		if shouldConfirmStopTask {
			launchGUI(guiOptionStopRunningTask)
			return
		}
		stopRunningTask()
	case constants.TimesheetStatusError:
		lastError := appstate.GetLastError()
		if lastError == nil {
			log.Error().Msg("last error was nil but timesheet status is error; THIS IS UNEXPECTED")
			return
		}
		err := beeep.Alert("Timetracker status error", lastError.Error(), "")
		if err != nil {
			log.Err(err).Msg("error sending notification about status error")
		}
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

func handleLastStartedClick(index uint) {
	log := logger.GetFuncLogger(trayLogger, "handleLastStartedClick").
		With().
		Uint("index", index).
		Logger()
	// Request to handle an index beyond what we have; do nothing
	if index >= recentlyStartedTasks {
		return
	}
	// Get the synopsis from the menu item
	taskSyn := lastStartedItemSynopses[index]
	// Ensure the task exists
	task := models.NewTask()
	task.Data().Synopsis = taskSyn
	err := task.Load(false)
	if err != nil {
		log.Err(err).Msgf("error loading task with synopsis %s: %s", taskSyn, err.Error())
		return
	}
	log.Debug().Msgf("loaded task ID %d (%s)", task.Data().ID, task.Data().Synopsis)
	// Stop the current task if any
	stopRunningTask()
	// Start the new task
	err = startTask(task.Data())
	if err != nil {
		log.Err(err).Msgf("error starting task %s: %s", task.Data().Synopsis, err.Error())
	}
}

func stopRunningTask() {
	log := logger.GetFuncLogger(trayLogger, "stopRunningTask")
	// Stop the task
	stoppedTimesheet, err := models.NewTask().StopRunningTask()
	if err != nil {
		stoppedTimesheet = nil
		if !errors.Is(err, tterrors.ErrNoRunningTask{}) {
			log.Err(err).Msg("error stopping the running task")
			err = beeep.Alert(
				"Error Stopping Task",
				fmt.Sprintf("Error stopping the running task: %s", err.Error()),
				"",
			)
			if err != nil {
				log.Err(err).Msg("error sending notification for stop task error")
			}
			return
		}
	}
	// Show notification that the task has stopped
	if stoppedTimesheet != nil {
		notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)
		notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp))
		err = beeep.Notify(notificationTitle, notificationContents, "")
		if err != nil {
			log.Err(err).Msg("error sending notification about stopped task")
		}
		appstate.SetRunningTimesheet(nil)
	}
}

func startTask(taskData *models.TaskData) (err error) {
	log := logger.GetFuncLogger(trayLogger, "startTask")
	if taskData == nil {
		return errors.New("cannot start a nil task")
	}
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *taskData
	timesheet.Data().StartTime = time.Now()
	err = timesheet.Create()
	if err != nil {
		return
	}
	// Show notification that task started
	notificationTitle := fmt.Sprintf("Task %s started", taskData.Synopsis)
	notificationContents := fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp))
	err = beeep.Notify(notificationTitle, notificationContents, "")
	if err != nil {
		log.Err(err).Msg("error sending notification about started task")
	}
	appstate.SetRunningTimesheet(timesheet.Data())
	return
}
