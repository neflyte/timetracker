package gui

import (
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/ui/gui/windows"
	"github.com/neflyte/timetracker/lib/ui/icons"
)

var (
	// fyneApp is the main fyne app instance
	fyneApp fyne.App

	guiInitialized = false
	mainWindow     windows.TimetrackerWindow
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
func InitGUI(appVersion string) *fyne.App {
	log := logger.GetFuncLogger(guiLogger, "initGUI")
	if guiInitialized {
		log.Warn().
			Msg("GUI already initialized")
		return nil
	}
	// Set up fyne
	fyneApp = app.NewWithID("cc.ethereal.timetracker")
	fyneApp.SetIcon(icons.IconV2)
	// Create the main timetracker window
	mainWindow = windows.NewTimetrackerWindow(fyneApp, appVersion)
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
		defer log.Trace().
			Msg("done")
		appInstance := *appPtr
		// Start Signal catcher
		signalFuncQuitChan := make(chan bool, 1)
		go signalFunc(signalFuncQuitChan, appPtr)
		// Set gui started state
		guiStarted = true
		defer func() {
			guiStarted = false
		}()
		// start Fyne
		log.Trace().
			Msg("calling appInstance.Run()")
		appInstance.Run()
		log.Trace().
			Msg("fyne exited")
		// stop signal catcher
		signalFuncQuitChan <- true
	} else {
		log.Error().
			Msg("appPtr was nil; this is unexpected")
	}
}

func signalFunc(quitChan chan bool, appPtr *fyne.App) {
	log := logger.GetFuncLogger(guiLogger, "signalFunc")
	if appPtr != nil {
		// Create a channel to catch OS signals
		signalChan := make(chan os.Signal, 1)
		// Catch OS interrupt and SIGTERM signals
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		// Start signal loop
		log.Trace().
			Msg("starting")
		defer log.Trace().
			Msg("done")
		for {
			select {
			case <-signalChan:
				log.Warn().
					Msg("caught os.Interrupt or SIGTERM; shutting down GUI")
				(*appPtr).Quit()
				return
			case <-quitChan:
				log.Trace().
					Msg("quit channel fired; exiting function")
				return
			}
		}
	} else {
		log.Error().
			Msg("appPtr was nil; this is unexpected")
	}
}
