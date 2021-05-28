package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
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

type TimetrackerWindow interface {
	Show()
	ShowAbout()
	ShowWithError(err error)
	ShowWithManageWindow()
	ShowAndStopRunningTask()
	Hide()
	Close()
	Get() timetrackerWindow
}

type timetrackerWindow struct {
	App            *fyne.App
	Window         fyne.Window
	Container      *fyne.Container
	StatusBox      *fyne.Container
	SubStatusBox   *fyne.Container
	ButtonBox      *fyne.Container
	TaskList       *widgets.Tasklist
	BtnStartTask   *widget.Button
	BtnStopTask    *widget.Button
	BtnManageTasks *widget.Button
	BtnAbout       *widget.Button
	BtnQuit        *widget.Button
	Log            zerolog.Logger
	mngWindow      ManageWindow

	LblStatus      *widget.Label
	LblStartTime   *widget.Label
	LblElapsedTime *widget.Label

	BindRunningTask binding.String
	BindStartTime   binding.String
	BindElapsedTime binding.String
}

func NewTimetrackerWindow(app fyne.App) TimetrackerWindow {
	ttw := &timetrackerWindow{
		App:    &app,
		Window: app.NewWindow("Timetracker"),
		Log:    logger.GetStructLogger("TimetrackerWindow"),
	}
	ttw.Init()
	return ttw
}

func (t *timetrackerWindow) Init() {
	log := logger.GetFuncLogger(t.Log, "Init")
	log.Debug().Msg("started")
	t.TaskList = widgets.NewTasklist()
	t.TaskList.OnChanged = appstate.SetSelectedTask
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
	t.Window.Resize(MinimumWindowSize)
	// Make sure we hide the window instead of closing it, otherwise the app will exit
	t.Window.SetCloseIntercept(t.Hide)
	// Set up our observables
	t.setupObservables()
	// Load the window's data
	t.InitWindowData()
	// Set up the Manage Window as well
	t.mngWindow = NewManageWindow(*t.App)
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

func (t *timetrackerWindow) InitWindowData() {
	log := logger.GetFuncLogger(t.Log, "InitWindowData")
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

func (t *timetrackerWindow) setupObservables() {
	log := logger.GetFuncLogger(t.Log, "setupObservables")
	log.Debug().Msg("ObsRunningTimesheet")
	appstate.Observables()[appstate.KeyRunningTimesheet].ForEach(
		func(item interface{}) {
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
					if appstate.GetSelectedTask() != "" {
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
		},
		func(err error) {
			log.Err(err).Msg("error getting running timesheet from rxgo observable")
		},
		func() {
			log.Trace().Msg("running timesheet observable is done")
		},
	)
	log.Debug().Msg("ObsSelectedTask")
	appstate.Observables()[appstate.KeySelectedTask].ForEach(
		func(item interface{}) {
			selectedTask, ok := item.(string)
			if ok {
				if selectedTask != "" {
					if appstate.GetRunningTimesheet() == nil {
						t.BtnStartTask.Enable()
					} else {
						t.BtnStartTask.Disable()
					}
					return
				}
				t.BtnStartTask.Disable()
			}
		},
		func(err error) {
			log.Err(err).Msg("error getting selected task from rxgo observable")
		},
		func() {
			log.Trace().Msg("selected task observable is done")
		},
	)
}

func (t *timetrackerWindow) doStartTask() {
	log := logger.GetFuncLogger(t.Log, "doStartTask")
	log.Trace().Msg("started")
	selectedTask := appstate.GetSelectedTask()
	if selectedTask == "" {
		log.Error().Msg("no task was selected")
		dialog.NewError(
			fmt.Errorf("please select a task to start"),
			t.Window,
		).Show()
		return
	}
	// TODO: convert from selectedTask string to task ID so we can start a new timesheet
	if taskNameRE.MatchString(selectedTask) {
		matches := taskNameRE.FindStringSubmatch(selectedTask)
		taskIDString := matches[1]
		taskIDInt, err := strconv.Atoi(taskIDString)
		if err != nil {
			log.Err(err).Msgf("err converting taskIDString '%s' to int", taskIDString)
			dialog.NewError(err, t.Window).Show()
			return
		}
		taskData := new(models.TaskData)
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
		appstate.UpdateRunningTimesheet()
	}
	log.Trace().Msg("done")
}

func (t *timetrackerWindow) doStopTask() {
	log := logger.GetFuncLogger(t.Log, "doStopTask")
	log.Trace().Msg("started")
	runningTS := appstate.GetRunningTimesheet()
	if runningTS == nil {
		log.Warn().Msg("no timesheet is running")
		return
	}
	// Stop the running task
	log.Debug().Msgf("stopping task %s", runningTS.Task.Synopsis)
	err := models.Task(new(models.TaskData)).StopRunningTask()
	if err != nil {
		log.Err(err).Msg(errors.StopRunningTaskError)
		dialog.NewError(err, t.Window).Show()
	}
	// Get a new timesheet and update the appstate
	appstate.UpdateRunningTimesheet()
	log.Trace().Msg("done")
}

func (t *timetrackerWindow) doManageTasks() {
	t.mngWindow.Show()
}

func (t *timetrackerWindow) doQuit() {
	StopGUI()
}

func (t *timetrackerWindow) doAbout() {
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
		fmt.Sprintf("Timetracker v%s\n\nhttps://github.com/neflyte/timetracker", appVersion),
		t.Window,
	).Show()
}

func (t *timetrackerWindow) Show() {
	t.Window.Show()
	t.TaskList.Refresh()
}

func (t *timetrackerWindow) ShowAndStopRunningTask() {
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

func (t *timetrackerWindow) ShowWithManageWindow() {
	t.Show()
	t.doManageTasks()
}

func (t *timetrackerWindow) ShowWithError(err error) {
	t.Show()
	dialog.NewError(err, t.Window).Show()
}

func (t *timetrackerWindow) ShowAbout() {
	t.Show()
	t.doAbout()
}

func (t *timetrackerWindow) Hide() {
	if t.mngWindow != nil {
		t.mngWindow.Hide()
	}
	t.Window.Hide()
}

func (t *timetrackerWindow) Close() {
	t.Window.Close()
}

func (t *timetrackerWindow) Get() timetrackerWindow {
	return *t
}

func (t *timetrackerWindow) maybeStopRunningTask(stopTask bool) {
	log := logger.GetFuncLogger(t.Log, "maybeStopRunningTask")
	if !stopTask {
		return
	}
	td := new(models.TaskData)
	err := td.StopRunningTask()
	if err != nil {
		log.Err(err).Msg("error stopping the running task")
		dialog.NewError(err, t.Window).Show()
	}
}
