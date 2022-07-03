package gui

import (
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/icons"
)

const (
	statusStartTaskTitle = "Start new task"
	// statusStartTaskDescription = "Display a task selector and start a task"
	// statusStopTaskDescription  = "Stop the running task"

	recentlyStartedTasks = 5
)

var (
	trayMenu                   *fyne.Menu
	trayMenuItems              []*fyne.MenuItem
	mStatus                    *fyne.MenuItem
	mManage                    *fyne.MenuItem
	mCreateAndStart            *fyne.MenuItem
	mTrayOptions               *fyne.MenuItem
	mTrayOptionConfirmStopTask *fyne.MenuItem
	mLastStarted               *fyne.MenuItem
	lastStartedItems           [recentlyStartedTasks]*fyne.MenuItem
	lastStartedItemSynopses    [recentlyStartedTasks]string
	mAbout                     *fyne.MenuItem
	desktopApp                 *desktop.App
)

// trayInit initializes the system tray and its menus
func trayInit(app *fyne.App) error {
	if app == nil {
		return errors.New("unexpected nil app")
	}
	dt, ok := (*app).(desktop.App)
	if !ok {
		return errors.New("app is not running in desktop mode; cannot initialize tray")
	}
	if desktopApp == nil {
		desktopApp = &dt
	}
	if trayMenuItems == nil {
		trayMenuItems = createTrayMenuItems()
	}
	if trayMenu == nil {
		trayMenu = fyne.NewMenu("Timetracker", trayMenuItems...)
	}
	dt.SetSystemTrayMenu(trayMenu)
	dt.SetSystemTrayIcon(icons.CheckIcon())
	return nil
}

func createTrayMenuItems() []*fyne.MenuItem {
	if mStatus == nil {
		mStatus = fyne.NewMenuItem(statusStartTaskTitle, func() {})
	}
	if mCreateAndStart == nil {
		mCreateAndStart = fyne.NewMenuItem("Create and Start new task", func() {})
	}
	if mManage == nil {
		mManage = fyne.NewMenuItem("Manage tasks", func() {})
	}
	if mTrayOptions == nil {
		mTrayOptions = fyne.NewMenuItem("Tray options", func() {})
		if mTrayOptionConfirmStopTask == nil {
			mTrayOptionConfirmStopTask = fyne.NewMenuItem("Confirm when stopping a task", func() {})
		}
		mTrayOptions.ChildMenu = fyne.NewMenu("", mTrayOptionConfirmStopTask)
	}
	if mLastStarted == nil {
		mLastStarted = fyne.NewMenuItem("Recent tasks", func() {})
		for x := 0; x < recentlyStartedTasks; x++ {
			lastStartedItems[x] = fyne.NewMenuItem("--", func() {})
			lastStartedItemSynopses[x] = ""
		}
		mLastStarted.ChildMenu = fyne.NewMenu("", lastStartedItems[:]...)
	}
	if mAbout == nil {
		mAbout = fyne.NewMenuItem("About Timetracker", func() {})
	}
	return []*fyne.MenuItem{
		mStatus,
		mCreateAndStart,
		mManage,
		fyne.NewMenuItemSeparator(),
		mTrayOptions,
		fyne.NewMenuItemSeparator(),
		mLastStarted,
		fyne.NewMenuItemSeparator(),
		mAbout,
	}
}

func TrayUpdate(tsd *models.TimesheetData) {
	log := logger.GetFuncLogger(guiLogger, "TrayUpdate")
	// handle the last state
	lastState := appstate.GetLastState()
	if lastState == constants.TimesheetStatusError {
		log.Trace().Msgf("got error for lastState")
		// Get the error
		lastStateError := appstate.GetLastError()
		if lastStateError != nil {
			log.Trace().Msgf("got error: %s", lastStateError.Error())
		}
		(*desktopApp).SetSystemTrayIcon(icons.ErrorIcon())
		mStatus.Label = "Error (click for details)"
	} else {
		// Check the supplied timesheet
		if tsd == nil {
			// No running timesheet
			log.Trace().Msg("got nil running timesheet item")
			(*desktopApp).SetSystemTrayIcon(icons.CheckIcon())
			mStatus.Label = statusStartTaskTitle
		} else {
			log.Trace().Msgf("got running timesheet object: %s", tsd.String())
			(*desktopApp).SetSystemTrayIcon(icons.RunningIcon())
			statusText := fmt.Sprintf(
				"Stop task %s (%s)",
				tsd.Task.Synopsis,
				time.Since(tsd.StartTime).Truncate(time.Second).String(),
			)
			mStatus.Label = statusText
		}
	}
	// Last 5 started tasks
	lastStartedTasks, err := models.NewTimesheet().LastStartedTasks(recentlyStartedTasks)
	if err != nil {
		log.Err(err).Msg("error loading recently-started tasks")
	} else {
		log.Debug().Msgf("len(lastStartedTasks)=%d", len(lastStartedTasks))
		// Hide entries we don't need
		if len(lastStartedTasks) < recentlyStartedTasks {
			log.Debug().Msgf("len(lastStartedTasks) < recentlyStartedTasks; %d < %d", len(lastStartedTasks), recentlyStartedTasks)
			for x := recentlyStartedTasks - 1; x > len(lastStartedItems)-1; x-- {
				log.Debug().Msgf("hiding item at index %d; item=%s", x, lastStartedItems[x].Label)
				lastStartedItems[x].Label = "--"
				lastStartedItems[x].Disabled = true
				lastStartedItemSynopses[x] = ""
			}
		}
		// Fill in the entries we have
		for x := 0; x < len(lastStartedTasks); x++ {
			log.Debug().Msgf("showing item at index %d; synopsis=%s", x, lastStartedTasks[x].Synopsis)
			lastStartedItems[x].Disabled = false
			lastStartedItems[x].Label = lastStartedTasks[x].Synopsis
			lastStartedItemSynopses[x] = lastStartedTasks[x].Synopsis
		}
	}
}
