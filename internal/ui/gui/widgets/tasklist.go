package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/rs/zerolog"
)

const (
	tasklistMinWidth = 350 // tasklistMinWidth is the minimum width of the widget in pixels
)

type Tasklist struct {
	widget.Select
	log zerolog.Logger
}

func NewTasklist(changed func(string)) *Tasklist {
	tl := new(Tasklist)
	tl.ExtendBaseWidget(tl)
	tl.log = logger.GetStructLogger("Tasklist")
	if changed != nil {
		tl.OnChanged = changed
	}
	go tl.Init()
	return tl
}

func (t *Tasklist) Init() {
	log := t.log.With().Str("func", "Init").Logger()
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
	if minsize.Width < tasklistMinWidth {
		minsize.Width = tasklistMinWidth
	}
	return minsize
}
