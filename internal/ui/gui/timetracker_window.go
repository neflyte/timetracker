package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"regexp"
	"strconv"
	"time"
)

// ttw is the singleton instance of the Timetracker Window
var (
	ttw        = TTWindow{}
	taskNameRE = regexp.MustCompile(`^\[([0-9]+)].*$`)
)

type TTWindow struct {
	App               *fyne.App
	Window            *fyne.Window
	Container         *fyne.Container
	TaskList          *widgets.Tasklist
	BtnStartTask      *widget.Button
	BtnStopTask       *widget.Button
	BtnManageTasks    *widget.Button
	EventLoopQuitChan chan bool
}

func TimetrackerWindow(app *fyne.App) *fyne.Window {
	if ttw.Window == nil && app != nil {
		ttw.App = app
		ttwin := (*app).NewWindow("Timetracker")
		initTimetrackerWindow(ttwin)
		go initWindowData() // Load data in a goroutine
		ttw.Window = &ttwin
	}
	return ttw.Window
}

func ShowTimetrackerWindow(app *fyne.App) {
	ttwin := TimetrackerWindow(app)
	if ttwin != nil {
		if !appstate.GetTTWindowEventLoopRunning() {
			appstate.SetTTWindowEventLoopRunning(true)
			go eventLoop(ttw.EventLoopQuitChan)
		}
		(*ttwin).Show()
	}
}

func initTimetrackerWindow(w fyne.Window) {
	ttw.TaskList = widgets.NewTasklist(func(s string) {
		appstate.SetSelectedTask(s)
		if s == "" {
			ttw.BtnStartTask.Disable()
		} else {
			ttw.BtnStartTask.Enable()
		}
	})
	ttw.BtnStartTask = widget.NewButtonWithIcon("START", icons.ResourcePlayCircleOutlineWhitePng, doStartTask)
	ttw.BtnStartTask.Disable() // Disable the start button by default
	ttw.BtnStopTask = widget.NewButtonWithIcon("STOP", icons.ResourceStopCircleOutlineWhitePng, doStopTask)
	ttw.BtnManageTasks = widget.NewButtonWithIcon("MANAGE", icons.ResourceDotsHorizontalCircleOutlineWhitePng, doManageTasks)
	ttw.Container = container.NewVBox(
		ttw.TaskList,
		ttw.BtnStartTask,
		ttw.BtnStopTask,
		ttw.BtnManageTasks,
	)
	ttw.EventLoopQuitChan = make(chan bool, 1)
	w.SetContent(ttw.Container)
	w.SetFixedSize(true)
	w.SetCloseIntercept(func() {
		ttw.EventLoopQuitChan <- true
		w.Hide()
	})
}

func initWindowData() {
	log := logger.GetLogger("initWindowData")
	log.Trace().Msg("started")
	// Load the running task
	runningTS := appstate.GetRunningTimesheet()
	if runningTS != nil {
		newSelectedTask := (*runningTS).Task.String()
		if newSelectedTask != "" {
			ttw.TaskList.SetSelected(newSelectedTask)
			log.Debug().Msgf("ttwSelectedTask=%s", newSelectedTask)
		}
	}
	log.Debug().Msg("done")
}

func eventLoop(quitChan chan bool) {
	log := logger.GetLogger("eventLoop")
	log.Debug().Msg("getting observables")
	chanRunningTimesheet := appstate.ObsRunningTimesheet.Observe()
	log.Trace().Msg("starting event loop")
	for {
		select {
		case runningTSItem := <-chanRunningTimesheet:
			runningTS := runningTSItem.V.(*models.TimesheetData)
			if runningTS == nil {
				log.Debug().Msg("runningTimesheet is NIL")
				ttw.BtnStopTask.Disable()
			} else {
				log.Debug().Msgf("runningTS=%s", (*runningTS).Task.String())
				ttw.BtnStopTask.Enable()
			}
			break
		case <-quitChan:
			log.Debug().Msg("received quit signal; ending event loop")
			appstate.SetTTWindowEventLoopRunning(false)
			return
		}
	}
}

func doStartTask() {
	log := logger.GetLogger("doStartTask")
	log.Trace().Msg("started")
	selectedTask := appstate.GetSelectedTask()
	if selectedTask == "" {
		NewErrorDialogWindow(
			*ttw.App,
			"no task selected",
			fmt.Errorf("please select a task to start"),
			nil,
			nil).Show()
		return
	}
	// TODO: convert from selectedTask string to task ID so we can start a new timesheet
	if taskNameRE.MatchString(selectedTask) {
		matches := taskNameRE.FindStringSubmatch(selectedTask)
		taskIdString := matches[1]
		taskIdInt, err := strconv.Atoi(taskIdString)
		if err != nil {
			NewErrorDialogWindow(
				*ttw.App,
				"invalid task id",
				err,
				nil,
				nil).Show()
			return
		}
		taskData := new(models.TaskData)
		taskData.ID = uint(taskIdInt)
		err = models.Task(taskData).Load(false)
		if err != nil {
			NewErrorDialogWindow(
				*ttw.App,
				"error loading task",
				err,
				nil,
				nil).Show()
			return
		}
		timesheetData := new(models.TimesheetData)
		timesheetData.Task = *taskData
		timesheetData.StartTime = time.Now()
		err = models.Timesheet(timesheetData).Create()
		if err != nil {
			NewErrorDialogWindow(
				*ttw.App,
				"error starting task",
				err,
				nil,
				nil).Show()
			return
		}
		ttw.BtnStopTask.Enable()
	}
	log.Trace().Msg("done")
}

func doStopTask() {
	// TODO: Implementation...
}

func doManageTasks() {
	// TODO: Show manage tasks modal
}
