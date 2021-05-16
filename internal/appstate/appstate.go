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
	// chanStatusError      = make(chan rxgo.Item, ChannelBufferSize)
	// ObsStatusError       = rxgo.FromEventSource(chanStatusError)
	// chanLastState        = make(chan rxgo.Item, ChannelBufferSize)
	// ObsLastState         = rxgo.FromEventSource(chanLastState)
	chanRunningTimesheet = make(chan rxgo.Item, ChannelBufferSize)
	ObsRunningTimesheet  = rxgo.FromEventSource(chanRunningTimesheet)
	chanSelectedTask     = make(chan rxgo.Item, ChannelBufferSize)
	ObsSelectedTask      = rxgo.FromEventSource(chanSelectedTask)
	appstateLog          = logger.GetLogger("appstate")
)

// syncMap is a synchronized map[interface{}]interface{} which holds the application state
var syncMap = sync.Map{}

// Map allows direct access to the synchronized map
func Map() *sync.Map {
	return &syncMap
}

func GetStatusError() error {
	err, ok := syncMap.LoadOrStore(KeyStatusError, nil)
	if !ok {
		appstateLog.Trace().
			Str("key", KeyStatusError).
			Msg("key not found; storing + loading default")
		return nil
	}
	appstateLog.Trace().
		Str("key", KeyStatusError).
		Msgf("loaded %#v", err)
	return err.(error)
}

func SetStatusError(newStatusError error) {
	appstateLog.Trace().
		Str("key", KeyStatusError).
		Msgf("storing %#v", newStatusError)
	syncMap.Store(KeyStatusError, newStatusError)
	// chanStatusError<- rxgo.Of(newStatusError)
}

func GetLastState() int {
	lstate, ok := syncMap.LoadOrStore(KeyLastState, constants.TimesheetStatusIdle)
	if !ok {
		appstateLog.Trace().
			Str("key", KeyLastState).
			Msg("key not found; storing + loading default")
		return constants.TimesheetStatusIdle
	}
	appstateLog.Trace().
		Str("key", KeyLastState).
		Msgf("loading %#v", lstate)
	return lstate.(int)
}

func SetLastState(newLastState int) {
	appstateLog.Trace().
		Str("key", KeyLastState).
		Msgf("storing %#v", newLastState)
	syncMap.Store(KeyLastState, newLastState)
	// chanLastState<- rxgo.Of(newLastState)
}

func GetRunningTimesheet() *models.TimesheetData {
	tsd, ok := syncMap.LoadOrStore(KeyRunningTimesheet, nil)
	if !ok {
		appstateLog.Trace().
			Str("key", KeyRunningTimesheet).
			Msg("key not found; storing + loading default")
		return nil
	}
	if tsd == nil {
		appstateLog.Trace().
			Str("key", KeyRunningTimesheet).
			Msg("loading nil")
		return nil
	}
	appstateLog.Trace().
		Str("key", KeyRunningTimesheet).
		Msgf("loading %#v", tsd)
	return tsd.(*models.TimesheetData)
}

func SetRunningTimesheet(newTimesheet *models.TimesheetData) {
	if newTimesheet == nil {
		appstateLog.Trace().
			Str("key", KeyRunningTimesheet).
			Msg("storing nil")
	} else {
		appstateLog.Trace().
			Str("key", KeyRunningTimesheet).
			Msgf("storing %#v", newTimesheet)
	}
	syncMap.Store(KeyRunningTimesheet, newTimesheet)
	chanRunningTimesheet <- rxgo.Of(newTimesheet)
}

func GetSelectedTask() string {
	task, ok := syncMap.LoadOrStore(KeySelectedTask, "")
	if !ok {
		appstateLog.Trace().
			Str("key", KeySelectedTask).
			Msg("key not found; storing + loading default")
		return ""
	}
	appstateLog.Trace().
		Str("key", KeySelectedTask).
		Msgf("loading %#v", task)
	return task.(string)
}

func SetSelectedTask(newTask string) {
	appstateLog.Trace().
		Str("key", KeySelectedTask).
		Msgf("storing %#v", newTask)
	syncMap.Store(KeySelectedTask, newTask)
	chanSelectedTask <- rxgo.Of(newTask)
}

func GetGUIStarted() bool {
	started, ok := syncMap.LoadOrStore(KeyGUIStarted, false)
	if !ok {
		appstateLog.Trace().
			Str("key", KeyGUIStarted).
			Msg("key not found; storing + loading default")
		return false
	}
	appstateLog.Trace().
		Str("key", KeyGUIStarted).
		Msgf("loading %#v", started)
	return started.(bool)
}

func SetGUIStarted(isStarted bool) {
	appstateLog.Trace().
		Str("key", KeyGUIStarted).
		Msgf("storing %#v", isStarted)
	syncMap.Store(KeyGUIStarted, isStarted)
}
