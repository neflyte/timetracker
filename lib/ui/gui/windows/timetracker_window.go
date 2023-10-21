package windows

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/neflyte/timetracker/lib/constants"
	tterrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	ttmonitor "github.com/neflyte/timetracker/lib/monitor"
	"github.com/neflyte/timetracker/lib/ui/gui/dialogs"
	"github.com/neflyte/timetracker/lib/ui/gui/widgets"
	"github.com/neflyte/timetracker/lib/ui/icons"
	tttoast "github.com/neflyte/timetracker/lib/ui/toast"
	"github.com/neflyte/timetracker/lib/utils"
	"github.com/rs/zerolog"
)

const (
	// recentlyStartedTasks is the number of recent tasks to display
	recentlyStartedTasks = uint(5)
)

// TimetrackerWindow is the main timetracker GUI window interface
type TimetrackerWindow interface {
	windowBase
	ShowAbout()
	ShowWithError(err error)
	ShowWithManageWindow()
	ShowAndStopRunningTask()
	ShowAndDisplayCreateAndStartDialog()
}

// timetrackerWindowData is the struct underlying the TimetrackerWindow interface
type timetrackerWindowData struct {
	log                         zerolog.Logger
	rptWindow                   reportWindow
	createNewTaskAndStartDialog dialogs.CreateAndStartTaskDialog
	fyne.Window
	selectedTask        models.Task
	mngWindowV2         manageWindowV2
	toast               tttoast.Toast
	monitor             ttmonitor.Service
	monitorQuitChan     chan bool
	runningTimesheet    *models.TimesheetData
	container           *fyne.Container
	elapsedTimeQuitChan chan bool
	elapsedTimeTicker   *time.Ticker
	taskSelector        *widgets.TaskSelector
	app                 *fyne.App
	compactUI           *widgets.CompactUI
	appVersion          string
	selectedTaskMtx     sync.RWMutex
	elapsedTimeRunning  bool
}

// NewTimetrackerWindow creates and initializes a new timetracker window
func NewTimetrackerWindow(app fyne.App, appVersion string) TimetrackerWindow {
	ttw := &timetrackerWindowData{
		app:                 &app,
		appVersion:          appVersion,
		Window:              app.NewWindow("Timetracker"),
		log:                 logger.GetStructLogger("timetrackerWindowData"),
		elapsedTimeRunning:  false,
		elapsedTimeQuitChan: make(chan bool, 1),
		monitorQuitChan:     make(chan bool, 1),
		toast:               tttoast.NewToast(),
		selectedTaskMtx:     sync.RWMutex{},
	}
	err := ttw.Init()
	if err != nil {
		ttw.log.
			Err(err).
			Msg("error initializing window")
	}
	return ttw
}

// Init initializes the window
func (t *timetrackerWindowData) Init() error {
	// Initialize the UI
	err := t.initUI()
	if err != nil {
		return err
	}
	// Initialize monitor service
	t.monitor = ttmonitor.NewService(t.monitorQuitChan)
	// Initialize observables
	t.initObservables()
	// Initialize window display data
	err = t.initWindowData()
	if err != nil {
		return err
	}
	return nil
}

// initUI initializes the UI part of the window
func (t *timetrackerWindowData) initUI() error {
	if t.app == nil {
		return errors.New("t.app was nil; this is unexpected")
	}
	t.compactUI = widgets.NewCompactUI()
	t.createNewTaskAndStartDialog = dialogs.NewCreateAndStartTaskDialog((*t.app).Preferences(), t.createAndStartTaskDialogCallback, t.Window)
	t.taskSelector = widgets.NewTaskSelector()
	t.container = container.NewPadded(t.compactUI)
	t.Window.SetContent(t.container)
	t.Window.SetIcon(icons.IconV2)
	// get the size of the content with everything visible
	siz := t.Window.Content().Size()
	// Resize the window to the minimum size^H^H^H^H^H height
	newWindowSize := fyne.NewSize(siz.Width, t.Window.Canvas().Size().Height)
	if newWindowSize.Height < minimumWindowHeight {
		newWindowSize.Height = minimumWindowHeight
	}
	if newWindowSize.Width < minimumWindowWidth {
		newWindowSize.Width = minimumWindowWidth
	}
	t.Window.Resize(newWindowSize)
	// Resize the container to the original size before we resized the window
	t.container.Resize(siz)
	// intercept the window close event
	t.Window.SetCloseIntercept(t.Close)
	// Set up the manage window and hide it
	t.mngWindowV2 = newManageWindowV2(*t.app)
	t.mngWindowV2.Hide()
	// Also set up the report window and hide it
	t.rptWindow = newReportWindow(*t.app)
	t.rptWindow.Hide()
	return nil
}

