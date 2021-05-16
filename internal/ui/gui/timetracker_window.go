package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/neflyte/timetracker/internal/ui/icons"
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
	ShowWithError(err error)
	Hide()
	Close()
	Get() ttWindow
}

type ttWindow struct {
	App            *fyne.App
	Window         fyne.Window
	Container      *fyne.Container
	TaskList       *widgets.Tasklist
	BtnStartTask   *widget.Button
	BtnStopTask    *widget.Button
	BtnManageTasks *widget.Button
	Log            zerolog.Logger
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
	t.BtnStartTask = widget.NewButtonWithIcon("START", icons.ResourcePlayCircleOutlineWhitePng, t.doStartTask)
	t.BtnStartTask.Disable() // Disable the start button by default
	t.BtnStopTask = widget.NewButtonWithIcon("STOP", icons.ResourceStopCircleOutlineWhitePng, t.doStopTask)
	t.BtnManageTasks = widget.NewButtonWithIcon("MANAGE", icons.ResourceDotsHorizontalCircleOutlineWhitePng, t.doManageTasks)
	t.Container = container.NewVBox(
		t.TaskList,
		t.BtnStartTask,
		t.BtnStopTask,
		t.BtnManageTasks,
	)
	t.Window.SetContent(t.Container)
	t.Window.SetFixedSize(true)
	t.Window.Resize(MinimumWindowSize)
	// Make sure we hide the window instead of closing it, otherwise the app will exit
	t.Window.SetCloseIntercept(func() {
		t.Window.Hide()
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
			dialog.NewError(err, t.Window).Show()
			return
		}
		taskData := new(models.TaskData)
		taskData.ID = uint(taskIdInt)
		err = models.Task(taskData).Load(false)
		if err != nil {
			dialog.NewError(err, t.Window).Show()
			return
		}
		timesheetData := new(models.TimesheetData)
		timesheetData.Task = *taskData
		timesheetData.StartTime = time.Now()
		err = models.Timesheet(timesheetData).Create()
		if err != nil {
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
	// TODO: Implementation...
}

func (t *ttWindow) doManageTasks() {
	// TODO: Show manage tasks modal
}

func (t *ttWindow) Show() {
	t.Window.Show()
}

func (t *ttWindow) ShowWithError(err error) {
	t.Show()
	dialog.NewError(err, t.Window).Show()
}

func (t *ttWindow) Hide() {
	t.Window.Hide()
}

func (t *ttWindow) Close() {
	t.Window.Close()
}

func (t *ttWindow) Get() ttWindow {
	return *t
}
