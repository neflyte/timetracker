package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
)

var (
	// FyneApp is the main fyne app instance
	FyneApp fyne.App

	guiInitialized = false
	mainWindow     TimetrackerWindow
	guiLogger      = logger.GetPackageLogger("gui")
)

func StartGUI(app *fyne.App) {
	if appstate.GetGUIStarted() {
		return
	}
	guiFunc(app)
}

func StopGUI() {
	if !appstate.GetGUIStarted() {
		return
	}
	FyneApp.Quit()
}

func InitGUI() *fyne.App {
	log := logger.GetFuncLogger(guiLogger, "initGUI")
	if guiInitialized {
		log.Warn().Msg("GUI already initialized")
		return nil
	}
	// Set up fyne
	FyneApp = app.NewWithID("Timetracker")
	// Create the main timetracker window
	mainWindow = NewTimetrackerWindow(FyneApp)
	mainWindow.Get().Window.SetMaster()
	guiInitialized = true
	return &FyneApp
}

func ShowTimetrackerWindow() {
	mainWindow.Show()
}

func ShowTimetrackerWindowWithAbout() {
	mainWindow.ShowAbout()
}

func ShowTimetrackerWindowWithManageWindow() {
	mainWindow.ShowWithManageWindow()
}

func ShowTimetrackerWindowAndStopRunningTask() {
	mainWindow.ShowAndStopRunningTask()
}

/*func ShowTimetrackerWindowWithError(err error) {
	mainWindow.ShowWithError(err)
}

func ShowTimetrackerWindowWithConfirm(title string, message string, cb func(bool), hideAfterConfirm bool) {
	mainWindow.Show()
	dialog.NewConfirm(
		title,
		message,
		func(res bool) {
			// Call the callback
			cb(res)
			// Hide the window if we were asked to
			if hideAfterConfirm {
				mainWindow.Hide()
			}
		},
		mainWindow.Get().Window,
	).Show()
}*/

func guiFunc(appPtr *fyne.App) {
	log := logger.GetFuncLogger(guiLogger, "guiFunc")
	if appPtr != nil {
		fyneApp := *appPtr
		// Start ActionLoop
		actionLoopQuitChan := make(chan bool, 1)
		go appstate.ActionLoop(actionLoopQuitChan)
		// Set gui started state
		appstate.SetGUIStarted(true)
		defer appstate.SetGUIStarted(false)
		// start Fyne
		log.Trace().Msg("calling fyneApp.Run()")
		fyneApp.Run()
		log.Trace().Msg("fyne exited")
		// stop actionloop
		actionLoopQuitChan <- true
	} else {
		log.Error().Msg("appPtr was nil; this is unexpected")
	}
	log.Trace().Msg("done")
}