func (t *timetrackerWindowData) initObservables() {
	t.compactUI.Observable().ForEach(
		t.handleCompactUIEvent,
		utils.ObservableErrorHandler("compactUI", t.log),
		utils.ObservableCloseHandler("compactUI", t.log),
	)
	t.taskSelector.Observable().ForEach(
		t.handleTaskSelectorEvent,
		utils.ObservableErrorHandler("taskSelector", t.log),
		utils.ObservableCloseHandler("taskSelector", t.log),
	)
	t.monitor.Observable().ForEach(
		t.handleMonitorServiceEvent,
		utils.ObservableErrorHandler("monitor", t.log),
		utils.ObservableCloseHandler("monitor", t.log),
	)
}

// initWindowData primes the window with some data
func (t *timetrackerWindowData) initWindowData() error {
	log := logger.GetFuncLogger(t.log, "initWindowData")
	// Load the last running task list
	t.refreshTaskList()
	// Load the running task, if any
	runningTimesheet, err := models.NewTimesheet().RunningTimesheet()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("unable to get running timesheet")
		return err
	}
	var runningTS *models.TimesheetData
	if runningTimesheet != nil {
		runningTS = runningTimesheet.Data()
	}
	t.setRunningTimesheet(runningTS)
	return nil
}

func (t *timetrackerWindowData) setSelectedTask(selectedTask models.Task) {
	t.selectedTaskMtx.Lock()
	t.selectedTask = selectedTask
	t.selectedTaskMtx.Unlock()
}

// refreshTaskList loads the last started tasks and sends them to the CompactUI
func (t *timetrackerWindowData) refreshTaskList() {
	log := logger.GetFuncLogger(t.log, "refreshTaskList")
	recentTasks, err := models.NewTimesheet().LastStartedTasks(recentlyStartedTasks)
	if err != nil {
		log.Err(err).
			Uint("recentlyStartedTasks", recentlyStartedTasks).
			Msg("unable to load last started tasks")
		return
	}
	log.Debug().
		Int("numTasks", len(recentTasks)).
		Msg("loaded last started tasks")
	taskList := make(models.TaskList, len(recentTasks))
	for idx := range recentTasks {
		taskList[idx] = models.NewTaskWithData(recentTasks[idx])
		log.Debug().
			Int("idx", idx).
			Str("task", taskList[idx].DisplayString()).
			Msg("add to task list")
	}
	t.compactUI.SetTaskList(taskList)
}

// setRunningTimesheet updates CompactUI and handles the elapsedTimeLoop
func (t *timetrackerWindowData) setRunningTimesheet(tsd *models.TimesheetData) {
	log := logger.GetFuncLogger(t.log, "setRunningTimesheet")
	t.runningTimesheet = tsd
	running := false
	taskName := "idle"
	elapsedTime := ""
	if tsd != nil {
		running = true
		taskName = tsd.Task.Synopsis
		elapsedTime = t.elapsedTime(tsd.StartTime)
	}
	log.Debug().
		Bool("running", running).
		Str("taskName", taskName).
		Str("elapsedTime", elapsedTime).
		Msg("update CompactUI")
	t.compactUI.SetTaskName(taskName)
	t.compactUI.SetElapsedTime(elapsedTime)
	t.compactUI.SetRunning(running)
	// Handle the elapsedTimeLoop
	switch running {
	case true:
		// A task is running but the loop is not. Start the loop.
		if !t.elapsedTimeRunning {
			go t.elapsedTimeLoop(tsd.StartTime, t.elapsedTimeQuitChan)
		}
	case false:
		// A task is not running but the loop is. Stop the loop.
		if t.elapsedTimeRunning {
			t.elapsedTimeQuitChan <- true
		}
	}
}

