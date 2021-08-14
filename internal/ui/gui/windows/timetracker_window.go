package windows

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/appstate"
	tterrors "github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/dialogs"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/rs/zerolog"
	"regexp"
	"sync"
	"time"
)

// TODO: rework the UI layout to better contain the components (e.g. use themes and a custom layout, not just a card)

const (
	// taskNameTrimLength is the maximum length of the task name string before trimming
	taskNameTrimLength = 32
)

var (
	// taskNameRE is a regular expression used to split the Synopsis from the Description in
	// the Tasklist widget
	taskNameRE = regexp.MustCompile(`^(.*): .*$`)
)

// TimetrackerWindow is the main timetracker GUI window interface
type TimetrackerWindow interface {
	windowBase

	ShowAbout()
	ShowWithError(err error)
	ShowWithManageWindow()
	ShowAndStopRunningTask()
	ShowAndDisplayCreateAndStartDialog()
	Get() *timetrackerWindowData
}

// timetrackerWindowData is the struct underlying the TimetrackerWindow interface
type timetrackerWindowData struct {
	fyne.Window

	App                         *fyne.App
	Container                   *fyne.Container
	StatusBox                   *fyne.Container
	SubStatusBox                *fyne.Container
	ButtonBox                   *fyne.Container
	TaskList                    *widgets.Tasklist
	BtnStartTask                *widget.Button
	BtnStopTask                 *widget.Button
	BtnManageTasks              *widget.Button
	BtnReport                   *widget.Button
	BtnAbout                    *widget.Button
	createNewTaskAndStartDialog dialogs.CreateAndStartTaskDialog
	Log                         zerolog.Logger
	mngWindow                   manageWindow
	rptWindow                   reportWindow

	LblStatus      *widget.Label
	LblStartTime   *widget.Label
	LblElapsedTime *widget.Label

	BindRunningTask binding.String
	BindStartTime   binding.String
	BindElapsedTime binding.String

	elapsedTimeTicker       *time.Ticker
	elapsedTimeRunningMutex sync.RWMutex
	elapsedTimeRunning      bool
	elapsedTimeQuitChan     chan bool
}

// NewTimetrackerWindow creates and initializes a new timetracker window
func NewTimetrackerWindow(app fyne.App) TimetrackerWindow {
	ttw := &timetrackerWindowData{
		App:                     &app,
		Window:                  app.NewWindow("Timetracker"),
		Log:                     logger.GetStructLogger("timetrackerWindow"),
		elapsedTimeRunningMutex: sync.RWMutex{},
		elapsedTimeRunning:      false,
		elapsedTimeQuitChan:     make(chan bool, 1),
	}
	err := ttw.Init()
	if err != nil {
		ttw.Log.Err(err).Msg("error initializing window")
	}
	return ttw
}

