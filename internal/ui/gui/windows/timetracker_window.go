package windows

import (
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
)

const (
	// taskNameTrimLength is the maximum length of the task name string before trimming
	taskNameTrimLength = 32
	// windowHeightBuffer is the number of pixels to increase a window height by to try to fit its contents correctly
	windowHeightBuffer = 50
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
	fyne.Window
	app                         *fyne.App
	log                         zerolog.Logger
	container                   *fyne.Container
	statusBox                   *fyne.Container
	subStatusBox                *fyne.Container
	buttonBox                   *fyne.Container
	btnSelectTask               *widget.Button
	btnCreateAndStart           *widget.Button
	btnStartTask                *widget.Button
	btnStopTask                 *widget.Button
	btnManageTasksV2            *widget.Button
	btnReport                   *widget.Button
	btnAbout                    *widget.Button
	createNewTaskAndStartDialog dialogs.CreateAndStartTaskDialog
	mngWindowV2                 manageWindowV2
	rptWindow                   reportWindow
	taskSelector                *widgets.TaskSelector
	lblStatus                   *widget.Label
	lblStartTime                *widget.Label
	lblElapsedTime              *widget.Label
	bindRunningTask             binding.String
	bindStartTime               binding.String
	bindElapsedTime             binding.String
	textSelectedTask            *canvas.Text
	selectedTaskBinding         binding.String
	selectedTask                models.Task
	elapsedTimeTicker           *time.Ticker
	elapsedTimeRunning          bool
	elapsedTimeQuitChan         chan bool
}

// NewTimetrackerWindow creates and initializes a new timetracker window
func NewTimetrackerWindow(app fyne.App) TimetrackerWindow {
	ttw := &timetrackerWindowData{
		app:                 &app,
		Window:              app.NewWindow("Timetracker"),
		log:                 logger.GetStructLogger("timetrackerWindowData"),
		bindElapsedTime:     binding.NewString(),
		bindStartTime:       binding.NewString(),
		bindRunningTask:     binding.NewString(),
		selectedTaskBinding: binding.NewString(),
		elapsedTimeRunning:  false,
		elapsedTimeQuitChan: make(chan bool, 1),
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
	if t.app == nil {
		return errors.New("t.app was nil; this is unexpected")
	}
	t.btnStartTask = widget.NewButtonWithIcon("START", theme.MediaPlayIcon(), t.doStartTask) // i18n
	t.btnStopTask = widget.NewButtonWithIcon("STOP", theme.MediaStopIcon(), t.doStopTask)    // i18n
	t.createNewTaskAndStartDialog = dialogs.NewCreateAndStartTaskDialog((*t.app).Preferences(), t.createAndStartTaskDialogCallback, t.Window)
	t.btnSelectTask = widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), t.doSelectTask)
	t.taskSelector = widgets.NewTaskSelector()
	t.btnManageTasksV2 = widget.NewButtonWithIcon("MANAGE v2", theme.SettingsIcon(), t.doManageTasksV2) // i18n
	t.btnReport = widget.NewButtonWithIcon("REPORT", theme.FileIcon(), t.doReport)                      // i18n
	t.btnAbout = widget.NewButton("ABOUT", t.doAbout)                                                   // i18n
	t.btnCreateAndStart = widget.NewButton("CREATE AND START", t.doCreateAndStartTask)                  // i18n
	t.buttonBox = container.NewCenter(container.NewVBox(
		container.NewHBox(
			t.btnStartTask,
			t.btnStopTask,
			t.btnManageTasksV2,
			t.btnReport,
			t.btnAbout,
		),
		container.NewHBox(t.btnCreateAndStart),
	))
	t.textSelectedTask = canvas.NewText("", theme.ForegroundColor())
	t.selectedTaskBinding.AddListener(binding.NewDataListener(t.selectedTaskChanged))
	t.lblStatus = widget.NewLabelWithData(t.bindRunningTask)
	t.lblStartTime = widget.NewLabelWithData(t.bindStartTime)
	t.lblElapsedTime = widget.NewLabelWithData(t.bindElapsedTime)
	t.subStatusBox = container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Start time:"), // i18n
			t.lblStartTime,
		),
		container.NewHBox(
			widget.NewLabel("Elapsed time:"), // i18n
			t.lblElapsedTime,
		),
	)
	t.statusBox = container.NewVBox(
		t.lblStatus,
		t.subStatusBox,
	)
	t.container = container.NewPadded(
		container.NewVBox(
			t.statusBox,
			widget.NewSeparator(),
			container.NewBorder(
				nil,
				nil,
				container.NewHBox(widget.NewLabel("Task:"), t.textSelectedTask), // i18n
				t.btnSelectTask,
			),
			t.buttonBox,
		),
	)
	t.Window.SetContent(t.container)
	t.Window.SetIcon(icons.IconV2)
	// get the size of the content with everything visible
	siz := t.Window.Content().Size()
	// HACK: add a bit of a height buffer, so we can try to fit everything in the window nicely
	siz.Height += float32(windowHeightBuffer)
	// resize the window to fit the content
	t.Window.Resize(siz)
	// hide stuff now that we resized
	t.subStatusBox.Hide()
	t.Window.SetCloseIntercept(t.Close)
	// Set up our observables
	appstate.Observables()[appstate.KeyRunningTimesheet].ForEach(
		t.runningTimesheetChanged,
		utils.ObservableErrorHandler(appstate.KeyRunningTimesheet, t.log),
		utils.ObservableCloseHandler(appstate.KeyRunningTimesheet, t.log),
	)
	// Load the window's data
	t.primeWindowData()
	// Also set up the manage window and hide it
	t.mngWindowV2 = newManageWindowV2(*t.app)
	t.mngWindowV2.Hide()
	// Also set up the report window and hide it
	t.rptWindow = newReportWindow(*t.app)
	t.rptWindow.Hide()
	return nil
}

