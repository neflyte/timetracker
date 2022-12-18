package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
	log := logger.GetFuncLogger(m.log, "Init")
	m.createButton = widget.NewButtonWithIcon("New", theme.ContentAddIcon(), func() {})
	m.editButton = widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), m.doEditTask)
	m.deleteButton = widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {})
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
	log.Debug().
		Float32("width", siz.Width).
		Float32("height", siz.Height).
		Msg("content size")
	if siz.Width < minimumWindowWidth {
		siz.Width = minimumWindowWidth
	}
	if siz.Height < minimumWindowHeight {
		siz.Height = minimumWindowHeight
	}
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
		log.Err(err).
			Msg("error reading all tasks")
		return
	}
	log.Debug().
		Int("count", len(tasks)).
		Msg("read tasks successfully")
	m.taskSelector.SetList(models.TaskDatas(tasks).AsTaskList())
}

func (m *manageWindowV2Impl) handleTaskSelectorEvent(item interface{}) {
	log := logger.GetFuncLogger(m.log, "handleTaskSelectorEvent")
	switch event := item.(type) {
	case widgets.TaskSelectorSelectedEvent:
		if event.SelectedTask != nil {
			log.Debug().
				Str("selected", event.SelectedTask.String()).
				Msg("got selected task")
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
	log := logger.GetFuncLogger(m.log, "handleEditTaskResult")
	editedTask := m.taskEditor.Task()
	if editedTask == nil {
		log.Error().
			Msg("edited task was nil; this is unexpected")
		return
	}
	// save task to database
	err := editedTask.Update(false)
	if err != nil {
		log.Err(err).
			Msg("error updating task")
		return
	}
	// refresh task list
	go m.refreshTasks()
}
