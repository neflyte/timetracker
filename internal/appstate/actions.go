package appstate

import (
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"sync"
	"time"
)

var (
	updateTSMutex     = sync.Mutex{}
	actionLoopRunning = false
)

// ActionLoop is a goroutine that periodically updates the running timesheet object in the appstate map
func ActionLoop(quitChannel chan bool) {
	log := logger.GetFuncLogger(appstateLog, "ActionLoop")
	if actionLoopRunning {
		log.Warn().Msg("action loop is already running")
		return
	}
	log.Debug().Msg("starting ActionLoop")
	actionLoopRunning = true
	defer func() {
		actionLoopRunning = false
	}()
	for {
		updateRunningTimesheet()
		log.Trace().Msgf("delaying %d seconds until next action loop", constants.ActionLoopDelaySeconds)
		select {
		case <-quitChannel:
			log.Debug().Msg("quit channel fired; exiting function")
			return
		case <-time.After(constants.ActionLoopDelaySeconds * time.Second):
			break
		}
	}
}

// updateRunningTimesheet gets the latest running timesheet object and sets the appropriate status
func updateRunningTimesheet() {
	log := logger.GetFuncLogger(appstateLog, "updateRunningTimesheet")
	updateTSMutex.Lock()
	log.Trace().Msg("lock acquired successfully")
	defer func() {
		log.Trace().Msg("releasing lock")
		updateTSMutex.Unlock()
		log.Trace().Msg("lock released successfully")
	}()
	timesheets, err := models.NewTimesheet().SearchOpen()
	if err != nil {
		SetRunningTimesheet(nil) // Reset running timesheet
		log.Err(err).Msg("error getting running timesheet")
		setLastError(err)
		if GetLastState() != constants.TimesheetStatusError {
			setLastState(constants.TimesheetStatusError)
		}
	} else {
		setLastError(nil)
		if len(timesheets) == 0 {
			// No running task
			log.Trace().Msg("there are no running tasks")
			if GetLastState() != constants.TimesheetStatusIdle {
				SetRunningTimesheet(nil) // Reset running timesheet
				setLastState(constants.TimesheetStatusIdle)
			}
		} else {
			// Running task...
			log.Trace().Msgf("there are %d running tasks", len(timesheets))
			SetRunningTimesheet(&timesheets[0])
			if GetLastState() != constants.TimesheetStatusRunning {
				setLastState(constants.TimesheetStatusRunning)
			}
		}
	}
}