// doCreateAndStartTask shows the Create and Start Task dialog box
func (t *timetrackerWindowData) doCreateAndStartTask() {
	log := logger.GetFuncLogger(t.log, "doCreateAndStartTask")
	log.Debug().
		Msg("hide closeWindow checkbox")
	t.createNewTaskAndStartDialog.HideCloseWindowCheckbox()
	log.Debug().
		Msg("show dialog")
	t.createNewTaskAndStartDialog.Show()
}

func (t *timetrackerWindowData) doStartSelectedTask() {
	log := logger.GetFuncLogger(t.log, "doStartSelectedTask")
	log.Debug().
		Msg("check for running timesheet")
	runningTS, err := models.NewTimesheet().RunningTimesheet()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("unable to get the running timesheet")
		return
	}
	if runningTS != nil {
		log.Debug().
			Msg("a timesheet is running; ask the user if it should stop")
		stopTaskDialog := dialogs.NewStopTaskDialog(
			runningTS.Data().Task,
			(*t.app).Preferences(),
			t.handleStopTaskDialogResult,
			t.Window,
		)
		stopTaskDialog.SetCloseWindowCheckbox(true)
		stopTaskDialog.Show()
		return
	}
	t.doStartTask()
}

// doStartTask starts the task in t.selectedTask if a task isn't already running
func (t *timetrackerWindowData) doStartTask() {
	log := logger.GetFuncLogger(t.log, "doStartTask")
	if t.isTimesheetOpen() {
		log.Warn().
			Msg("a timesheet is already open; will not start a task")
		return
	}
	// Lock t.selectedTask for read
	t.selectedTaskMtx.RLock()
	if t.selectedTask == nil {
		log.Error().
			Msg("no task was selected")
		dialog.NewError(
			fmt.Errorf("please select a task to start"), // i18n
			t.Window,
		).Show()
		// Release the lock before returning
		t.selectedTaskMtx.RUnlock()
		return
	}
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *t.selectedTask.Data()
	// Release the lock since we are done using t.selectedTask
	t.selectedTaskMtx.RUnlock()
	timesheet.Data().StartTime = time.Now()
	err := timesheet.Create()
	if err != nil {
		log.Err(err).
			Msg("error creating new timesheet")
		dialog.NewError(err, t.Window).Show()
		return
	}
	// Show notification that task started
	t.notify(
		fmt.Sprintf("Task %s started", timesheet.Data().Task.Synopsis),              // i18n
		fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp)), // i18n
	)
	t.setRunningTimesheet(timesheet.Data())
	// Tell the monitor that we've got a running task
	t.monitor.SetRunningTimesheet(timesheet)
	// Refresh task list
	t.refreshTaskList()
}

// doStopTask attempts to stop the running task, if there is any.
// If there is no task running, the function exits without error.
func (t *timetrackerWindowData) doStopTask() {
	log := logger.GetFuncLogger(t.log, "doStopTask")
	// Stop the running task
	log.Debug().
		Msg("stopping running task")
	stoppedTimesheet, err := models.NewTask().StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg(tterrors.StopRunningTaskError)
		dialog.NewError(err, t.Window).Show()
	}
	if stoppedTimesheet == nil {
		// There was no task running; nothing more to do
		log.Debug().
			Msg("no tasks were running; nothing to do")
		return
	}
	// Show notification that task has stopped
	t.notify(
		fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis),                  // i18n
		fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)), // i18n
	)
	t.setRunningTimesheet(nil)
	t.monitor.SetRunningTimesheet(nil)
}

func (t *timetrackerWindowData) doStopAndStartTask() {
	t.doStopTask()
	t.doStartTask()
}

