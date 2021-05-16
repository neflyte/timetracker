package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

func ShowTimetrackerWindowWithError(err error) {
	if !appstate.GetGUIStarted() {
		return
	}
	ttWin.Show()
}

func guiFunc(app *fyne.App) {
	log := logger.GetLogger("guiFunc")
	appstate.SetGUIStarted(true)
	defer appstate.SetGUIStarted(false)
	if app != nil {
		log.Trace().Msg("calling app.Run()")
		(*app).Run()
		log.Trace().Msg("fyne exited")
	}
	log.Trace().Msg("done")
}
