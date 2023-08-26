package tray

import (
	"errors"
	"time"

	"github.com/neflyte/timetracker/lib/constants"
	ttErrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
)

var (
	actionLoopRunning = false
)

// actionLoop is a goroutine that periodically updates the running timesheet object in the appstate map
func actionLoop(quitChannel chan bool, startChannel chan bool) {
	log := logger.GetFuncLogger(trayLogger, "actionLoop")
	if actionLoopRunning {
		log.Warn().
			Msg("action loop is already running")
		return
	}
	log.Trace().
		Msg("actionLoop started")
	actionLoopRunning = true
	defer func() {
		actionLoopRunning = false
		log.Trace().
			Msg("actionLoop stopped")
	}()
	log.Debug().
		Msg("waiting for start channel")
	<-startChannel
	log.Debug().
		Msg("received from start channel; starting loop")
	for {
		updateRunningTimesheet()
		log.Trace().
			Msgf("delaying %d seconds until next action loop", constants.ActionLoopDelaySeconds)
		select {
		case <-quitChannel:
			log.Trace().
				Msg("quit channel fired; exiting function")
			return
		case <-time.After(constants.ActionLoopDelaySeconds * time.Second):
			break
		}
	}
}

// updateRunningTimesheet gets the latest running timesheet object and sets the appropriate status
func updateRunningTimesheet() {
	log := logger.GetFuncLogger(trayLogger, "updateRunningTimesheet")
	defer updateStatus(runningTimesheet)
	runningTS, err := models.NewTimesheet().RunningTimesheet()
	if err != nil && !errors.Is(err, ttErrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("unable to get running timesheet")
		runningTimesheet = nil
		lastState = constants.TimesheetStatusError
		lastError = err
		return
	}
	if runningTS == nil {
		log.Trace().
			Msg("there is no running timesheet")
		runningTimesheet = nil // Reset running timesheet
		lastState = constants.TimesheetStatusIdle
		lastError = nil
		return
	}
	log.Trace().
		Msg("a timesheet is running")
	runningTimesheet = runningTS.Data()
	lastState = constants.TimesheetStatusRunning
	lastError = nil
	//timesheets, err := models.NewTimesheet().SearchOpen()
	//if err != nil {
	//	runningTimesheet = nil // Reset running timesheet
	//	log.Err(err).
	//		Msg("error getting running timesheet")
	//	lastError = err
	//	lastState = constants.TimesheetStatusError
	//} else {
	//	lastError = nil
	//	if len(timesheets) == 0 {
	//		// No running task
	//		log.Trace().
	//			Msg("there are no running tasks")
	//		runningTimesheet = nil // Reset running timesheet
	//		lastState = constants.TimesheetStatusIdle
	//	} else {
	//		// Running task...
	//		log.Trace().
	//			Msgf("there are %d running tasks", len(timesheets))
	//		runningTimesheet = &timesheets[0]
	//		lastState = constants.TimesheetStatusRunning
	//	}
	//}
	//updateStatus(runningTimesheet)
}
