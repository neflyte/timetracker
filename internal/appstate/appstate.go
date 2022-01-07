package appstate

import (
	"fmt"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/reactivex/rxgo/v2"
	"sync"
)

const (
	// KeyAppVersion is the map key for the application version
	KeyAppVersion = "app_version"
	// KeyRunningTimesheet is the map key for the running timesheet, if any
	KeyRunningTimesheet = "running_timesheet"

	// keyLastState is the map key for the last timesheet state
	keyLastState = "last_state"
	// keylastError is the map key for the last ActionLoop error
	keyLastError = "last_error"

	channelBufferSize = 5
)

var (
	chanRunningTimesheet = make(chan rxgo.Item, channelBufferSize)
	observablesMap       = map[string]rxgo.Observable{
		KeyRunningTimesheet: rxgo.FromEventSource(chanRunningTimesheet),
	}
	appstateLog = logger.GetPackageLogger("appstate")
)

// syncMap is a synchronized map[interface{}]interface{} which holds the application state
var syncMap = sync.Map{}

// Map allows direct access to the synchronized map
func Map() *sync.Map {
	return &syncMap
}

// Observables returns a map of the available observables
func Observables() map[string]rxgo.Observable {
	return observablesMap
}

// GetLastState returns the last timesheet load state
func GetLastState() int {
	log := appstateLog.With().
		Str("func", "GetLastState").
		Str("key", keyLastState).
		Logger()
	lstateIntf, ok := syncMap.LoadOrStore(keyLastState, constants.TimesheetStatusIdle)
	if !ok || lstateIntf == nil {
		log.Trace().Msg("key not found; storing + loading default")
		return constants.TimesheetStatusIdle
	}
	log.Trace().Msgf("loading %#v", lstateIntf)
	return lstateIntf.(int)
}

// setLastState sets the last timesheet load state
func setLastState(newLastState int) {
	log := appstateLog.With().
		Str("func", "setLastState").
		Str("key", keyLastState).
		Logger()
	log.Trace().Msgf("storing %#v", newLastState)
	syncMap.Store(keyLastState, newLastState)
}

// GetLastError returns the last timesheet status error
func GetLastError() error {
	log := appstateLog.With().
		Str("func", "GetLastError").
		Str("key", keyLastError).
		Logger()
	lastErr, ok := syncMap.LoadOrStore(keyLastError, nil)
	if !ok {
		log.Trace().Msg("key not found; storing + loading default")
		return nil
	}
	lastErrString := "(nil)"
	if lastErr != nil {
		lastErrString = fmt.Sprintf("%#v", lastErr)
	}
	log.Trace().Msgf("loading %s", lastErrString)
	return lastErr.(error)
}

// setLastError sets the last timesheet status error
func setLastError(newLastError error) {
	log := appstateLog.With().
		Str("func", "setLastError").
		Str("key", keyLastError).
		Logger()
	lastErrString := "(nil)"
	if newLastError != nil {
		lastErrString = newLastError.Error()
	}
	log.Trace().Msgf("storing %s", lastErrString)
	syncMap.Store(keyLastState, newLastError)
}

// GetRunningTimesheet gets the running timesheet object
func GetRunningTimesheet() *models.TimesheetData {
	log := appstateLog.With().
		Str("func", "GetRunningTimesheet").
		Str("key", KeyRunningTimesheet).
		Logger()
	tsd, ok := syncMap.LoadOrStore(KeyRunningTimesheet, nil)
	if !ok {
		log.Trace().Msg("key not found; storing + loading default")
		return nil
	}
	if tsd == nil {
		log.Trace().Msg("loading nil")
		return nil
	}
	log.Trace().Msgf("loading %#v", tsd)
	return tsd.(*models.TimesheetData)
}

// SetRunningTimesheet sets the timesheet object
func SetRunningTimesheet(newTimesheet *models.TimesheetData) {
	log := appstateLog.With().
		Str("func", "setRunningTimesheet").
		Str("key", KeyRunningTimesheet).
		Logger()
	if newTimesheet == nil {
		log.Trace().Msg("storing nil")
	} else {
		log.Trace().Msgf("storing %#v", newTimesheet)
	}
	syncMap.Store(KeyRunningTimesheet, newTimesheet)
	chanRunningTimesheet <- rxgo.Of(newTimesheet)
}
