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
	App                  *fyne.App
	Window               fyne.Window
	Container            *fyne.Container
	TaskList             *widgets.Tasklist
	BtnStartTask         *widget.Button
	BtnStopTask          *widget.Button
	BtnManageTasks       *widget.Button
	EventLoopStartedChan chan bool
	EventLoopQuitChan    chan bool
	Log                  zerolog.Logger
}

func NewTimetrackerWindow(app fyne.App) TTWindow {
	ttw := &ttWindow{
		App:               &app,
		Window:            app.NewWindow("Timetracker"),
		EventLoopQuitChan: make(chan bool),
		Log:               logger.GetStructLogger("TTWindow"),
	}
	ttw.Init()
	return ttw
}

func (t *ttWindow) Init() {
	t.TaskList = widgets.NewTasklist(func(s string) {
		appstate.SetSelectedTask(s)
		if s == "" {
			t.BtnStartTask.Disable()
		} else {
			t.BtnStartTask.Enable()
		}
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
	t.Window.Resize(fyne.NewSize(MinimumWindowWidth, MinimumWindowHeight))
	// Make sure we hide the window instead of closing it, otherwise the app will exit
	t.Window.SetCloseIntercept(func() {
		t.Window.Hide()
	})
	// Spawn a goroutine to load the window's data
	go t.InitWindowData()
}

func (t *ttWindow) InitWindowData() {
	log := t.Log.With().Str("func", "InitWindowData").Logger()
	log.Trace().Msg("started")
	// Load the running task
	runningTS := appstate.GetRunningTimesheet()
	if runningTS != nil {
		newSelectedTask := (*runningTS).Task.String()
		if newSelectedTask != "" {
			t.TaskList.SetSelected(newSelectedTask)
			log.Debug().Msgf("ttwSelectedTask=%s", newSelectedTask)
		}
	}
	log.Debug().Msg("done")
}

func (t *ttWindow) startEventLoop() {
	log := t.Log.With().Str("func", "startEventLoop").Logger()
	if appstate.GetTTWindowEventLoopRunning() {
		log.Warn().Msg("eventLoop is already running")
		return
	}
	log.Debug().Msg("starting eventLoop")
	t.EventLoopStartedChan = make(chan bool)
	t.EventLoopQuitChan = make(chan bool)
	go t.eventLoop()
	log.Debug().Msg("waiting for loop to start")
	<-t.EventLoopStartedChan
	log.Debug().Msg("eventLoop started")
}

func (t *ttWindow) stopEventLoop() {
	log := t.Log.With().Str("func", "stopEventLoop").Logger()
	if !appstate.GetTTWindowEventLoopRunning() {
		log.Warn().Msg("eventLoop is not running")
		return
	}
	log.Debug().Msg("stopping eventLoop")
	t.EventLoopQuitChan <- true
}

func (t *ttWindow) eventLoop() {
	log := t.Log.With().Str("func", "eventLoop").Logger()
	t.EventLoopStartedChan <- true
	appstate.SetTTWindowEventLoopRunning(true)
	defer appstate.SetTTWindowEventLoopRunning(false)
	log.Debug().Msg("getting observables")
	chanRunningTimesheet := appstate.ObsRunningTimesheet.Observe()
	log.Trace().Msg("starting event loop")
	for {
		select {
		case runningTSItem := <-chanRunningTimesheet:
			runningTS := runningTSItem.V.(*models.TimesheetData)
			if runningTS == nil {
				log.Debug().Msg("runningTimesheet is NIL")
				t.BtnStopTask.Disable()
			} else {
				log.Debug().Msgf("runningTS=%s", (*runningTS).Task.String())
				t.BtnStopTask.Enable()
			}
			break
		case <-t.EventLoopQuitChan:
			log.Debug().Msg("received quit signal; ending event loop")
			appstate.SetTTWindowEventLoopRunning(false)
			return
		}
	}
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
	t.startEventLoop()
}

func (t *ttWindow) ShowWithError(err error) {
	t.Show()
	dialog.NewError(err, t.Window).Show()
}

func (t *ttWindow) Hide() {
	t.stopEventLoop()
	t.Window.Hide()
}

func (t *ttWindow) Close() {
	t.stopEventLoop()
	t.Window.Close()
}

func (t *ttWindow) Get() ttWindow {
	return *t
}