// primeWindowData primes the window with some data
func (t *timetrackerWindowData) primeWindowData() {
	log := logger.GetFuncLogger(t.log, "primeWindowData")
	log.Trace().
		Msg("started")
	err := t.selectedTaskBinding.Set("select a task -->")
	if err != nil {
		log.Err(err).
			Msg("error setting selected task binding to none")
	}
	// Load the running task
	runningTS := appstate.GetRunningTimesheet()
	if runningTS != nil {
		// Task is running
		t.btnStopTask.Enable()
		t.selectedTask = models.NewTaskWithData(runningTS.Task)
		newSelectedTask := runningTS.Task.String()
		if newSelectedTask != "" {
			err = t.selectedTaskBinding.Set(newSelectedTask)
			if err != nil {
				log.Err(err).
					Str("newValue", newSelectedTask).
					Msg("error setting selected task binding")
			}
			t.btnStartTask.Disable()
		} else {
			t.btnStartTask.Enable()
		}
		// Start elapsed time counter
		go t.elapsedTimeLoop(runningTS.StartTime, t.elapsedTimeQuitChan)
	} else {
		// Task is not running
		t.btnStopTask.Disable()
		t.selectedTask = nil
	}
	log.Trace().
		Msg("done")
}

func (t *timetrackerWindowData) selectedTaskChanged() {
	log := logger.GetFuncLogger(t.log, "selectedTaskChanged")
	selectedTask, err := t.selectedTaskBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("error getting selected task string from binding")
		return
	}
	t.textSelectedTask.Text = selectedTask
	t.textSelectedTask.Refresh()
}

func (t *timetrackerWindowData) setNoRunningTimesheet() {
	log := logger.GetFuncLogger(t.log, "setNoRunningTimesheet")
	// No task is running
	err := t.bindRunningTask.Set("No task is running") // i18n
	if err != nil {
		log.Err(err).
			Msg("error setting running task binding to none")
	}
	t.subStatusBox.Hide()
	t.btnStopTask.Disable()
	selection, err := t.selectedTaskBinding.Get()
	if err != nil {
		log.Err(err).
			Msg("error getting selection from binding")
	}
	if selection != "" {
		// A task is selected
		t.btnStartTask.Enable()
	} else {
		// No task is selected
		t.btnStartTask.Disable()
	}
	// Stop the elapsed time counter if it's running
	if t.elapsedTimeRunning {
		t.elapsedTimeQuitChan <- true
	}
}

func (t *timetrackerWindowData) runningTimesheetChanged(item interface{}) {
	log := logger.GetFuncLogger(t.log, "runningTimesheetChanged")
	runningTS, ok := item.(*models.TimesheetData)
	if ok {
		if runningTS == nil {
			t.setNoRunningTimesheet()
			return
		}
		// A task is running
		t.btnStopTask.Enable()
		t.btnStartTask.Disable()
		t.selectedTask = models.NewTaskWithData(runningTS.Task)
		runningTaskString := fmt.Sprintf(
			"Running task: %s", // i18n
			utils.TrimWithEllipsis(runningTS.Task.String(), taskNameTrimLength),
		)
		err := t.bindRunningTask.Set(runningTaskString)
		if err != nil {
			log.Err(err).
				Str("newTask", runningTS.Task.String()).
				Msg("error setting running task")
		}
		startTimeDisplay := runningTS.StartTime.Format(time.RFC1123Z)
		err = t.bindStartTime.Set(startTimeDisplay)
		if err != nil {
			log.Err(err).
				Str("startTime", startTimeDisplay).
				Msg("error setting start time")
		}
		elapsedTimeDisplay := time.Since(runningTS.StartTime).Truncate(time.Second).String()
		err = t.bindElapsedTime.Set(elapsedTimeDisplay)
		if err != nil {
			log.Err(err).
				Str("elapsedTime", elapsedTimeDisplay).
				Msg("error setting elapsed time")
		}
		// Start the elapsed time counter
		go t.elapsedTimeLoop(runningTS.StartTime, t.elapsedTimeQuitChan)
		t.subStatusBox.Show()
	}
}

func (t *timetrackerWindowData) doCreateAndStartTask() {
	t.createNewTaskAndStartDialog.HideCloseWindowCheckbox()
	t.createNewTaskAndStartDialog.Show()
}

