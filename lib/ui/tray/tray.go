package tray

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"fyne.io/systray"
	"github.com/neflyte/timetracker/lib/constants"
	tterrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/neflyte/timetracker/lib/ui/icons"
	tttoast "github.com/neflyte/timetracker/lib/ui/toast"
	"github.com/spf13/viper"
)

const (
	guiOptionStopRunningTask    = "-stop-running-task"
	guiOptionCreateAndStartTask = "-create-and-start"
	guiOptionShowManageWindow   = "-manage"
	guiOptionShowAboutWindow    = "-about"

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
	cleanupFunc                func()
	runningTimesheet           *models.TimesheetData
	lastError                  error
	lastState                  int
	toast                      tttoast.Toast
)

// Run starts the systray app
func Run(cleanupFn func()) {
	cleanupFunc = cleanupFn
	systray.Run(onReady, onExit)
}

func onReady() {
	log := logger.GetFuncLogger(trayLogger, "onReady")
	log.Debug().
		Msg("loading config")
	err := readConfig()
	if err != nil {
		log.Panic().
			Err(err).
			Msg("error reading app config")
		return
	}
	toast = tttoast.NewToast()
	setTrayTitle("Timetracker")
	systray.SetTooltip("Timetracker")
	systray.SetIcon(icons.IconV2NotRunning.StaticContent)
	mStatus = systray.AddMenuItem(statusStartTaskTitle, statusStartTaskDescription)
	mCreateAndStart = systray.AddMenuItem("Create and Start new task", "Display a dialog to input new task details and then start the task") // i18n
	mManage = systray.AddMenuItem("Manage tasks", "Display the Manage Tasks window to add, change, or remove tasks")                         // i18n
	// List the top 5 last-started tasks as easy-start options
	systray.AddSeparator()
	mLastStarted = systray.AddMenuItem("Recent tasks", "Select a recently started task to start it again") // i18n
	for x := range lastStartedItems {
		lastStartedItems[x] = mLastStarted.AddSubMenuItem("-", "")
		lastStartedItemSynopses[x] = ""
	}
	systray.AddSeparator()
	mTrayOptions = systray.AddMenuItem("Tray options", "Set system tray icon options") // i18n
	mTrayOptionConfirmStopTask = mTrayOptions.AddSubMenuItemCheckbox(
		"Confirm when stopping a task",                         // i18n
		"Prompt for confirmation when stopping a running task", // i18n
		viper.GetBool(keyStopTaskConfirm),
	)
	systray.AddSeparator()
	mAbout = systray.AddMenuItem("About Timetracker", "About the Timetracker app") // i18n
	mQuit = systray.AddMenuItem("Quit", "Quit the Timetracker tray app")           // i18n
	// Start mainLoop
	trayQuitChan = make(chan bool, 1)
	log.Trace().
		Msg("go mainLoop(...)")
	go mainLoop(trayQuitChan)
	// Start the ActionLoop
	actionLoopQuitChan = make(chan bool, 1)
	actionLoopStartChan := make(chan bool, 1)
	defer func() {
		actionLoopStartChan <- true
	}()
	log.Trace().
		Msg("go actionLoop(...)")
	go actionLoop(actionLoopQuitChan, actionLoopStartChan)
	log.Trace().
		Msg("done")
}

func onExit() {
	log := logger.GetFuncLogger(trayLogger, "onExit")
	// Save config
	err := writeConfig()
	if err != nil {
		log.Err(err).
			Msg("error saving app config")
	}
	// Shut down mainLoop
	trayQuitChan <- true
	// Shut down ActionLoop
	actionLoopQuitChan <- true
	// Run cleanup function if it is defined
	if cleanupFunc != nil {
		cleanupFunc()
	}
}

