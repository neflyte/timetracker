package appstate

import (
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/reactivex/rxgo/v2"
	"sync"
)

const (
	KeyAppVersion        = "app_version"
	KeyActionLoopStarted = "action_loop_started"
	KeyStatusError       = "status_error"
	KeyLastState         = "last_state"
	KeyRunningTimesheet  = "running_timesheet"
	KeySelectedTask      = "selected_task"
	KeyGUIStarted        = "gui_started"

	ChannelBufferSize = 5
)

var (
	chanRunningTimesheet = make(chan rxgo.Item, ChannelBufferSize)
	ObsRunningTimesheet  = rxgo.FromEventSource(chanRunningTimesheet)
	chanSelectedTask     = make(chan rxgo.Item, ChannelBufferSize)
	ObsSelectedTask      = rxgo.FromEventSource(chanSelectedTask)
	appstateLog          = logger.GetPackageLogger("appstate")
)

// syncMap is a synchronized map[interface{}]interface{} which holds the application state
var syncMap = sync.Map{}

// Map allows direct access to the synchronized map
func Map() *sync.Map {
	return &syncMap
}

func GetStatusError() error {
	log := appstateLog.With().
		Str("func", "GetStatusError").
		Str("key", KeyStatusError).
		Logger()
	err, ok := syncMap.LoadOrStore(KeyStatusError, nil)
	if !ok {
		log.Trace().Msg("key not found; storing + loading default")
		return nil
	}
	log.Trace().Msgf("loaded %#v", err)
	return err.(error)
}

func SetStatusError(newStatusError error) {
	log := appstateLog.With().
		Str("func", "SetStatusError").
		Str("key", KeyStatusError).
		Logger()
	log.Trace().Msgf("storing %#v", newStatusError)
	syncMap.Store(KeyStatusError, newStatusError)
}

func GetLastState() int {
	log := appstateLog.With().
		Str("func", "GetLastState").
		Str("key", KeyLastState).
		Logger()
	lstate, ok := syncMap.LoadOrStore(KeyLastState, constants.TimesheetStatusIdle)
	if !ok {
		log.Trace().Msg("key not found; storing + loading default")
		return constants.TimesheetStatusIdle
	}
	log.Trace().Msgf("loading %#v", lstate)
	return lstate.(int)
}

func SetLastState(newLastState int) {
	log := appstateLog.With().
		Str("func", "SetLastState").
		Str("key", KeyLastState).
		Logger()
	log.Trace().Msgf("storing %#v", newLastState)
	syncMap.Store(KeyLastState, newLastState)
}

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

func SetRunningTimesheet(newTimesheet *models.TimesheetData) {
	log := appstateLog.With().
		Str("func", "SetRunningTimesheet").
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

func GetSelectedTask() string {
	log := appstateLog.With().
		Str("func", "GetSelectedTask").
		Str("key", KeySelectedTask).
		Logger()
	task, ok := syncMap.LoadOrStore(KeySelectedTask, "")
	if !ok {
		log.Trace().Msg("key not found; storing + loading default")
		return ""
	}
	appstateLog.Trace().Msgf("loading %#v", task)
	return task.(string)
}

func SetSelectedTask(newTask string) {
	log := appstateLog.With().
		Str("func", "SetSelectedTask").
		Str("key", KeySelectedTask).
		Logger()
	log.Trace().Msgf("storing %#v", newTask)
	syncMap.Store(KeySelectedTask, newTask)
	chanSelectedTask <- rxgo.Of(newTask)
}

func GetGUIStarted() bool {
	log := appstateLog.With().
		Str("func", "GetGUIStarted").
		Str("key", KeyGUIStarted).
		Logger()
	started, ok := syncMap.LoadOrStore(KeyGUIStarted, false)
	if !ok {
		log.Trace().Msg("key not found; storing + loading default")
		return false
	}
	log.Trace().Msgf("loading %#v", started)
	return started.(bool)
}

func SetGUIStarted(isStarted bool) {
	log := appstateLog.With().
		Str("func", "SetGUIStarted").
		Str("key", KeyGUIStarted).
		Logger()
	log.Trace().Msgf("storing %#v", isStarted)
	syncMap.Store(KeyGUIStarted, isStarted)
}
