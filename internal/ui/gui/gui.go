package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
)

var (
	// fyneApp is the main fyne app instance
	fyneApp fyne.App

	guiInitialized = false
	mainWindow     timetrackerWindow
	guiLogger      = logger.GetPackageLogger("gui")
	guiStarted     = false
)

// StartGUI starts the GUI app
func StartGUI(app *fyne.App) {
	if guiStarted {
		return
	}
	guiFunc(app)
}

// InitGUI initializes the GUI app
func InitGUI() *fyne.App {
	log := logger.GetFuncLogger(guiLogger, "initGUI")
	if guiInitialized {
		log.Warn().Msg("GUI already initialized")
		return nil
	}
	// Set up fyne
	fyneApp = app.NewWithID("Timetracker")
	// Create the main timetracker window
	mainWindow = newTimetrackerWindow(fyneApp)
	mainWindow.Get().Window.SetMaster()
	guiInitialized = true
	return &fyneApp
}

// ShowTimetrackerWindow shows the main timetracker window
func ShowTimetrackerWindow() {
	mainWindow.Show()
}

// ShowTimetrackerWindowWithAbout shows the main timetracker window and then shows the about dialog
func ShowTimetrackerWindowWithAbout() {
	mainWindow.ShowAbout()
}

// ShowTimetrackerWindowWithManageWindow shows the main timetracker window and then shows the manage window
func ShowTimetrackerWindowWithManageWindow() {
	mainWindow.ShowWithManageWindow()
}

// ShowTimetrackerWindowAndStopRunningTask shows the main timetracker window and then confirms if the running task should be stopped
func ShowTimetrackerWindowAndStopRunningTask() {
	mainWindow.ShowAndStopRunningTask()
}

// ShowTimetrackerWindowAndShowCreateAndStartDialog shows the main timetracker window and then shows the Create and Start dialog
func ShowTimetrackerWindowAndShowCreateAndStartDialog() {
	mainWindow.ShowAndDisplayCreateAndStartDialog()
}

func guiFunc(appPtr *fyne.App) {
	log := logger.GetFuncLogger(guiLogger, "guiFunc")
	if appPtr != nil {
		appInstance := *appPtr
		// Start ActionLoop
		actionLoopQuitChan := make(chan bool, 1)
		go appstate.ActionLoop(actionLoopQuitChan)
		// Set gui started state
		guiStarted = true
		defer func() {
			guiStarted = false
		}()
		// start Fyne
		log.Trace().Msg("calling appInstance.Run()")
		appInstance.Run()
		log.Trace().Msg("fyne exited")
		// stop actionloop
		actionLoopQuitChan <- true
	} else {
		log.Error().Msg("appPtr was nil; this is unexpected")
	}
	log.Trace().Msg("done")
}
