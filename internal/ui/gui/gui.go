package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
)

var (
	guiInitialized = false
	FyneApp        fyne.App
	mainWindow     TimetrackerWindow
	guiLogger      = logger.GetPackageLogger("gui")
)

func StartGUI() {
	if appstate.GetGUIStarted() {
		return
	}
	guiFunc(initGUI())
}

func StopGUI() {
	if !appstate.GetGUIStarted() {
		return
	}
	FyneApp.Quit()
}

func initGUI() *fyne.App {
	log := logger.GetFuncLogger(guiLogger, "initGUI")
	if guiInitialized {
		log.Debug().Msg("GUI already initialized")
		return nil
	}
	// Set up fyne
	log.Debug().Msg("setting up FyneApp")
	FyneApp = app.NewWithID("Timetracker")
	// Create the main timetracker window
	log.Debug().Msg("creating timetracker window")
	mainWindow = NewTimetrackerWindow(FyneApp)
	log.Debug().Msg("set mainWindow as master")
	mainWindow.Get().Window.SetMaster()
	log.Debug().Msg("GUI initialized")
	guiInitialized = true
	return &FyneApp
}

func ShowTimetrackerWindow() {
	if !appstate.GetGUIStarted() {
		return
	}
	mainWindow.Show()
}

func ShowTimetrackerWindowWithAbout() {
	if !appstate.GetGUIStarted() {
		return
	}
	mainWindow.ShowAbout()
}

func ShowTimetrackerWindowWithManageWindow() {
	if !appstate.GetGUIStarted() {
		return
	}
	mainWindow.ShowWithManageWindow()
}

func ShowTimetrackerWindowWithError(err error) {
	if !appstate.GetGUIStarted() {
		return
	}
	mainWindow.ShowWithError(err)
}

func ShowTimetrackerWindowWithConfirm(title string, message string, cb func(bool), hideAfterConfirm bool) {
	if !appstate.GetGUIStarted() {
		return
	}
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
}

func guiFunc(appPtr *fyne.App) {
	log := logger.GetFuncLogger(guiLogger, "guiFunc")
	if appPtr != nil {
		fyneApp := *appPtr
		appstate.SetGUIStarted(true)
		defer appstate.SetGUIStarted(false)
		log.Trace().Msg("calling app.Run()")
		fyneApp.Run()
		log.Trace().Msg("fyne exited")
	} else {
		log.Error().Msg("appPtr was nil; this is unexpected")
	}
	log.Trace().Msg("done")
}