func (t *timetrackerWindowData) doStartTask() {
	log := logger.GetFuncLogger(t.log, "doStartTask")
	log.Trace().
		Msg("started")
	if t.selectedTask == nil {
		log.Error().
			Msg("no task was selected")
		dialog.NewError(
			fmt.Errorf("please select a task to start"), // i18n
			t.Window,
		).Show()
		return
	}
	if t.selectedTask.Data() == nil {
		log.Error().
			Msg("selectedTask.Data() is nil; this is unexpected")
		dialog.NewError(
			errors.New("the selected task is invalid"), // i18n
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
	notificationTitle := fmt.Sprintf("Task %s started", t.selectedTask.Data().Synopsis)                 // i18n
	notificationContents := fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp)) // i18n
	(*t.app).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	t.btnStopTask.Enable()
	t.btnStartTask.Disable()
	appstate.SetRunningTimesheet(timesheet.Data())
	log.Trace().
		Msg("done")
}

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
	// Show notification that task has stopped
	notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)                     // i18n
	notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)) // i18n
	(*t.app).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	// Update the appstate
	appstate.SetRunningTimesheet(nil)
}

func (t *timetrackerWindowData) doManageTasksV2() {
	t.mngWindowV2.Show()
}

func (t *timetrackerWindowData) doSelectTask() {
	t.taskSelector.Reset()
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
	err := t.selectedTaskBinding.Set(selectedTask.String())
	if err != nil {
		log.Err(err).
			Str("newValue", selectedTask.String()).
			Msg("error setting selected task binding")
	}
	t.selectedTask = selectedTask
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
		"About Timetracker", // i18n
		fmt.Sprintf("Timetracker %s\n\nhttps://github.com/neflyte/timetracker", appVersion),
		t.Window,
	).Show()
}

// Show shows the main window
func (t *timetrackerWindowData) Show() {
	t.Window.Show()
}

// ShowAndStopRunningTask shows the main window and asks the user if they want to stop the running task
func (t *timetrackerWindowData) ShowAndStopRunningTask() {
	openTimesheets, searchErr := models.NewTimesheet().SearchOpen()
	if searchErr != nil {
		t.ShowWithError(searchErr) // i18n
		return
	}
	if len(openTimesheets) == 0 {
		t.ShowWithError(fmt.Errorf("a task is not running; please start a task first")) // i18n
		return
	}
	t.Show()
	dialogs.NewStopTaskDialog(openTimesheets[0].Task, (*t.app).Preferences(), t.maybeStopRunningTask, t.Window).Show()
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
	if err != nil {
		log.Err(err).
			Msg("error stopping the running task")
		dialog.NewError(err, t.Window).Show()
	}
	// Show notification that the task has stopped
	notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)                     // i18n
	notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)) // i18n
	(*t.app).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	appstate.SetRunningTimesheet(nil)
	// Check if we should close the main window
	shouldCloseMainWindow := (*t.app).Preferences().BoolWithFallback(prefKeyCloseWindowStopTask, false)
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
	if stoppedTimesheet != nil {
		log.Debug().
			Str("task", stoppedTimesheet.String()).
			Msg("stopped running task")
		// Show notification that task has stopped
		notificationTitle := fmt.Sprintf("Task %s stopped", stoppedTimesheet.Task.Synopsis)                     // i18n
		notificationContents := fmt.Sprintf("Stopped at %s", stoppedTimesheet.StopTime.Time.Format(time.Stamp)) // i18n
		(*t.app).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	}
	// Start the new task
	timesheet := models.NewTimesheet()
	timesheet.Data().Task = *taskData
	timesheet.Data().StartTime = time.Now()
	err = timesheet.Create()
	if err != nil {
		log.Err(err).
			Msg("error stopping current task")
		return
	}
	// Show notification that task has started
	notificationTitle := fmt.Sprintf("Task %s started", taskData.Synopsis)                              // i18n
	notificationContents := fmt.Sprintf("Started at %s", timesheet.Data().StartTime.Format(time.Stamp)) // i18n
	(*t.app).SendNotification(fyne.NewNotification(notificationTitle, notificationContents))
	log.Debug().
		Str("task", taskData.String()).
		Str("startTime", timesheet.Data().StartTime.String()).
		Msg("task started")
	appstate.SetRunningTimesheet(timesheet.Data())
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
	defer func() {
		err := t.bindElapsedTime.Set("")
		if err != nil {
			log.Err(err).
				Msg("error setting elapsed time display to empty")
		}
	}()
	log.Debug().
		Msg("loop starting")
	defer log.Debug().Msg("loop ending")
	for {
		select {
		case <-t.elapsedTimeTicker.C:
			elapsedTime := time.Since(startTime).Truncate(time.Second).String()
			err := t.bindElapsedTime.Set(elapsedTime)
			if err != nil {
				log.Err(err).
					Str("elapsedTime", elapsedTime).
					Msg("error setting elapsed time binding")
			}
		case <-quitChan:
			return
		}
	}
}
