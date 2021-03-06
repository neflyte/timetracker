package gui

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
	"github.com/rs/zerolog"
	"regexp"
	"strconv"
	"time"
)

// TODO: rework the UI layout to better contain the components (e.g. use themes and a custom layout, not just a card)
// TODO: implement an "elapsed time" counter if a task is running

var (
	taskNameRE = regexp.MustCompile(`^\[([0-9]+)].*$`)
)

// timetrackerWindow is the main timetracker GUI window interface
type timetrackerWindow interface {
	Show()
	ShowAbout()
	ShowWithError(err error)
	ShowWithManageWindow()
	ShowAndStopRunningTask()
	ShowAndDisplayCreateAndStartDialog()
	Hide()
	Close()
	Get() timetrackerWindowData
}

type timetrackerWindowData struct {
	App                         *fyne.App
	Window                      fyne.Window
	Container                   *fyne.Container
	StatusBox                   *fyne.Container
	SubStatusBox                *fyne.Container
	ButtonBox                   *fyne.Container
	TaskList                    *widgets.Tasklist
	BtnStartTask                *widget.Button
	BtnStopTask                 *widget.Button
	BtnManageTasks              *widget.Button
	BtnAbout                    *widget.Button
	BtnQuit                     *widget.Button
	createNewTaskAndStartDialog *dialogs.CreateAndStartTaskDialog
	Log                         zerolog.Logger
	mngWindow                   manageWindow

	LblStatus      *widget.Label
	LblStartTime   *widget.Label
	LblElapsedTime *widget.Label

	BindRunningTask binding.String
	BindStartTime   binding.String
	BindElapsedTime binding.String
}

// newTimetrackerWindow creates and initializes a new timetracker window
func newTimetrackerWindow(app fyne.App) timetrackerWindow {
	ttw := &timetrackerWindowData{
		App:    &app,
		Window: app.NewWindow("Timetracker"),
		Log:    logger.GetStructLogger("timetrackerWindow"),
	}
	ttw.initWindow()
	return ttw
}

// initWindow initializes the window
func (t *timetrackerWindowData) initWindow() {
	log := logger.GetFuncLogger(t.Log, "initWindow")
	log.Debug().Msg("started")
	t.createNewTaskAndStartDialog = dialogs.NewCreateAndStartTaskDialog(t.createAndStartTaskDialogCallback, t.Window)
	t.TaskList = widgets.NewTasklist()
	t.TaskList.SelectionBinding().AddListener(binding.NewDataListener(func() {
		t.tasklistSelectionChanged()
	}))
	t.BtnStartTask = widget.NewButtonWithIcon("START", theme.MediaPlayIcon(), t.doStartTask)
	t.BtnStopTask = widget.NewButtonWithIcon("STOP", theme.MediaStopIcon(), t.doStopTask)
	t.BtnManageTasks = widget.NewButtonWithIcon("MANAGE", theme.SettingsIcon(), t.doManageTasks)
	t.BtnAbout = widget.NewButton("ABOUT", t.doAbout)
	t.BtnQuit = widget.NewButton("QUIT", t.doQuit)
	t.ButtonBox = container.NewCenter(
		container.NewHBox(
			t.BtnStartTask,
			t.BtnStopTask,
			t.BtnManageTasks,
			t.BtnAbout,
			t.BtnQuit,
		),
	)
	t.BindRunningTask = binding.NewString()
	t.LblStatus = widget.NewLabelWithData(t.BindRunningTask)
	t.LblStatus.TextStyle = fyne.TextStyle{
		Monospace: true,
	}
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
	// Make sure we hide the window instead of closing it, otherwise the app will exit
	t.Window.SetCloseIntercept(t.Hide)
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
			log.Err(err).Msg("error getting running timesheet from rxgo observable")
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
			return
		}
		// A task is running
		t.BtnStopTask.Enable()
		t.BtnStartTask.Disable()
		err := t.BindRunningTask.Set(runningTS.Task.String())
		if err != nil {
			log.Err(err).Msgf("error setting running task to %s", runningTS.Task.String())
		}
		err = t.BindStartTime.Set(runningTS.StartTime.String())
		if err != nil {
			log.Err(err).Msgf("error setting start time to %s", runningTS.StartTime.String())
		}
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
	// TODO: convert from selectedTask string to task ID so we can start a new timesheet
	if taskNameRE.MatchString(selection) {
		var taskIDInt int
		matches := taskNameRE.FindStringSubmatch(selection)
		taskIDString := matches[1]
		taskIDInt, err = strconv.Atoi(taskIDString)
		if err != nil {
			log.Err(err).Msgf("err converting taskIDString '%s' to int", taskIDString)
			dialog.NewError(err, t.Window).Show()
			return
		}
		taskData := models.NewTaskData()
		taskData.ID = uint(taskIDInt)
		err = models.Task(taskData).Load(false)
		if err != nil {
			log.Err(err).Msgf("error loading task id %d", taskData.ID)
			dialog.NewError(err, t.Window).Show()
			return
		}
		timesheetData := new(models.TimesheetData)
		timesheetData.Task = *taskData
		timesheetData.StartTime = time.Now()
		err = models.Timesheet(timesheetData).Create()
		if err != nil {
			log.Err(err).Msg("error creating new timesheet")
			dialog.NewError(err, t.Window).Show()
			return
		}
		t.BtnStopTask.Enable()
		t.BtnStartTask.Disable()
		appstate.SetRunningTimesheet(timesheetData)
	}
	log.Trace().Msg("done")
}

