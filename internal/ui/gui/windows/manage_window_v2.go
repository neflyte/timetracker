package windows

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/gui/dialogs"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
)

const (
	manageWindowV2EventChannelSize = 2
)

type ManageWindowV2TasksChangedEvent struct{}

type manageWindowV2 interface {
	windowBase
	Show()
	Hide()
	Close()
	Observable() rxgo.Observable
}

var _ fyne.Window = (*manageWindowV2Impl)(nil)

type manageWindowV2Impl struct {
	fyne.Window
	log          zerolog.Logger
	container    *fyne.Container
	buttonHBox   *fyne.Container
	createButton *widget.Button
	editButton   *widget.Button
	deleteButton *widget.Button
	taskSelector *widgets.TaskSelector
	taskEditor   *widgets.TaskEditorV2
	eventChan    chan rxgo.Item
}

func newManageWindowV2(app fyne.App) manageWindowV2 {
	mw := &manageWindowV2Impl{
		log:       logger.GetStructLogger("manageWindowV2Impl"),
		eventChan: make(chan rxgo.Item, manageWindowV2EventChannelSize),
		Window:    app.NewWindow("Manage Tasks"), // i18n
	}
	err := mw.Init()
	if err != nil {
		mw.log.
			Err(err).
			Msg("error initializing window")
	}
	return mw
}

func (m *manageWindowV2Impl) Init() error {
	m.createButton = widget.NewButtonWithIcon("New", theme.ContentAddIcon(), m.doCreateTask)
	m.editButton = widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), m.doEditTask)
	m.deleteButton = widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), m.doDeleteTask)
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
	m.Window.SetIcon(icons.IconV2)
	// resize the window to fit the content
	resizeToMinimum(m.Window, minimumWindowWidth, minimumWindowHeight)
	return nil
}

func (m *manageWindowV2Impl) Hide() {
	m.Window.Hide()
}

func (m *manageWindowV2Impl) Close() {
	m.Window.Close()
}

func (m *manageWindowV2Impl) Show() {
	m.taskSelector.Reset()
	m.Window.Show()
}

func (m *manageWindowV2Impl) Observable() rxgo.Observable {
	return rxgo.FromEventSource(m.eventChan)
}

func (m *manageWindowV2Impl) handleTaskSelectorEvent(item interface{}) {
	log := logger.GetFuncLogger(m.log, "handleTaskSelectorEvent")
	if event, ok := item.(widgets.TaskSelectorSelectedEvent); ok {
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
		editDialog := dialog.NewCustomConfirm(
			"Edit task", // i18n
			"SAVE",      // i18n
			"CANCEL",    // i18n
			m.taskEditor,
			m.handleEditTaskResult,
			m.Window,
		)
		dialogs.ResizeDialogToWindowWithPadding(editDialog, m.Window, dialogSizeOffset)
		editDialog.Show()
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
	// re-filter task list
	go m.taskSelector.FilterTasks()
	// send a refresh event
	m.eventChan <- rxgo.Of(ManageWindowV2TasksChangedEvent{})
}

func (m *manageWindowV2Impl) doCreateTask() {
	m.taskEditor.Reset()
	createDialog := dialog.NewCustomConfirm(
		"Create new task", // i18n
		"SAVE",            // i18n
		"CANCEL",          // i18n
		m.taskEditor,
		m.handleCreateTaskResult,
		m.Window,
	)
	dialogs.ResizeDialogToWindowWithPadding(createDialog, m.Window, dialogSizeOffset)
	createDialog.Show()
}

func (m *manageWindowV2Impl) handleCreateTaskResult(created bool) {
	if !created {
		return
	}
	log := logger.GetFuncLogger(m.log, "handleCreateTaskResult")
	newTask := m.taskEditor.Task()
	if newTask == nil {
		log.Error().
			Msg("new task was nil; this is unexpected")
		return
	}
	err := newTask.Create()
	if err != nil {
		log.Err(err).
			Msg("error creating new task")
		return
	}
	// re-filter task list
	go m.taskSelector.FilterTasks()
	// send a refresh event
	m.eventChan <- rxgo.Of(ManageWindowV2TasksChangedEvent{})
}

func (m *manageWindowV2Impl) doDeleteTask() {
	selectedTask := m.taskSelector.Selected()
	if selectedTask == nil {
		return
	}
	dialog.NewConfirm(
		"Delete task", // i18n
		fmt.Sprintf("Are you sure you want to delete this task?\n\n%s", selectedTask), // i18n
		m.handleDeleteTaskResult,
		m.Window,
	).Show()
}

func (m *manageWindowV2Impl) handleDeleteTaskResult(deleted bool) {
	if !deleted {
		return
	}
	selectedTask := m.taskSelector.Selected()
	if selectedTask == nil {
		return
	}
	log := logger.GetFuncLogger(m.log, "handleDeleteTaskResult")
	err := selectedTask.Delete()
	if err != nil {
		log.Err(err).
			Str("task", selectedTask.String()).
			Msg("error deleting task")
	}
	log.Debug().
		Msg("deleted task successfully")
	// re-filter task list
	go m.taskSelector.FilterTasks()
	// send a refresh event
	m.eventChan <- rxgo.Of(ManageWindowV2TasksChangedEvent{})
}
