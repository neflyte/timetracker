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
	runningIntf, _ := Map().LoadOrStore(KeyActionLoopStarted, false)
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

func UpdateRunningTimesheet() {
	log := logger.GetLogger("appstate.UpdateRunningTimesheet")
	log.Trace().Msg("acquiring lock")
	updateTSMutex.Lock()
	log.Trace().Msg("lock acquired successfully")
	defer func() {
		log.Trace().Msg("releasing lock")
		updateTSMutex.Unlock()
		log.Trace().Msg("lock released successfully")
	}()
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
			log.Trace().Msg("there are no running tasks")
			if GetLastState() != constants.TimesheetStatusIdle {
				SetRunningTimesheet(nil) // Reset running timesheet
				SetLastState(constants.TimesheetStatusIdle)
			}
		} else {
			// Running task...
			log.Trace().Msgf("there are %d running tasks", len(timesheets))
			SetRunningTimesheet(&timesheets[0])
			if GetLastState() != constants.TimesheetStatusRunning {
				SetLastState(constants.TimesheetStatusRunning)
			}
		}
	}
}
