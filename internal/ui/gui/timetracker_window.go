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

var (
	taskNameRE = regexp.MustCompile(`^\[([0-9]+)].*$`)
)

type TTWindow interface {
	Show()
	ShowAbout()
	ShowWithError(err error)
	Hide()
	Close()
	Get() ttWindow
}

type ttWindow struct {
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

	BindRunningTask binding.ExternalString
	BindStartTime   binding.ExternalString
	BindElapsedTime binding.ExternalString

	runningTask string
	startTime   string
	elapsedTime string
}

func NewTimetrackerWindow(app fyne.App) TTWindow {
	ttw := &ttWindow{
		App:    &app,
		Window: app.NewWindow("Timetracker"),
		Log:    logger.GetStructLogger("TTWindow"),
	}
	ttw.Init()
	return ttw
}

func (t *ttWindow) Init() {
	t.TaskList = widgets.NewTasklist(func(s string) {
		appstate.SetSelectedTask(s)
	})
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
	t.runningTask = "none"
	t.BindRunningTask = binding.BindString(&t.runningTask)
	t.LblStatus = widget.NewLabelWithData(t.BindRunningTask)
	t.LblStatus.TextStyle = fyne.TextStyle{
		Monospace: true,
	}
	t.startTime = ""
	t.BindStartTime = binding.BindString(&t.startTime)
	t.LblStartTime = widget.NewLabelWithData(t.BindStartTime)
	t.elapsedTime = ""
	t.BindElapsedTime = binding.BindString(&t.elapsedTime)
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
	t.Window.SetContent(t.Container)
	t.Window.SetFixedSize(true)
	t.Window.Resize(MinimumWindowSize)
	// Make sure we hide the window instead of closing it, otherwise the app will exit
	t.Window.SetCloseIntercept(func() {
		t.Hide()
	})
	// Set up our observables
	t.setupObservables()
	// Spawn a goroutine to load the window's data
	go t.InitWindowData()
}

func (t *ttWindow) InitWindowData() {
	funcLogger := t.Log.With().Str("func", "InitWindowData").Logger()
	funcLogger.Trace().Msg("started")
	// Load the running task
	runningTS := appstate.GetRunningTimesheet()
	if runningTS != nil {
		// Task is running
		t.BtnStopTask.Enable()
		newSelectedTask := (*runningTS).Task.String()
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
	funcLogger.Debug().Msg("done")
}

func (t *ttWindow) setupObservables() {
	log := t.Log.With().Str("func", "setupObservables").Logger()
	log.Trace().Msg("ObsRunningTimesheet")
	appstate.ObsRunningTimesheet.ForEach(
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
	log.Trace().Msg("ObsSelectedTask")
	appstate.ObsSelectedTask.ForEach(
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

func (t *ttWindow) doStartTask() {
	log := t.Log.With().Str("func", "doStartTask").Logger()
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
		taskIdString := matches[1]
		taskIdInt, err := strconv.Atoi(taskIdString)
		if err != nil {
			log.Err(err).Msgf("err converting taskIdString '%s' to int", taskIdString)
			dialog.NewError(err, t.Window).Show()
			return
		}
		taskData := new(models.TaskData)
		taskData.ID = uint(taskIdInt)
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

func (t *ttWindow) doStopTask() {
	log := t.Log.With().Str("func", "doStopTask").Logger()
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

func (t *ttWindow) doManageTasks() {
	if t.mngWindow == nil {
		t.mngWindow = NewManageWindow(*t.App)
	}
	t.mngWindow.Show()
}

func (t *ttWindow) doQuit() {
	if t.App != nil {
		app := *t.App
		app.Quit()
	}
}

func (t *ttWindow) doAbout() {
	appVersion := "??"
	appVersionIntf, ok := appstate.Map().Load(appstate.KeyAppVersion)
	if ok {
		appVersionStr := appVersionIntf.(string)
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

func (t *ttWindow) Show() {
	t.Window.Show()
}

func (t *ttWindow) ShowWithError(err error) {
	t.Show()
	dialog.NewError(err, t.Window).Show()
}

func (t *ttWindow) ShowAbout() {
	t.Show()
	t.doAbout()
}

func (t *ttWindow) Hide() {
	if t.mngWindow != nil {
		t.mngWindow.Hide()
	}
	t.Window.Hide()
}

func (t *ttWindow) Close() {
	t.Window.Close()
}

func (t *ttWindow) Get() ttWindow {
	return *t
}