func updateStatus(tsd *models.TimesheetData) {
	log := logger.GetFuncLogger(trayLogger, "updateStatus")
	// Update the last 5 started tasks regardless of what happens in this function
	defer updateLast5StartedTasks()
	// Check if last status was error and show the error icon
	if lastState == constants.TimesheetStatusError {
		log.Trace().
			Msg("got error for lastState")
		// Get the error
		if lastError != nil {
			log.Trace().
				Err(lastError).
				Msg("got last error")
		}
		systray.SetIcon(icons.IconV2Error.StaticContent)
		mStatus.SetTitle("Error (click for details)")                   // i18n
		mStatus.SetTooltip("An error occurred; click for more details") // i18n
		return
	}
	// Check the supplied timesheet
	if tsd == nil {
		// No running timesheet
		log.Trace().
			Msg("got nil running timesheet item")
		systray.SetIcon(icons.IconV2NotRunning.StaticContent)
		mStatus.SetTitle(statusStartTaskTitle)
		mStatus.SetTooltip(statusStartTaskDescription)
	} else {
		log.Trace().
			Str("object", tsd.String()).
			Msg("got running timesheet")
		systray.SetIcon(icons.IconV2Running.StaticContent)
		statusText := fmt.Sprintf(
			"Stop task %s (%s)", // i18n
			tsd.Task.Synopsis,
			time.Since(tsd.StartTime).Truncate(time.Second).String(),
		)
		mStatus.SetTitle(statusText)
		mStatus.SetTooltip(statusStopTaskDescription)
	}
}

func updateLast5StartedTasks() {
	log := logger.GetFuncLogger(trayLogger, "updateLast5StartedTasks")
	lastStartedTasks, err := models.NewTimesheet().LastStartedTasks(recentlyStartedTasks)
	if err != nil {
		log.Err(err).
			Msg("error loading recently-started tasks")
		return
	}
	log.Debug().
		Int("length", len(lastStartedTasks)).
		Msg("loaded last-started tasks")
	for idx := range lastStartedTasks {
		log.Debug().
			Int("index", idx).
			Str("task", lastStartedTasks[idx].String()).
			Msg("lastStartedTask")
	}
	// Hide entries we don't need
	if len(lastStartedTasks) < recentlyStartedTasks {
		log.Debug().
			Msgf("len(lastStartedTasks) < recentlyStartedTasks; %d < %d", len(lastStartedTasks), recentlyStartedTasks)
		for x := recentlyStartedTasks - 1; x > len(lastStartedItems)-1; x-- {
			log.Debug().
				Int("index", x).
				Str("item", lastStartedItems[x].String()).
				Msg("hiding item")
			lastStartedItems[x].Disable()
			lastStartedItems[x].SetTitle("-")
			lastStartedItems[x].SetTooltip("")
			lastStartedItemSynopses[x] = ""
		}
	}
	// Fill in the entries we have
	for x := range lastStartedTasks {
		log.Debug().
			Int("index", x).
			Str("synopsis", lastStartedTasks[x].Synopsis).
			Msg("showing item")
		lastStartedItems[x].SetTitle(lastStartedTasks[x].Synopsis)
		lastStartedItems[x].SetTooltip(fmt.Sprintf("Start task %s", lastStartedTasks[x].Synopsis)) // i18n
		lastStartedItems[x].Enable()
		lastStartedItemSynopses[x] = lastStartedTasks[x].Synopsis
	}
}

// FIXME: figure out if there is a way to safely reduce complexity in mainLoop below to remove the nolint directive

func mainLoop(quitChan chan bool) { //nolint:cyclop
	log := logger.GetFuncLogger(trayLogger, "mainLoop")
	// Create a channel to catch OS signals
	signalChan := make(chan os.Signal, 1)
	// Catch OS interrupt and SIGTERM signals
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	// Start main loop
	log.Trace().
		Msg("starting")
	for {
		select {
		case <-mStatus.ClickedCh:
			handleStatusClick()
		case <-mManage.ClickedCh:
			launchGUI(guiOptionShowManageWindow)
		case <-mTrayOptionConfirmStopTask.ClickedCh:
			toggleConfirmStopTask()
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
		case <-signalChan:
			log.Trace().
				Msg("caught interrupt or SIGTERM signal; calling systray.Quit() and exiting function")
			systray.Quit()
			return
		case <-mQuit.ClickedCh:
			log.Trace().
				Msg("quit option clicked; calling systray.Quit() and exiting function")
			systray.Quit()
			return
		case <-quitChan:
			log.Trace().
				Msg("quit channel fired; exiting function")
			return
		}
	}
}