// Init initializes the window
func (t *timetrackerWindowData) Init() error {
	log := logger.GetFuncLogger(t.Log, "initWindow")
	log.Debug().Msg("started")
	if t.App == nil {
		log.Error().Msg("t.App was nil; THIS IS UNEXPECTED")
		return errors.New("t.App was nil; THIS IS UNEXPECTED")
	}
	t.createNewTaskAndStartDialog = dialogs.NewCreateAndStartTaskDialog((*t.App).Preferences(), t.createAndStartTaskDialogCallback, t.Window)
	t.TaskList = widgets.NewTasklist()
	t.TaskList.SelectionBinding().AddListener(binding.NewDataListener(func() {
		t.tasklistSelectionChanged()
	}))
	t.BtnStartTask = widget.NewButtonWithIcon("START", theme.MediaPlayIcon(), t.doStartTask)
	t.BtnStopTask = widget.NewButtonWithIcon("STOP", theme.MediaStopIcon(), t.doStopTask)
	t.BtnManageTasks = widget.NewButtonWithIcon("MANAGE", theme.SettingsIcon(), t.doManageTasks)
	t.BtnReport = widget.NewButtonWithIcon("REPORT", theme.FileTextIcon(), t.doReport)
	t.BtnAbout = widget.NewButton("ABOUT", t.doAbout)
	t.ButtonBox = container.NewCenter(
		container.NewHBox(
			t.BtnStartTask,
			t.BtnStopTask,
			t.BtnManageTasks,
			t.BtnReport,
			t.BtnAbout,
		),
	)
	t.BindRunningTask = binding.NewString()
	t.LblStatus = widget.NewLabelWithData(t.BindRunningTask)
	t.BindStartTime = binding.NewString()
	t.LblStartTime = widget.NewLabelWithData(t.BindStartTime)
	t.BindElapsedTime = binding.NewString()
	t.LblElapsedTime = widget.NewLabelWithData(t.BindElapsedTime)
	t.SubStatusBox = container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Start time:"),
			t.LblStartTime,
		),
		container.NewHBox(
			widget.NewLabel("Elapsed time:"),
			t.LblElapsedTime,
		),
	)
	t.SubStatusBox.Hide()
	t.StatusBox = container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle(
				"Running Task:",
				fyne.TextAlignLeading,
				fyne.TextStyle{
					Bold: true,
				},
			),
			t.LblStatus,
		),
		t.SubStatusBox,
	)
	t.Container = container.NewPadded(
		widget.NewCard(
			"Timetracker",
			"",
			container.NewPadded(container.NewVBox(
				t.StatusBox,
				widget.NewSeparator(),
				container.NewHBox(
					widget.NewLabel("Task:"),
					t.TaskList,
				),
				t.ButtonBox,
			)),
		),
	)
	log.Debug().Msg("set content")
	t.Window.SetContent(t.Container)
	t.Window.SetFixedSize(true)
	t.Window.Resize(minimumWindowSize)
	t.Window.SetCloseIntercept(t.Close)
	// Set up our observables
	t.setupObservables()
	// Load the window's data
	t.primeWindowData()
	// Set up the Manage Window as well
	t.mngWindow = newManageWindow(*t.App)
	t.mngWindow.Get().TaskListChangedObservable.ForEach(
		func(item interface{}) {
			changed, ok := item.(bool)
			if ok && changed {
				t.TaskList.Refresh()
			}
		},
		func(err error) {
			t.Log.Err(err).Msg("error from tasklist changed observable")
		},
		func() {
			t.Log.Trace().Msg("tasklist changed observable is finished")
		},
	)
	// Hide the Manage Window by default
	t.mngWindow.Hide()
	// Also set up the report window and hide it
	t.rptWindow = newReportWindow(*t.App)
	t.rptWindow.Hide()
	return nil
}

// primeWindowData primes the window with some data
func (t *timetrackerWindowData) primeWindowData() {
	log := logger.GetFuncLogger(t.Log, "primeWindowData")
	log.Debug().Msg("started")
	// Load the running task
	runningTS := appstate.GetRunningTimesheet()
	if runningTS != nil {
		// Task is running
		t.BtnStopTask.Enable()
		newSelectedTask := runningTS.Task.String()
		if newSelectedTask != "" {
			t.TaskList.SetSelected(newSelectedTask)
			t.BtnStartTask.Disable()
		} else {
			t.BtnStartTask.Enable()
		}
		// Start elapsed time counter
		go t.elapsedTimeLoop(runningTS.StartTime, t.elapsedTimeQuitChan)
	} else {
		// Task is not running
		t.BtnStopTask.Disable()
	}
	log.Debug().Msg("done")
}

func (t *timetrackerWindowData) setupObservables() {
	log := logger.GetFuncLogger(t.Log, "setupObservables")
	appstate.Observables()[appstate.KeyRunningTimesheet].ForEach(
		t.runningTimesheetChanged,
		func(err error) {
			log.Err(err).Msg("error getting running timesheet from RxGO observable")
		},
		func() {
			log.Trace().Msg("running timesheet observable is done")
		},
	)
}

