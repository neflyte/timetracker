package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/phf/go-queue/queue"
	"sync"
	"time"
)

const (
	windowCloserLoopTick = 5
)

var (
	ErrWindowCloserAlreadyRunning = errors.New("window closer is already running")
	ErrWindowCloserNotRunning     = errors.New("window closer is not running")

	mtx                  = sync.RWMutex{}
	windowQueue          = queue.New()
	windowCloserQuitChan = make(chan bool)
	windowCloserRunning  = false
)

func windowCloser() {
	log := logger.GetLogger("windowCloser")
	log.Trace().Msg("starting")
	defer log.Trace().Msg("done")
	for {
		mtx.Lock()
		log.Trace().Msg("locked mtx")
		// get ttw
		/*		if ttw.Window != nil {
				queueLen := windowQueue.Len()
				if queueLen > 0 {
					log.Trace().Msgf("queueLen=%d", queueLen)
					windowsToClose := make([]fyne.Window, queueLen)
					for idx := 0; idx < queueLen; idx++ {
						windowsToClose[idx] = windowQueue.PopFront().(fyne.Window)
					}
					// if the main window is hidden, show it
					windowIsShowing := appstate.GetTTWindowEventLoopRunning()
					if !windowIsShowing {
						log.Trace().Msg("showing main window")
						ShowTimetrackerWindow(nil)
					}
					// Close the windows we need to close
					for _, window := range windowsToClose {
						windowTitle := window.Title()
						log.Trace().Str("title", windowTitle).Msg("closing window")
						window.Close()
					}
					// close the window if it's still open; it will hide instead
					if appstate.GetTTWindowEventLoopRunning() && !windowIsShowing {
						log.Trace().Msg("closing/hiding main window")
						CloseTimetrackerWindow(nil)
					}
				}
			}*/
		mtx.Unlock()
		log.Trace().Msgf("unlocked mtx; waiting %d seconds", windowCloserLoopTick)
		select {
		case <-windowCloserQuitChan:
			log.Trace().Msg("quit channel fired; exiting function")
			return
		case <-time.After(windowCloserLoopTick * time.Second):
			break
		}
	}
}

func StartWindowCloser() error {
	if windowCloserRunning {
		return ErrWindowCloserAlreadyRunning
	}
	windowCloserRunning = true
	go windowCloser()
	return nil
}

func StopWindowCloser() error {
	if !windowCloserRunning {
		return ErrWindowCloserNotRunning
	}
	windowCloserQuitChan <- true
	windowCloserRunning = false
	return nil
}

func CloseWindow(w fyne.Window) {
	mtx.Lock()
	defer mtx.Unlock()
	windowQueue.PushBack(w)
}