func (t *timetrackerWindowData) doManageTasksV2() {
	t.mngWindowV2.Show()
}

func (t *timetrackerWindowData) doSelectTask() {
	t.taskSelector.Reset()
	t.taskSelector.FilterTasks()
	selectTaskDialog := dialog.NewCustomConfirm(
		"Select a task", // i18n
		"SELECT",        // i18n
		"CANCEL",        // i18n
		t.taskSelector,
		t.handleSelectTaskResult,
		t.Window,
	)
	// Resize the dialog so it is wider than normal
	dialogs.ResizeDialogToWindowWithPadding(selectTaskDialog, t.Window, dialogSizeOffset)
	// Show the dialog
	selectTaskDialog.Show()
}

func (t *timetrackerWindowData) handleSelectTaskResult(selected bool) {
	log := logger.GetFuncLogger(t.log, "handleSelectTaskResult")
	if !selected {
		return
	}
	selectedTask := t.taskSelector.Selected()
	if selectedTask == nil {
		log.Error().
			Msg("selected task is nil; this is unexpected")
		return
	}
	t.setSelectedTask(selectedTask)
	t.doStartSelectedTask()
}

// handleTaskSelectorEvent logs the events coming from the TaskSelector. No further action
// beyond logging is necessary.
func (t *timetrackerWindowData) handleTaskSelectorEvent(item interface{}) {
	log := logger.GetFuncLogger(t.log, "handleTaskSelectorEvent")
	switch event := item.(type) {
	case widgets.TaskSelectorSelectedEvent:
		if event.SelectedTask == nil {
			log.Warn().
				Msg("nil task in SelectedEvent")
			break
		}
		log.Debug().
			Str("selected", event.SelectedTask.String()).
			Msg("user selected a task from the taskSelector")
	case widgets.TaskSelectorErrorEvent:
		if event.Err != nil {
			log.Err(event.Err).
				Msg("error from task selector")
		}
	}
}

func (t *timetrackerWindowData) handleCompactUIEvent(item interface{}) {
	log := logger.GetFuncLogger(t.log, "handleCompactUIEvent")
	switch event := item.(type) {
	case widgets.CompactUISelectTaskEvent:
		t.doSelectTask()
	case widgets.CompactUICreateAndStartEvent:
		t.doCreateAndStartTask()
	case widgets.CompactUIManageEvent:
		t.doManageTasksV2()
	case widgets.CompactUIReportEvent:
		t.doReport()
	case widgets.CompactUIAboutEvent:
		t.doAbout()
	case widgets.CompactUIQuitEvent:
		t.Close()
	case widgets.CompactUITaskEvent:
		if event.ShouldStopTask() {
			t.doStopTask()
			return
		}
		t.selectedTaskMtx.RLock()
		if t.selectedTask != nil && t.selectedTask.Equals(event.Task) && t.compactUI.IsRunning() {
			log.Debug().
				Msg("selected task is the same as the event's task, and it is running; doing nothing")
			t.selectedTaskMtx.RUnlock()
			return
		}
		t.selectedTaskMtx.RUnlock()
		t.setSelectedTask(event.Task)
		t.doStartSelectedTask()
	}
}

func (t *timetrackerWindowData) handleMonitorServiceEvent(item interface{}) {
	log := logger.GetFuncLogger(t.log, "handleMonitorServiceEvent")
	if _, ok := item.(ttmonitor.ServiceUpdateEvent); ok {
		switch t.monitor.TimesheetStatus() {
		case constants.TimesheetStatusError:
			tsErr := t.monitor.TimesheetError()
			if tsErr != nil {
				log.Err(tsErr).
					Msg("error from TimesheetStatus")
			}
			// TODO: should we do anything else here?
		case constants.TimesheetStatusIdle:
			// if we're running, stop.
			if t.runningTimesheet != nil {
				t.setRunningTimesheet(nil)
			}
		case constants.TimesheetStatusRunning:
			runningTS := t.monitor.RunningTimesheet()
			if runningTS != nil {
				// if we're not running or if we're running something different, update.
				if t.runningTimesheet == nil || !t.runningTimesheet.Equals(runningTS) {
					t.setRunningTimesheet(runningTS.Data())
				}
			}
		}
	}
}