func (t *timetrackerWindowData) tasklistSelectionChanged() {
	log := logger.GetFuncLogger(t.Log, "tasklistSelectionChanged")
	selection, err := t.TaskList.SelectionBinding().Get()
	if err != nil {
		log.Err(err).Msg("error getting selection from binding")
		return
	}
	if selection != "" {
		if appstate.GetRunningTimesheet() == nil {
			t.BtnStartTask.Enable()
		} else {
			t.BtnStartTask.Disable()
		}
		return
	}
	t.BtnStartTask.Disable()
}

func (t *timetrackerWindowData) runningTimesheetChanged(item interface{}) {
	log := logger.GetFuncLogger(t.Log, "runningTimesheetChanged")
	runningTS, ok := item.(*models.TimesheetData)
	if ok {
		if runningTS == nil {
			// No task is running
			err := t.BindRunningTask.Set("none")
			if err != nil {
				log.Err(err).Msg("error setting running task to none")
			}
			t.SubStatusBox.Hide()
			t.BtnStopTask.Disable()
			selection, err := t.TaskList.SelectionBinding().Get()
			if err != nil {
				log.Err(err).Msg("error getting selection from binding")
			}
			if selection != "" {
				// A task is selected
				t.BtnStartTask.Enable()
			} else {
				// No task is selected
				t.BtnStartTask.Disable()
			}
			// Stop the elapsed time counter if it's running
			t.elapsedTimeRunningMutex.RLock()
			defer t.elapsedTimeRunningMutex.RUnlock()
			if t.elapsedTimeRunning {
				t.elapsedTimeQuitChan <- true
			}
			return
		}
		// A task is running
		t.BtnStopTask.Enable()
		t.BtnStartTask.Disable()
		err := t.BindRunningTask.Set(utils.TrimWithEllipsis(runningTS.Task.String(), taskNameTrimLength))
		if err != nil {
			log.Err(err).Msgf("error setting running task to %s", runningTS.Task.String())
		}
		startTimeDisplay := runningTS.StartTime.Format(time.RFC1123Z)
		err = t.BindStartTime.Set(startTimeDisplay)
		if err != nil {
			log.Err(err).Msgf("error setting start time to %s", startTimeDisplay)
		}
		elapsedTimeDisplay := time.Since(runningTS.StartTime).Truncate(time.Second).String()
		err = t.BindElapsedTime.Set(elapsedTimeDisplay)
		if err != nil {
			log.Err(err).Msgf("error setting elapsed time to %s", elapsedTimeDisplay)
		}
		// Start the elapsed time counter
		go t.elapsedTimeLoop(runningTS.StartTime, t.elapsedTimeQuitChan)
		t.SubStatusBox.Show()
	}
}

func (t *timetrackerWindowData) doStartTask() {
	log := logger.GetFuncLogger(t.Log, "doStartTask")
	log.Trace().Msg("started")
	selection, err := t.TaskList.SelectionBinding().Get()
	if err != nil {
		dialog.NewError(err, t.Window).Show()
		log.Err(err).Msg("error getting selection from binding")
		return
	}
	if selection == "" {
		log.Error().Msg("no task was selected")
		dialog.NewError(
			fmt.Errorf("please select a task to start"),
			t.Window,
		).Show()
		return
	}
	// TODO: convert from selectedTask string to task ID so we can start a new timesheet (instead of parsing a string)
	if taskNameRE.MatchString(selection) {
		matches := taskNameRE.FindStringSubmatch(selection)
		taskSynopsisString := matches[1]
		task := models.NewTask()
		task.Data().Synopsis = taskSynopsisString
		err = task.Load(false)
		if err != nil {
			log.Err(err).Msgf("error loading task with synopsis '%s'", task.Data().Synopsis)
			dialog.NewError(err, t.Window).Show()
			return
		}
		timesheet := models.NewTimesheet()
		timesheet.Data().Task = *task.Data()
		timesheet.Data().StartTime = time.Now()
		err = timesheet.Create()
		if err != nil {
			log.Err(err).Msg("error creating new timesheet")
			dialog.NewError(err, t.Window).Show()
			return
		}
		// Show notification that task started
		notificationTitle := fmt.Sprintf("Task %s started", task.Data().Synopsis)
		notificationContents := fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp))
		(*t.App).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
		t.BtnStopTask.Enable()
		t.BtnStartTask.Disable()
		appstate.SetRunningTimesheet(timesheet.Data())
	}
	log.Trace().Msg("done")
}

