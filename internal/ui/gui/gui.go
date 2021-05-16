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
	ttWin          TTWindow
)

func StartGUI() {
	if appstate.GetGUIStarted() {
		return
	}
	initGUI()
	guiFunc(&FyneApp)
}

func StopGUI() {
	if !appstate.GetGUIStarted() {
		return
	}
	FyneApp.Quit()
}

func initGUI() {
	log := logger.GetLogger("initGUI")
	if guiInitialized {
		log.Debug().Msg("GUI already initialized")
		return
	}
	// Set up fyne
	log.Trace().Msg("setting up FyneApp")
	FyneApp = app.New()
	// Create the main timetracker window
	log.Trace().Msg("creating timetracker window")
	ttWin = NewTimetrackerWindow(FyneApp)
	if ttWin != nil {
		log.Trace().Msg("set ttWin as master")
		ttWin.Get().Window.SetMaster()
	}
	log.Debug().Msg("GUI initialized")
	guiInitialized = true
}

func ShowTimetrackerWindow() {
	if !appstate.GetGUIStarted() {
		return
	}
	ttWin.Show()
}

func ShowTimetrackerWindowWithAbout() {
	if !appstate.GetGUIStarted() {
		return
	}
	ttWin.ShowAbout()
}

func ShowTimetrackerWindowWithError(err error) {
	if !appstate.GetGUIStarted() {
		return
	}
	ttWin.ShowWithError(err)
}

func ShowTimetrackerWindowWithConfirm(title string, message string, cb func(bool), hideAfterConfirm bool) {
	if !appstate.GetGUIStarted() {
		return
	}
	ttWin.Show()
	dialog.NewConfirm(
		title,
		message,
		func(res bool) {
			// Call the callback
			cb(res)
			// Hide the window if we were asked to
			if hideAfterConfirm {
				ttWin.Hide()
			}
		},
		ttWin.Get().Window,
	).Show()
}

func guiFunc(appPtr *fyne.App) {
	log := logger.GetLogger("guiFunc")
	if appPtr != nil {
		fyneApp := *appPtr
		appstate.SetGUIStarted(true)
		defer appstate.SetGUIStarted(false)
		log.Trace().Msg("calling app.Run()")
		fyneApp.Run()
		log.Trace().Msg("fyne exited")
	}
	log.Trace().Msg("done")
}