func (t *timetrackerWindowData) doStopTask() {
	log := logger.GetFuncLogger(t.Log, "doStopTask")
	// Stop the running task
	log.Debug().Msg("stopping running task")
	_, err := models.Task(new(models.TaskData)).StopRunningTask()
	if err != nil && !errors.Is(err, tterrors.ErrNoRunningTask{}) {
		log.Err(err).Msg(tterrors.StopRunningTaskError)
		dialog.NewError(err, t.Window).Show()
	}
	// Update the appstate
	appstate.SetRunningTimesheet(nil)
}

func (t *timetrackerWindowData) doManageTasks() {
	t.mngWindow.Show()
}

func (t *timetrackerWindowData) doQuit() {
	StopGUI()
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
	timesheetData := new(models.TimesheetData)
	openTimesheets, searchErr := timesheetData.SearchOpen()
	if searchErr != nil {
		t.ShowWithError(searchErr)
		return
	}
	if len(openTimesheets) == 0 {
		t.ShowWithError(fmt.Errorf("a task is not running; please start a task first"))
		return
	}
	t.Show()
	confirmMessage := fmt.Sprintf("Do you want to stop task %s?", openTimesheets[0].Task.Synopsis)
	dialog.NewConfirm("Stop Running Task", confirmMessage, t.maybeStopRunningTask, t.Window).Show()
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
	t.Window.Close()
}

// Get returns the underlying data structure
func (t *timetrackerWindowData) Get() timetrackerWindowData {
	return *t
}

func (t *timetrackerWindowData) maybeStopRunningTask(stopTask bool) {
	log := logger.GetFuncLogger(t.Log, "maybeStopRunningTask")
	if !stopTask {
		return
	}
	td := new(models.TaskData)
	_, err := td.StopRunningTask()
	if err != nil {
		log.Err(err).Msg("error stopping the running task")
		dialog.NewError(err, t.Window).Show()
	}
	appstate.SetRunningTimesheet(nil)
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
	}
	// Start the new task
	timesheetData := new(models.TimesheetData)
	timesheetData.Task = *taskData
	timesheetData.StartTime = time.Now()
	err = timesheetData.Create()
	if err != nil {
		log.Err(err).Msg("error stopping current task")
		return
	}
	log.Debug().Msgf("task %s started at %s", taskData.String(), timesheetData.StartTime.String())
	appstate.SetRunningTimesheet(timesheetData)
}
