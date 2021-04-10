package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
)

const (
	minWidth = 350 // minWidth is the minimum width of the widget in pixels
)

type Tasklist struct {
	widget.Select
}

func NewTasklist(changed func(string)) *Tasklist {
	tl := new(Tasklist)
	tl.ExtendBaseWidget(tl)
	if changed != nil {
		tl.OnChanged = changed
	}
	go tl.Init()
	return tl
}

func (t *Tasklist) Init() {
	log := logger.GetLogger("Tasklist.Init")
	log.Debug().Msg("start")
	td := new(models.TaskData)
	tasks, err := td.LoadAll(false)
	if err != nil {
		log.Err(err).Msg("unable to load task list")
		return
	}
	log.Debug().Msgf("len(tasks)=%d", len(tasks))
	taskStrings := make([]string, len(tasks))
	for idx, task := range tasks {
		taskStrings[idx] = task.String()
	}
	log.Debug().Msgf("taskStrings=%#v", taskStrings)
	t.Options = taskStrings
	log.Debug().Msg("done")
}

func (t *Tasklist) MinSize() fyne.Size {
	minsize := t.Select.MinSize()
	if minsize.Width < minWidth {
		minsize.Width = minWidth
	}
	return minsize
}
