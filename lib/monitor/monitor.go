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

// ServiceUpdateEvent represents an event indicating data has been refreshed
type ServiceUpdateEvent struct{}

// ServiceData is the main data struct of the Service
type ServiceData struct {
	log                 zerolog.Logger
	runningTimesheet    models.Timesheet
	timesheetError      error
	tsModel             models.Timesheet
	quitChan            chan bool
	commandChan         chan rxgo.Item
	timesheetStatus     int
	runningTimesheetMtx sync.RWMutex
	timesheetStatusMtx  sync.RWMutex
	timesheetErrorMtx   sync.RWMutex
	running             bool
}

// Service is the interface to the monitor service functions
type Service interface {
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

// NewService returns an initialized Service using the supplied quit channel
func NewService(quitChannel chan bool) Service {
	if quitChannel == nil {
		panic("NewService(): quitChan cannot be nil")
	}
	return &ServiceData{
		log:                 logger.GetStructLogger("ServiceData"),
		quitChan:            quitChannel,
		commandChan:         make(chan rxgo.Item, monitorServiceCommandChanSize),
		runningTimesheetMtx: sync.RWMutex{},
		timesheetStatus:     constants.TimesheetStatusIdle,
		timesheetStatusMtx:  sync.RWMutex{},
		timesheetErrorMtx:   sync.RWMutex{},
		tsModel:             models.NewTimesheet(),
	}
}

// Start starts the monitor, optionally using a start channel
func (m *ServiceData) Start(startChannel chan bool) {
	log := logger.GetFuncLogger(m.log, "Start")
	if m.IsRunning() {
		log.Warn().
			Msg("service is already running; will not start it again")
		return
	}
	go m.actionLoop(startChannel)
}

// Stop stops the monitor
func (m *ServiceData) Stop() {
	if !m.IsRunning() {
		return
	}
	m.quitChan <- true
}

// IsRunning returns the running status of the monitor
func (m *ServiceData) IsRunning() bool {
	return m.running
}

// Observable returns an RxGo Observable for the monitor's command channel
func (m *ServiceData) Observable() rxgo.Observable {
	return rxgo.FromEventSource(m.commandChan)
}

func (m *ServiceData) TimesheetStatus() int {
	m.timesheetStatusMtx.RLock()
	defer m.timesheetStatusMtx.RUnlock()
	return m.timesheetStatus
}

func (m *ServiceData) SetTimesheetStatus(status int) {
	m.timesheetStatusMtx.Lock()
	defer m.timesheetStatusMtx.Unlock()
	m.timesheetStatus = status
}

func (m *ServiceData) TimesheetError() error {
	m.timesheetErrorMtx.RLock()
	defer m.timesheetErrorMtx.RUnlock()
	return m.timesheetError
}

func (m *ServiceData) SetTimesheetError(err error) {
	m.timesheetErrorMtx.Lock()
	defer m.timesheetErrorMtx.Unlock()
	m.timesheetError = err
}

func (m *ServiceData) RunningTimesheet() models.Timesheet {
	m.runningTimesheetMtx.RLock()
	defer m.runningTimesheetMtx.RUnlock()
	return m.runningTimesheet
}

func (m *ServiceData) SetRunningTimesheet(timesheet models.Timesheet) {
	m.runningTimesheetMtx.Lock()
	defer m.runningTimesheetMtx.Unlock()
	m.runningTimesheet = timesheet
}

func (m *ServiceData) actionLoop(startChan chan bool) {
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
		m.commandChan <- rxgo.Of(ServiceUpdateEvent{})
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

func (m *ServiceData) updateTimesheet() {
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
