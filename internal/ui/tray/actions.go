package tray

import (
	"time"

	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
)

var (
	// updateTSMutex     = sync.Mutex{}
	actionLoopRunning = false
)

// actionLoop is a goroutine that periodically updates the running timesheet object in the appstate map
func actionLoop(quitChannel chan bool) {
	log := logger.GetFuncLogger(trayLogger, "actionLoop")
	if actionLoopRunning {
		log.Warn().Msg("action loop is already running")
		return
	}
	log.Trace().Msg("starting actionLoop")
	actionLoopRunning = true
	defer func() {
		actionLoopRunning = false
	}()
	for {
		updateRunningTimesheet()
		log.Trace().Msgf("delaying %d seconds until next action loop", constants.ActionLoopDelaySeconds)
		select {
		case <-quitChannel:
			log.Trace().Msg("quit channel fired; exiting function")
			return
		case <-time.After(constants.ActionLoopDelaySeconds * time.Second):
			break
		}
	}
}

// updateRunningTimesheet gets the latest running timesheet object and sets the appropriate status
func updateRunningTimesheet() {
	log := logger.GetFuncLogger(trayLogger, "updateRunningTimesheet")
	// log.Trace().Msg("acquiring lock")
	// updateTSMutex.Lock()
	// log.Trace().Msg("lock acquired successfully")
	// defer func() {
	// 	log.Trace().Msg("releasing lock")
	// 	updateTSMutex.Unlock()
	// 	log.Trace().Msg("lock released successfully")
	// }()
	timesheets, err := models.NewTimesheet().SearchOpen()
	if err != nil {
		// appstate.SetRunningTimesheet(nil) // Reset running timesheet
		runningTimesheet = nil // Reset running timesheet
		log.Err(err).Msg("error getting running timesheet")
		// appstate.SetLastError(err)
		lastError = err
		// appstate.SetLastState(constants.TimesheetStatusError)
		lastState = constants.TimesheetStatusError
	} else {
		// appstate.SetLastError(nil)
		lastError = nil
		if len(timesheets) == 0 {
			// No running task
			log.Trace().Msg("there are no running tasks")
			// appstate.SetRunningTimesheet(nil) // Reset running timesheet
			runningTimesheet = nil // Reset running timesheet
			// appstate.SetLastState(constants.TimesheetStatusIdle)
			lastState = constants.TimesheetStatusIdle
		} else {
			// Running task...
			log.Trace().Msgf("there are %d running tasks", len(timesheets))
			// appstate.SetRunningTimesheet(&timesheets[0])
			runningTimesheet = &timesheets[0]
			// appstate.SetLastState(constants.TimesheetStatusRunning)
			lastState = constants.TimesheetStatusRunning
		}
	}
	updateStatus(runningTimesheet)
}
