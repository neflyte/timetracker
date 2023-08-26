package windows

import (
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/neflyte/timetracker/lib/constants"
	tterrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
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
	compactUI           *widgets.CompactUI
	runningTimesheet    *models.TimesheetData
	container           *fyne.Container
	elapsedTimeQuitChan chan bool
	elapsedTimeTicker   *time.Ticker
	taskSelector        *widgets.TaskSelector
	app                 *fyne.App
	appVersion          string
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
		toast:               tttoast.NewToast(),
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
}

// initWindowData primes the window with some data
func (t *timetrackerWindowData) initWindowData() error {
	log := logger.GetFuncLogger(t.log, "initWindowData")
	log.Trace().Msg("started")
	defer log.Trace().Msg("done")
	err := t.refreshTaskList()
	if err != nil {
		log.Err(err).
			Msg("unable to refresh task list")
		return err
	}
	// Load the running task
	runningTimesheet, err := models.NewTimesheet().RunningTimesheet()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("unable to get running timesheets")
		return err
	}
	if runningTimesheet == nil {
		// Task is not running
		t.setRunningTimesheet(nil)
		t.selectedTask = nil
		return nil
	}
	// Task is running
	runningTS := runningTimesheet.Data()
	t.setRunningTimesheet(runningTS)
	t.selectedTask = models.NewTaskWithData(runningTS.Task)
	return nil
}

func (t *timetrackerWindowData) refreshTaskList() error {
	log := logger.GetFuncLogger(t.log, "refreshTaskList")
	recentTasks, err := models.NewTimesheet().LastStartedTasks(recentlyStartedTasks)
	if err != nil {
		log.Err(err).
			Uint("recentlyStartedTasks", recentlyStartedTasks).
			Msg("unable to load last started tasks")
		return err
	}
	taskList := make([]string, len(recentTasks))
	for idx := range recentTasks {
		taskList[idx] = recentTasks[idx].Synopsis
	}
	t.compactUI.SetTaskList(taskList)
	return nil
}

func (t *timetrackerWindowData) setRunningTimesheet(tsd *models.TimesheetData) {
	log := logger.GetFuncLogger(t.log, "setRunningTimesheet")
	running := false
	taskName := ""
	elapsedTime := ""
	if tsd != nil {
		running = true
		taskName = tsd.Task.Synopsis
		elapsedTime = t.elapsedTime(tsd.StartTime)
	}
	t.compactUI.SetRunning(running)
	t.compactUI.SetTaskName(taskName)
	t.compactUI.SetElapsedTime(elapsedTime)
	// Refresh task list
	err := t.refreshTaskList()
	if err != nil {
		log.Err(err).
			Msg("unable to refresh task list")
	}
	// TODO: Should we be changing the selectedTask here?
	// t.selectedTask = models.NewTaskWithData(runningTS.Task)
	switch running {
	case true:
		if !t.elapsedTimeRunning {
			// Start the elapsed time counter
			go t.elapsedTimeLoop(tsd.StartTime, t.elapsedTimeQuitChan)
		}
	case false:
		if t.elapsedTimeRunning {
			t.elapsedTimeQuitChan <- true
		}
	}
}

func (t *timetrackerWindowData) runningTimesheetChanged(item interface{}) {
	// log := logger.GetFuncLogger(t.log, "runningTimesheetChanged")
	runningTS, ok := item.(*models.TimesheetData)
	if ok {
		t.setRunningTimesheet(runningTS)
	}
}

func (t *timetrackerWindowData) doCreateAndStartTask() {
	t.createNewTaskAndStartDialog.HideCloseWindowCheckbox()
	t.createNewTaskAndStartDialog.Show()
}

func (t *timetrackerWindowData) doStartSelectedTask() {
	log := logger.GetFuncLogger(t.log, "doStartSelectedTask")
	runningTS, err := models.NewTimesheet().RunningTimesheet()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("unable to get the running timesheet")
		return
	}
	if runningTS != nil {
		stopTaskDialog := dialogs.NewStopTaskDialog(
			runningTS.Data().Task,
			(*t.app).Preferences(),
			func(doStop bool) {
				if doStop {
					t.doStopAndStartTask()
				}
			},
			t.Window,
		)
		stopTaskDialog.SetCloseWindowCheckbox(true)
		stopTaskDialog.Show()
		return
	}
	t.doStartTask()
}

// doStartTask starts the task in t.selectedTask. It assumes there are no tasks running.
func (t *timetrackerWindowData) doStartTask() {
	log := logger.GetFuncLogger(t.log, "doStartTask")
	log.Trace().Msg("started")
	defer log.Trace().Msg("done")
	if t.selectedTask == nil {
		log.Error().
			Msg("no task was selected")
		dialog.NewError(
			fmt.Errorf("please select a task to start"), // i18n
			t.Window,
		).Show()
		return
	}
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *t.selectedTask.Data()
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
		fmt.Sprintf("Task %s started", t.selectedTask.Data().Synopsis),              // i18n
		fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp)), // i18n
	)
	t.runningTimesheet = timesheet.Data()
	t.runningTimesheetChanged(t.runningTimesheet)
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
	t.runningTimesheet = nil
	t.runningTimesheetChanged(t.runningTimesheet)
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
	t.selectedTask = selectedTask
	t.compactUI.SelectTask(selectedTask.Data().Synopsis)
	t.doStartSelectedTask()
}