// launchGUI launches this executable again with gui-specific parameters
func launchGUI(guiOptions ...string) {
	log := logger.GetFuncLogger(trayLogger, "launchGUI")
	log.Debug().
		Strs("guiOptions", guiOptions).
		Msg("function options")
	timetrackerExecutable, err := os.Executable()
	if err != nil {
		log.Err(err).
			Msg("error getting path and name of this program")
		return
	}
	timetrackerDir := path.Dir(timetrackerExecutable)
	guiExecutable := path.Join(timetrackerDir, "timetracker-gui")
	if runtime.GOOS == "windows" {
		guiExecutable += ".exe"
	}
	log.Debug().
		Str("guiExecutable", guiExecutable).
		Strs("guiOptions", guiOptions).
		Msg("command to launch gui")
	guiCmd := exec.Command(guiExecutable, guiOptions...)
	err = guiCmd.Start()
	if err != nil {
		log.Err(err).
			Str("command", guiCmd.String()).
			Msg("error launching gui")
		return
	}
	log.Debug().Msg("gui launched successfully")
}

func handleStatusClick() {
	log := logger.GetFuncLogger(trayLogger, "handleStatusClick")
	switch lastState {
	case constants.TimesheetStatusRunning:
		shouldConfirmStopTask := viper.GetBool(keyStopTaskConfirm)
		if shouldConfirmStopTask {
			launchGUI(guiOptionStopRunningTask)
			return
		}
		stopRunningTask()
	case constants.TimesheetStatusError:
		if lastError == nil {
			log.Error().
				Msg("last error was nil but timesheet status is error; THIS IS UNEXPECTED")
			return
		}
		err := toast.Notify("Timetracker status error", lastError.Error())
		if err != nil {
			log.Err(err).
				Msg("error sending notification about status error")
		}
	case constants.TimesheetStatusIdle:
		launchGUI()
	}
}

func setTrayTitle(title string) {
	// Only set the title if we're not on macOS, so we don't see the app name beside the icon in the menu bar
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
		log.Err(err).
			Str("synopsis", taskSyn).
			Msg("error loading task")
		return
	}
	log.Debug().
		Uint("id", task.Data().ID).
		Str("synopsis", task.Data().Synopsis).
		Msg("loaded task")
	// Stop the current task if any
	stopRunningTask()
	// Start the new task
	err = startTask(task.Data())
	if err != nil {
		log.Err(err).
			Str("synopsis", task.Data().Synopsis).
			Msg("error starting task")
	}
}

func stopRunningTask() {
	log := logger.GetFuncLogger(trayLogger, "stopRunningTask")
	// Stop the task
	stoppedTimesheet, err := models.NewTask().StopRunningTask()
	if err != nil {
		stoppedTimesheet = nil
		if !errors.Is(err, tterrors.ErrNoRunningTask{}) {
			log.Err(err).
				Msg("error stopping the running task")
			err = toast.Notify(
				"Error Stopping Task", // i18n
				fmt.Sprintf("Error stopping the running task: %s", err.Error()), // i18n
			)
			if err != nil {
				log.Err(err).
					Msg("error sending notification for stop task error")
			}
			return
		}
	}
	// Show notification that the task has stopped
	if stoppedTimesheet != nil {
		runningTimesheet = nil
		notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)                     // i18n
		notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)) // i18n
		err = toast.Notify(notificationTitle, notificationContents)
		if err != nil {
			log.Err(err).
				Msg("error sending notification about stopped task")
		}
	}
	updateStatus(runningTimesheet)
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
		log.Err(err).
			Msg("error creating new timesheet to start a task")
		return
	}
	runningTimesheet = timesheet.Data()
	updateStatus(runningTimesheet)
	// Show notification that task started
	notificationTitle := fmt.Sprintf("Task %s started", taskData.Synopsis)                              // i18n
	notificationContents := fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp)) // i18n
	err = toast.Notify(notificationTitle, notificationContents)
	if err != nil {
		log.Err(err).
			Msg("error sending notification about started task")
	}
	return
}

func toggleConfirmStopTask() {
	shouldConfirmStopTask := viper.GetBool(keyStopTaskConfirm)
	newConfirmValue := !shouldConfirmStopTask
	viper.Set(keyStopTaskConfirm, newConfirmValue)
	if newConfirmValue {
		mTrayOptionConfirmStopTask.Check()
	} else {
		mTrayOptionConfirmStopTask.Uncheck()
	}
}