func (t *timetrackerWindowData) handleStopTaskDialogResult(shouldStop bool) {
	if shouldStop {
		t.doStopAndStartTask()
	}
}

func (t *timetrackerWindowData) doReport() {
	t.rptWindow.Show()
}

func (t *timetrackerWindowData) doAbout() {
	dialog.NewInformation(
		"About Timetracker", // i18n
		fmt.Sprintf("Timetracker %s\n\nhttps://github.com/neflyte/timetracker", t.appVersion),
		t.Window,
	).Show()
}

func (t *timetrackerWindowData) maybeStopRunningTask(stopTask bool) {
	// log := logger.GetFuncLogger(t.log, "maybeStopRunningTask")
	if !stopTask {
		return
	}
	// Stop the task
	t.doStopTask()
	// Check if we should close the main window
	shouldCloseMainWindow := (*t.app).Preferences().BoolWithFallback(constants.PrefKeyCloseWindowStopTask, false)
	if shouldCloseMainWindow {
		t.Close()
	}
}

func (t *timetrackerWindowData) createAndStartTaskDialogCallback(createAndStart bool) {
	log := logger.GetFuncLogger(t.log, "createAndStartTaskDialogCallback").
		With().Bool("createAndStart", createAndStart).
		Logger()
	if !createAndStart {
		log.Debug().
			Msg("will not create/start the task")
		return
	}
	taskData := t.createNewTaskAndStartDialog.GetTask()
	if taskData == nil {
		log.Error().
			Msg("taskData was nil; this is unexpected")
		return
	}
	log.Debug().
		Str("newTask", taskData.DisplayString()).
		Msg("got taskData from dialog")
	// check for an existing task
	existingTasks, err := models.NewTask().SearchBySynopsis(taskData.Synopsis)
	if err != nil {
		// error checking for existing task
		log.Err(err).
			Str("synopsis", taskData.Synopsis).
			Msg("error checking for existing tasks")
		// display an error dialog
		errorDialog := dialog.NewError(
			fmt.Errorf("could not check for existing tasks with synopsis '%s': %w", taskData.Synopsis, err),
			t.Window,
		)
		// re-display the dialog after the error is dismissed
		errorDialog.SetOnClosed(func() {
			t.createNewTaskAndStartDialog.Show()
		})
		errorDialog.Show()
		return
	}
	if len(existingTasks) > 0 {
		// existing task!
		log.Error().
			Str("synopsis", taskData.Synopsis).
			Msg("there are existing tasks with the desired synopsis; please choose another synopsis")
		// display an error
		errorDialog := dialog.NewError(
			fmt.Errorf("there are existing tasks with synopsis '%s'\nplease choose another synopsis", taskData.Synopsis), // i18n
			t.Window,
		)
		// re-display the dialog after the error is dismissed
		errorDialog.SetOnClosed(func() {
			t.createNewTaskAndStartDialog.Show()
		})
		errorDialog.Show()
		return
	}
	// Create the new task
	err = taskData.Create()
	if err != nil {
		log.Err(err).
			Str("newTask", taskData.String()).
			Msg("error creating new task")
		return
	}
	log.Debug().
		Str("newTask", taskData.String()).
		Msg("created new task")
	// reset the create dialog now that the task has been created
	t.createNewTaskAndStartDialog.Reset()
	// Set the new task as the "selected task"
	t.setSelectedTask(models.NewTaskWithData(*taskData))
	t.selectedTaskMtx.RLock()
	t.compactUI.SetSelectedTask(t.selectedTask)
	t.selectedTaskMtx.RUnlock()
	// Start the new task
	t.doStartSelectedTask()
}