func (t *timetrackerWindowData) doStopTask() {
	log := logger.GetFuncLogger(t.Log, "doStopTask")
	// Stop the running task
	log.Debug().Msg("stopping running task")
	stoppedTimesheet, err := models.NewTask().StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).Msg(tterrors.StopRunningTaskError)
		dialog.NewError(err, t.Window).Show()
	}
	// Show notification that task has stopped
	notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)
	notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp))
	(*t.App).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	// Update the appstate
	appstate.SetRunningTimesheet(nil)
}

func (t *timetrackerWindowData) doManageTasks() {
	t.mngWindow.Show()
}

func (t *timetrackerWindowData) doReport() {
	t.rptWindow.Show()
}

func (t *timetrackerWindowData) doAbout() {
	appVersion := "??"
	appVersionIntf, ok := appstate.Map().Load(appstate.KeyAppVersion)
	if ok {
		appVersionStr, isString := appVersionIntf.(string)
		if !isString {
			appVersionStr = "!!"
		}
		if appVersionStr != "" {
			appVersion = appVersionStr
		}
	}
	dialog.NewInformation(
		"About Timetracker",
		fmt.Sprintf("Timetracker %s\n\nhttps://github.com/neflyte/timetracker", appVersion),
		t.Window,
	).Show()
}

// Show shows the main window
func (t *timetrackerWindowData) Show() {
	t.Window.Show()
	t.TaskList.Refresh()
}

// ShowAndStopRunningTask shows the main window and asks the user if they want to stop the running task
func (t *timetrackerWindowData) ShowAndStopRunningTask() {
	openTimesheets, searchErr := models.NewTimesheet().SearchOpen()
	if searchErr != nil {
		t.ShowWithError(searchErr)
		return
	}
	if len(openTimesheets) == 0 {
		t.ShowWithError(fmt.Errorf("a task is not running; please start a task first"))
		return
	}
	t.Show()
	dialogs.NewStopTaskDialog(openTimesheets[0].Task, (*t.App).Preferences(), t.maybeStopRunningTask, t.Window).Show()
}

// ShowWithManageWindow shows the main window followed by the Manage window
func (t *timetrackerWindowData) ShowWithManageWindow() {
	t.Show()
	t.doManageTasks()
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
	if t.mngWindow != nil {
		t.mngWindow.Hide()
	}
	t.Window.Hide()
}

// Close closes the main window
func (t *timetrackerWindowData) Close() {
	// Check if elapsed time counter is running and stop it if it is
	t.elapsedTimeRunningMutex.RLock()
	defer t.elapsedTimeRunningMutex.RUnlock()
	if t.elapsedTimeRunning {
		t.elapsedTimeQuitChan <- true
	}
	// Close the window
	t.Window.Close()
}

// Get returns the underlying data structure
func (t *timetrackerWindowData) Get() *timetrackerWindowData {
	return t
}

