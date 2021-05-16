package appstate

import (
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"sync"
	"time"
)

var (
	updateTSMutex = sync.Mutex{}
)

func ActionLoop(quitChannel chan bool) {
	log := logger.GetLogger("appstate.ActionLoop")
	runningIntf, ok := Map().LoadOrStore(KeyActionLoopStarted, false)
	if !ok {
		log.Error().Msgf("error loading key %s", KeyActionLoopStarted)
		return
	}
	running := runningIntf.(bool)
	if running {
		log.Warn().Msg("action loop is already running")
		return
	}
	log.Debug().Msg("starting ActionLoop")
	Map().Store(KeyActionLoopStarted, true)
	defer Map().Store(KeyActionLoopStarted, false)
	for {
		UpdateRunningTimesheet()
		log.Debug().Msgf("delaying %d seconds until next action loop", constants.ActionLoopDelaySeconds)
		select {
		case <-quitChannel:
			log.Debug().Msg("quit channel fired; exiting function")
			return
		case <-time.After(constants.ActionLoopDelaySeconds * time.Second):
			break
		}
	}
}

func UpdateRunningTimesheet() {
	log := logger.GetLogger("appstate.UpdateRunningTimesheet")
	updateTSMutex.Lock()
	defer updateTSMutex.Unlock()
	timesheets, err := models.Timesheet(new(models.TimesheetData)).SearchOpen()
	SetStatusError(err)
	if err != nil {
		SetRunningTimesheet(nil) // Reset running timesheet
		log.Err(err).Msg("error getting running timesheet")
		if GetLastState() != constants.TimesheetStatusError {
			SetLastState(constants.TimesheetStatusError)
		}
	} else {
		if len(timesheets) == 0 {
			// No running task
			if GetLastState() != constants.TimesheetStatusIdle {
				SetRunningTimesheet(nil) // Reset running timesheet
				SetLastState(constants.TimesheetStatusIdle)
			}
		} else {
			// Running task...
			SetRunningTimesheet(&timesheets[0])
			if GetLastState() != constants.TimesheetStatusRunning {
				SetLastState(constants.TimesheetStatusRunning)
			}
		}
	}
}
