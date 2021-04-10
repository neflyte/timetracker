package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"github.com/neflyte/timetracker/internal/appstate"
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
	for {
		mtx.Lock()
		// get ttw
		if ttw.Window != nil {
			queueLen := windowQueue.Len()
			if queueLen > 0 {
				windowsToClose := make([]fyne.Window, queueLen)
				for idx := 0; idx < queueLen; idx++ {
					windowsToClose[idx] = windowQueue.PopFront().(fyne.Window)
				}
				// if the main window is hidden, show it
				windowIsShowing := appstate.GetTTWindowEventLoopRunning()
				if !windowIsShowing {
					ShowTimetrackerWindow(nil)
				}
				// Close the windows we need to close
				for _, window := range windowsToClose {
					window.Close()
				}
				// close the window if it's still open; it will hide instead
				if appstate.GetTTWindowEventLoopRunning() && !windowIsShowing {
					if ttw.Window != nil {
						(*ttw.Window).Close()
					}
				}
			}
		}
		mtx.Unlock()
		select {
		case <-windowCloserQuitChan:
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