func (t *timetrackerWindowData) handleTaskSelectorEvent(item interface{}) {
	log := logger.GetFuncLogger(t.log, "handleTaskSelectorEvent")
	switch event := item.(type) {
	case widgets.TaskSelectorSelectedEvent:
		if event.SelectedTask != nil {
			log.Debug().
				Str("selected", event.SelectedTask.String()).
				Msg("got selected task")
		}
	case widgets.TaskSelectorErrorEvent:
		if event.Err != nil {
			log.Err(event.Err).
				Msg("error from task selector")
		}
	}
}

func (t *timetrackerWindowData) handleCompactUIEvent(item interface{}) {
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
		t.handleCompactUITaskEvent(event.TaskIndex, event.TaskSynopsis)
	}
}

func (t *timetrackerWindowData) handleCompactUITaskEvent(index int, synopsis string) {
	log := logger.GetFuncLogger(t.log, "handleCompactUITaskEvent")
	if index == -1 || synopsis == "" {
		t.doStopTask()
		return
	}
	tasks, err := models.NewTask().SearchBySynopsis(synopsis)
	if err != nil {
		log.Err(err).
			Str("synopsis", synopsis).
			Msg("unable to find task by synopsis")
		return
	}
	if len(tasks) == 0 {
		log.Error().
			Msg("did not find task by synopsis")
		return
	}
	t.selectedTask = models.NewTaskWithData(tasks[0])
	t.doStopAndStartTask()
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

// Show shows the main window
func (t *timetrackerWindowData) Show() {
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

// ShowAbout shows the about dialog box
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
	// Check if elapsed time counter is running and stop it if it is
	if t.elapsedTimeRunning {
		t.elapsedTimeQuitChan <- true
	}
	// Close the window
	t.Window.Close()
	// Quit
	(*t.app).Quit()
}

func (t *timetrackerWindowData) maybeStopRunningTask(stopTask bool) {
	log := logger.GetFuncLogger(t.log, "maybeStopRunningTask")
	if !stopTask {
		return
	}
	stoppedTimesheet, err := models.NewTask().StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("error stopping the running task")
		dialog.NewError(err, t.Window).Show()
	}
	if stoppedTimesheet != nil {
		// Show notification that the task has stopped
		t.notify(
			fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis),                  // i18n
			fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)), // i18n
		)
		t.runningTimesheet = nil
		t.runningTimesheetChanged(t.runningTimesheet)
	}
	// Check if we should close the main window
	shouldCloseMainWindow := (*t.app).Preferences().BoolWithFallback(constants.PrefKeyCloseWindowStopTask, false)
	if shouldCloseMainWindow {
		t.Close()
	}
}

func (t *timetrackerWindowData) createAndStartTaskDialogCallback(createAndStart bool) {
	log := logger.GetFuncLogger(t.log, "createAndStartTaskDialogCallback")
	if !createAndStart {
		return
	}
	taskData := t.createNewTaskAndStartDialog.GetTask()
	if taskData == nil {
		log.Error().
			Msg("taskData was nil; this is unexpected")
		return
	}
	// Create the new task
	err := taskData.Create()
	if err != nil {
		log.Err(err).
			Str("newTask", taskData.String()).
			Msgf("error creating new task")
		return
	}
	log.Debug().
		Str("newTask", taskData.String()).
		Msgf("created new task")
	// reset the create dialog now that the task has been created
	t.createNewTaskAndStartDialog.Reset()
	// Stop the running task
	stoppedTimesheet, err := taskData.StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("error stopping current task")
		return
	}
	// If a running task was actually stopped...
	if stoppedTimesheet != nil {
		log.Debug().
			Str("task", stoppedTimesheet.String()).
			Msg("stopped running task")
		// Show notification that task has stopped
		t.notify(
			fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis),                  // i18n
			fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)), // i18n
		)
	}
	// Start the new task
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *taskData
	timesheet.Data().StartTime = time.Now()
	err = timesheet.Create()
	if err != nil {
		log.Err(err).
			Msg("unable to start new task")
		return
	}
	// Show notification that task has started
	t.notify(
		fmt.Sprintf("Task %s started", taskData.Synopsis),                           // i18n
		fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp)), // i18n
	)
	log.Debug().
		Str("task", taskData.String()).
		Str("startTime", timesheet.Data().StartTime.String()).
		Msg("task started")
	t.runningTimesheet = timesheet.Data()
	t.runningTimesheetChanged(t.runningTimesheet)
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
