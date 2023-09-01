package monitor

import (
	"errors"
	"sync"
	"time"

	"github.com/neflyte/timetracker/lib/constants"
	ttErrors "github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
)

const (
	monitorServiceCommandChanSize = 5
)

type MonitorServiceUpdateEvent struct{}

type MonitorData struct {
	running             bool
	quitChan            chan bool
	commandChan         chan rxgo.Item
	log                 zerolog.Logger
	runningTimesheet    models.Timesheet
	runningTimesheetMtx sync.RWMutex
	timesheetStatus     int
	timesheetStatusMtx  sync.RWMutex
	timesheetError      error
	timesheetErrorMtx   sync.RWMutex
	tsModel             models.Timesheet
}

type MonitorService interface {
	Start(startChannel chan bool)
	Stop()
	IsRunning() bool
	Observable() rxgo.Observable
	RunningTimesheet() models.Timesheet
	SetRunningTimesheet(timesheet models.Timesheet)
	TimesheetStatus() int
	SetTimesheetStatus(status int)
	TimesheetError() error
	SetTimesheetError(err error)
}

func NewMonitorService(quitChannel chan bool) MonitorService {
	if quitChannel == nil {
		panic("NewMonitorService(): quitChan cannot be nil")
	}
	return &MonitorData{
		log:                 logger.GetStructLogger("MonitorData"),
		quitChan:            quitChannel,
		commandChan:         make(chan rxgo.Item, monitorServiceCommandChanSize),
		runningTimesheetMtx: sync.RWMutex{},
		timesheetStatus:     constants.TimesheetStatusIdle,
		timesheetStatusMtx:  sync.RWMutex{},
		timesheetErrorMtx:   sync.RWMutex{},
		tsModel:             models.NewTimesheet(),
	}
}

func (m *MonitorData) Start(startChannel chan bool) {
	log := logger.GetFuncLogger(m.log, "Start")
	if m.IsRunning() {
		log.Warn().
			Msg("service is already running; will not start it again")
		return
	}
	go m.actionLoop(startChannel)
}

func (m *MonitorData) Stop() {
	if !m.IsRunning() {
		return
	}
	m.quitChan <- true
}

func (m *MonitorData) IsRunning() bool {
	return m.running
}

func (m *MonitorData) Observable() rxgo.Observable {
	return rxgo.FromEventSource(m.commandChan)
}

func (m *MonitorData) TimesheetStatus() int {
	m.timesheetStatusMtx.RLock()
	defer m.timesheetStatusMtx.RUnlock()
	return m.timesheetStatus
}

func (m *MonitorData) SetTimesheetStatus(status int) {
	m.timesheetStatusMtx.Lock()
	defer m.timesheetStatusMtx.Unlock()
	m.timesheetStatus = status
}

func (m *MonitorData) TimesheetError() error {
	m.timesheetErrorMtx.RLock()
	defer m.timesheetErrorMtx.RUnlock()
	return m.timesheetError
}

func (m *MonitorData) SetTimesheetError(err error) {
	m.timesheetErrorMtx.Lock()
	defer m.timesheetErrorMtx.Unlock()
	m.timesheetError = err
}

func (m *MonitorData) RunningTimesheet() models.Timesheet {
	m.runningTimesheetMtx.RLock()
	defer m.runningTimesheetMtx.RUnlock()
	return m.runningTimesheet
}

func (m *MonitorData) SetRunningTimesheet(timesheet models.Timesheet) {
	m.runningTimesheetMtx.Lock()
	defer m.runningTimesheetMtx.Unlock()
	m.runningTimesheet = timesheet
}

func (m *MonitorData) actionLoop(startChan chan bool) {
	log := logger.GetFuncLogger(m.log, "actionLoop")
	if m.IsRunning() {
		log.Warn().
			Msg("service is already running; will not start it again")
		return
	}
	m.running = true
	defer func() {
		m.running = false
	}()
	if startChan != nil {
		log.Debug().
			Msg("startChan is non-nil; receiving from channel before starting")
		<-startChan
	}
	log.Debug().
		Msg("starting loop")
	for {
		log.Trace().
			Msg("updating timesheet")
		m.updateTimesheet()
		log.Trace().
			Msg("sending UpdateEvent")
		m.commandChan <- rxgo.Of(MonitorServiceUpdateEvent{})
		select {
		case <-m.quitChan:
			log.Debug().
				Msg("received from quitChan; exiting loop")
			return
		case <-time.After(constants.ActionLoopDelaySeconds * time.Second):
			log.Trace().
				Int("seconds", constants.ActionLoopDelaySeconds).
				Msg("finished delay")
			break
		}
	}
}

func (m *MonitorData) updateTimesheet() {
	log := logger.GetFuncLogger(m.log, "updateTimesheet")
	// Get the running timesheet, if any
	runningTS, err := m.tsModel.RunningTimesheet()
	// Acquire write locks before checking the error or the result
	m.runningTimesheetMtx.Lock()
	m.timesheetStatusMtx.Lock()
	m.timesheetErrorMtx.Lock()
	defer func() {
		m.runningTimesheetMtx.Unlock()
		m.timesheetStatusMtx.Unlock()
		m.timesheetErrorMtx.Unlock()
	}()
	// Error getting the timesheet
	if err != nil && !errors.Is(err, ttErrors.ErrNoRunningTask{}) {
		log.Err(err).
			Msg("unable to get running timesheet")
		m.runningTimesheet = nil
		m.timesheetStatus = constants.TimesheetStatusError
		m.timesheetError = err
		return
	}
	// No open timesheets
	if runningTS == nil {
		log.Trace().
			Msg("there is no running timesheet")
		m.runningTimesheet = nil // Reset running timesheet
		m.timesheetStatus = constants.TimesheetStatusIdle
		m.timesheetError = nil
		return
	}
	// A timesheet is open
	log.Trace().
		Msg("a timesheet is running")
	m.runningTimesheet = runningTS.Data()
	m.timesheetStatus = constants.TimesheetStatusRunning
	m.timesheetError = nil
}