// elapsedTimeLoop is a loop that draws the elapsed time since the running task was started
func (t *timetrackerWindowData) elapsedTimeLoop(startTime time.Time, quitChan chan bool) {
	log := logger.GetFuncLogger(t.log, "elapsedTimeLoop")
	if t.elapsedTimeRunning {
		return
	}
	t.elapsedTimeRunning = true
	defer func() {
		t.elapsedTimeRunning = false
	}()
	t.elapsedTimeTicker = time.NewTicker(time.Second)
	defer t.elapsedTimeTicker.Stop()
	// Clear the elapsed time display when the loop ends
	defer t.compactUI.SetElapsedTime("")
	log.Debug().
		Msg("loop starting")
	defer log.Debug().Msg("loop ending")
	for {
		select {
		case <-t.elapsedTimeTicker.C:
			t.compactUI.SetElapsedTime(t.elapsedTime(startTime))
		case <-quitChan:
			return
		}
	}
}

// notify sends a notification using the toast object
func (t *timetrackerWindowData) notify(title string, contents string) {
	err := t.toast.Notify(title, contents)
	if err != nil {
		log := logger.GetFuncLogger(t.log, "notify")
		log.Err(err).
			Str("title", title).
			Str("contents", contents).
			Msg("unable to send notification")
	}
}

func (t *timetrackerWindowData) elapsedTime(since time.Time) string {
	return time.Since(since).Truncate(time.Second).String()
}

func (t *timetrackerWindowData) isTimesheetOpen() bool {
	log := logger.GetFuncLogger(t.log, "isTimesheetOpen")
	openCount, err := models.NewTimesheet().CountOpen()
	if err != nil {
		log.Err(err).
			Msg("unable to get count of open timesheets; returning false")
		return false
	}
	return openCount > 0
}

// Show shows the main window
func (t *timetrackerWindowData) Show() {
	// Start the monitor service if it is not running
	if !t.monitor.IsRunning() {
		t.monitor.Start(nil)
	}
	// Show the window
	t.Window.Show()
}

// ShowAndStopRunningTask shows the main window and asks the user if they want to stop the running task
func (t *timetrackerWindowData) ShowAndStopRunningTask() {
	runningTS, err := models.NewTimesheet().RunningTimesheet()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		t.ShowWithError(err) // i18n
		return
	}
	if runningTS == nil {
		t.ShowWithError(fmt.Errorf("a task is not running; please start a task first")) // i18n
		return
	}
	t.Show()
	dialogs.NewStopTaskDialog(
		runningTS.Data().Task,
		(*t.app).Preferences(),
		t.maybeStopRunningTask,
		t.Window,
	).Show()
}

// ShowWithManageWindow shows the main window followed by the Manage window
func (t *timetrackerWindowData) ShowWithManageWindow() {
	t.Show()
	t.doManageTasksV2()
}

// ShowWithError shows the main window and then shows an error dialog
func (t *timetrackerWindowData) ShowWithError(err error) {
	t.Show()
	dialog.NewError(err, t.Window).Show()
}

// ShowAbout shows the About dialog box
func (t *timetrackerWindowData) ShowAbout() {
	t.Show()
	t.doAbout()
}

// ShowAndDisplayCreateAndStartDialog shows the main window and then shows the Create and Start New Task dialog
func (t *timetrackerWindowData) ShowAndDisplayCreateAndStartDialog() {
	t.Show()
	t.createNewTaskAndStartDialog.Show()
}

// Hide hides the main window and the manage window
func (t *timetrackerWindowData) Hide() {
	if t.mngWindowV2 != nil {
		t.mngWindowV2.Hide()
	}
	t.Window.Hide()
}

// Close closes the main window
func (t *timetrackerWindowData) Close() {
	// Clean up the notification object
	t.toast.Cleanup()
	// Check if elapsed time counter is running and stop it if it is
	if t.elapsedTimeRunning {
		t.elapsedTimeQuitChan <- true
	}
	// Check if monitor service is running and stop it if it is
	if t.monitor.IsRunning() {
		t.monitor.Stop()
	}
	// Close the window
	t.Window.Close()
	// Quit
	(*t.app).Quit()
}