func (t *timetrackerWindowData) maybeStopRunningTask(stopTask bool) {
	log := logger.GetFuncLogger(t.Log, "maybeStopRunningTask")
	if !stopTask {
		return
	}
	stoppedTimesheet, err := models.NewTask().StopRunningTask()
	if err != nil {
		log.Err(err).Msg("error stopping the running task")
		dialog.NewError(err, t.Window).Show()
	}
	// Show notification that the task has stopped
	notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)
	notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp))
	(*t.App).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	appstate.SetRunningTimesheet(nil)
	// Check if we should close the main window
	shouldCloseMainWindow := (*t.App).Preferences().BoolWithFallback(prefKeyCloseWindowStopTask, false)
	if shouldCloseMainWindow {
		t.Close()
	}
}

func (t *timetrackerWindowData) createAndStartTaskDialogCallback(createAndStart bool) {
	log := logger.GetFuncLogger(t.Log, "createAndStartTaskDialogCallback")
	if !createAndStart {
		return
	}
	taskData := t.createNewTaskAndStartDialog.GetTask()
	if taskData == nil {
		log.Error().Msg("taskData was nil; this is unexpected")
		return
	}
	// Create the new task
	err := taskData.Create()
	if err != nil {
		log.Err(err).Msgf("error creating new task %s", taskData.String())
		return
	}
	log.Debug().Msgf("created new task %s", taskData.String())
	t.createNewTaskAndStartDialog.Reset()
	// Stop the running task
	stoppedTimesheet, err := taskData.StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).Msg("error stopping current task")
		return
	}
	if stoppedTimesheet != nil {
		log.Debug().Msgf("stopped running task %s", stoppedTimesheet.String())
		// Show notification that task has stopped
		notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)
		notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp))
		(*t.App).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	}
	// Start the new task
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *taskData
	timesheet.Data().StartTime = time.Now()
	err = timesheet.Create()
	if err != nil {
		log.Err(err).Msg("error stopping current task")
		return
	}
	// Show notification that task has started
	notificationTitle := fmt.Sprintf("Task %s started", taskData.Synopsis)
	notificationContents := fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp))
	(*t.App).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	log.Debug().Msgf("task %s started at %s", taskData.String(), timesheet.Data().StartTime.String())
	appstate.SetRunningTimesheet(timesheet.Data())
	// Check if we should close the main window as well
	shouldCloseMainWindow := (*t.App).Preferences().BoolWithFallback(prefKeyCloseWindow, false)
	if shouldCloseMainWindow {
		t.Close()
	}
}

// elapsedTimeLoop is a loop that draws the elapsed time since the running task was started
func (t *timetrackerWindowData) elapsedTimeLoop(startTime time.Time, quitChan chan bool) {
	log := logger.GetFuncLogger(t.Log, "elapsedTimeLoop")
	t.elapsedTimeRunningMutex.RLock()
	if t.elapsedTimeRunning {
		t.elapsedTimeRunningMutex.RUnlock()
		return
	}
	t.elapsedTimeRunningMutex.RUnlock()
	t.elapsedTimeRunningMutex.Lock()
	t.elapsedTimeRunning = true
	t.elapsedTimeRunningMutex.Unlock()
	defer func() {
		t.elapsedTimeRunningMutex.Lock()
		t.elapsedTimeRunning = false
		t.elapsedTimeRunningMutex.Unlock()
	}()
	t.elapsedTimeTicker = time.NewTicker(time.Second)
	defer t.elapsedTimeTicker.Stop()
	// Clear the elapsed time display when the loop ends
	defer func() {
		err := t.BindElapsedTime.Set("")
		if err != nil {
			log.Err(err).Msg("error setting elapsed time display to empty")
		}
	}()
	log.Debug().Msg("loop starting")
	defer log.Debug().Msg("loop ending")
	for {
		select {
		case <-t.elapsedTimeTicker.C:
			elapsedTime := time.Since(startTime).Truncate(time.Second).String()
			err := t.BindElapsedTime.Set(elapsedTime)
			if err != nil {
				log.Err(err).Msgf("error setting elapsed time binding to %s: %s", elapsedTime, err.Error())
			}
		case <-quitChan:
			return
		}
	}
}
