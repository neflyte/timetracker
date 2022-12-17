package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/rs/zerolog"
)

type manageWindowV2 interface {
	windowBase
	Show()
	Hide()
	Close()
}

var _ fyne.Window = (*manageWindowV2Impl)(nil)

type manageWindowV2Impl struct {
	fyne.Window
	log          zerolog.Logger
	app          *fyne.App
	container    *fyne.Container
	buttonHBox   *fyne.Container
	createButton *widget.Button
	editButton   *widget.Button
	deleteButton *widget.Button
	taskSelector *widgets.TaskSelector
	taskEditor   *widgets.TaskEditorV2
}

func newManageWindowV2(app fyne.App) manageWindowV2 {
	mw := &manageWindowV2Impl{
		app:    &app,
		log:    logger.GetStructLogger("manageWindowV2Impl"),
		Window: app.NewWindow("Manage Tasks"),
	}
	err := mw.Init()
	if err != nil {
		mw.log.Err(err).Msg("error initializing window")
	}
	return mw
}

func (m *manageWindowV2Impl) Init() error {
	log := m.log.With().Str("func", "Init").Logger()
	m.createButton = widget.NewButton("New", func() {})
	m.editButton = widget.NewButton("Edit", m.doEditTask)
	m.deleteButton = widget.NewButton("Delete", func() {})
	m.buttonHBox = container.NewBorder(
		nil,
		nil,
		m.createButton,
		container.NewHBox(m.editButton, m.deleteButton),
	)
	m.taskSelector = widgets.NewTaskSelector()
	m.taskSelector.Observable().ForEach(
		m.handleTaskSelectorEvent,
		utils.ObservableErrorHandler("taskSelector", m.log),
		utils.ObservableCloseHandler("taskSelector", m.log),
	)
	m.container = container.NewBorder(m.buttonHBox, nil, nil, nil, m.taskSelector)
	m.taskEditor = widgets.NewTaskEditorV2()
	m.Window.SetCloseIntercept(m.Hide)
	m.Window.SetContent(m.container)
	// get the size of the content with everything visible
	siz := m.Window.Content().Size()
	log.Debug().Msgf("content size: %#v", siz)
	// HACK: add a bit of a height buffer, so we can try to fit everything in the window nicely
	siz.Height += float32(windowHeightBuffer)
	// resize the window to fit the content
	m.Window.Resize(siz)
	return nil
}

func (m *manageWindowV2Impl) Hide() {
	m.Window.Hide()
}

func (m *manageWindowV2Impl) Close() {
	m.Window.Close()
}

func (m *manageWindowV2Impl) Show() {
	m.Window.Show()
	go m.refreshTasks()
}

func (m *manageWindowV2Impl) refreshTasks() {
	log := logger.GetFuncLogger(m.log, "refreshTasks")
	tasks, err := models.NewTask().LoadAll(false)
	if err != nil {
		log.Err(err).Msg("error reading all tasks")
		return
	}
	log.Trace().Msgf("read %d tasks", len(tasks))
	m.taskSelector.SetList(models.TaskDatas(tasks).AsTaskList())
}

func (m *manageWindowV2Impl) handleTaskSelectorEvent(item interface{}) {
	log := m.log.With().Str("func", "handleTaskSelectorEvent").Logger()
	switch event := item.(type) {
	case widgets.TaskSelectorSelectedEvent:
		if event.SelectedTask != nil {
			log.Debug().Msgf("task selected: %s", event.SelectedTask.String())
		}
	}
}

func (m *manageWindowV2Impl) doEditTask() {
	if m.taskSelector.HasSelected() {
		m.taskEditor.SetTask(m.taskSelector.Selected())
		dialog.NewCustomConfirm(
			"Edit task", // i18n
			"SAVE",      // i18n
			"CANCEL",    // i18n
			container.NewMax(m.taskEditor),
			m.handleEditTaskResult,
			m.Window,
		).Show()
	}
}

func (m *manageWindowV2Impl) handleEditTaskResult(saved bool) {
	if !saved {
		return
	}
	editedTask := m.taskEditor.Task()
	// TODO: save task to database
	// TODO: refresh list
}
