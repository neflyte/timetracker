package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/rs/zerolog"
)

const (
	noSelectionIndex = widget.ListItemID(-1)
)

type ManageWindow interface {
	Show()
	Hide()
	Close()
	Get() manageWindow
}

type manageWindow struct {
	Log zerolog.Logger

	App        *fyne.App
	Window     fyne.Window
	Container  *fyne.Container
	ListTasks  *widget.List
	TaskEditor *widgets.TaskEditor

	BindTaskList binding.ExternalStringList

	isEditing      bool
	isDirty        bool
	selectedTaskID widget.ListItemID
	taskList       []string
}

func NewManageWindow(app fyne.App) ManageWindow {
	mw := &manageWindow{
		App:            &app,
		Window:         app.NewWindow("Manage Tasks"),
		Log:            logger.GetStructLogger("ManageWindow"),
		taskList:       make([]string, 0),
		selectedTaskID: noSelectionIndex,
	}
	mw.Init()
	return mw
}

func (m *manageWindow) Init() {
	// setup bindings
	m.BindTaskList = binding.BindStringList(&m.taskList)
	// setup widgets
	m.TaskEditor = widgets.NewTaskEditor(m.doEditSave, m.doEditCancel)
	m.ListTasks = widget.NewListWithData(m.BindTaskList, m.listTasksCreateItem, m.listTasksUpdateItem)
	m.ListTasks.OnSelected = m.taskWasSelected
	// setup layout
	m.Container = container.NewMax(container.NewHSplit(
		container.NewPadded(m.ListTasks),
		container.NewPadded(m.TaskEditor),
	))
	m.Window.SetCloseIntercept(m.Hide)
	m.Window.SetContent(m.Container)
	m.Window.SetFixedSize(true)
	m.Window.Resize(MinimumWindowSize)
}

func (m *manageWindow) Get() manageWindow {
	return *m
}

func (m *manageWindow) Show() {
	// Hide editor widgets by default
	m.toggleEditWidgets(false)
	m.refreshTasks()
	m.Window.Show()
}

func (m *manageWindow) Hide() {
	log := m.Log.With().Str("func", "Hide").Logger()
	err := m.BindTaskList.Set(make([]string, 0))
	if err != nil {
		log.Err(err).Msg("error resetting task list")
	}
	m.Window.Hide()
}

func (m *manageWindow) Close() {
	m.Window.Close()
}

func (m *manageWindow) refreshTasks() {
	log := m.Log.With().Str("func", "refreshTasks").Logger()
	tasks, err := models.Task(new(models.TaskData)).LoadAll(false)
	if err != nil {
		log.Err(err).Msg("error reading all tasks")
		return
	}
	log.Debug().Msgf("read %d tasks", len(tasks))
	for _, task := range tasks {
		log.Debug().Msgf("task=%s", task.String())
		err = m.BindTaskList.Append(task.Synopsis)
		if err != nil {
			log.Err(err).Msgf("error appending task %s", task.String())
		}
	}
}

func (m *manageWindow) doEditSave() {
	log := m.Log.With().Str("function", "doEditSave").Logger()
	if !m.TaskEditor.IsDirty() {
		return
	}
	err := m.saveChanges()
	if err != nil {
		// TODO: show error dialog
		log.Err(err).Msg("error saving changes to task")
		return
	}
	m.doneEditing()
}

func (m *manageWindow) doEditCancel() {
	log := m.Log.With().Str("function", "doEditCancel").Logger()
	if !m.isEditing {
		dialog.NewError(
			errors.New("a task is not being edited; please select a task to edit"),
			m.Window,
		).Show()
		return
	}
	// check if we are dirty and prompt to save if we are
	if m.TaskEditor.IsDirty() {
		dialog.NewConfirm(
			"Save changes?",
			"You have unsaved changes. Would you like to save them?",
			func(saveChanges bool) {
				if saveChanges {
					err := m.saveChanges()
					if err != nil {
						// TODO: show error dialog
						log.Err(err).Msg("error saving changes to task")
						return
					}
				}
				m.doneEditing()
			},
			m.Window,
		).Show()
		return
	}
	m.doneEditing()
}

func (m *manageWindow) saveChanges() error {
	log := m.Log.With().Str("function", "saveChanges").Logger()
	if m.isEditing && m.TaskEditor.IsDirty() {
		dirtyTask := m.TaskEditor.GetDirtyTask()
		if dirtyTask == nil {
			log.Error().Msg("dirty task is nil; this is unexpected")
			return errors.New("dirty task is nil")
		}
		taskSyn := m.taskList[m.selectedTaskID]
		td := new(models.TaskData)
		td.Synopsis = taskSyn
		err := td.Load(false)
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
			return err
		}
		td.Synopsis = dirtyTask.Synopsis
		td.Description = dirtyTask.Description
		err = td.Update(false)
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msgf("error updating task %s", td.String())
			return err
		}
		m.refreshTasks()
		return nil
	}
	return errors.New("task is not being edited or task is not dirty")
}

func (m *manageWindow) doneEditing() {
	log := m.Log.With().Str("function", "doneEditing").Logger()
	if !m.isEditing {
		return
	}
	m.isEditing = false
	m.ListTasks.Unselect(m.selectedTaskID)
	m.selectedTaskID = noSelectionIndex
	err := m.TaskEditor.SetTask(new(models.TaskData))
	if err != nil {
		log.Err(err).Msg("error setting task to empty task")
	}
	m.toggleEditWidgets(false)
}

func (m *manageWindow) taskWasSelected(id widget.ListItemID) {
	log := m.Log.With().
		Str("function", "taskWasSelected").
		Int("listItemID", id).
		Logger()
	if m.isEditing && m.isDirty {
		dialog.NewInformation(
			"Unsaved Changes",
			"You have unsaved changes\nSave them or cancel editing before selecting a different task",
			m.Window,
		).Show()
		m.ListTasks.Unselect(id)
		return
	}
	m.selectedTaskID = id
	m.isEditing = true
	m.isDirty = false
	m.toggleEditWidgets(true)
	taskSyn := m.taskList[id]
	td := new(models.TaskData)
	td.Synopsis = taskSyn
	err := td.Load(false)
	if err != nil {
		// TODO: show error dialog
		log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
		return
	}
	err = m.TaskEditor.SetTask(td)
	if err != nil {
		log.Err(err).Msg("error setting task")
	}
}

// TODO: Make a first-class widget for the task list item

func (m *manageWindow) listTasksCreateItem() fyne.CanvasObject {
	return widget.NewCard("", "", container.NewPadded(widget.NewLabel("")))
}

func (m *manageWindow) listTasksUpdateItem(item binding.DataItem, canvasObject fyne.CanvasObject) {
	log := m.Log.With().Str("function", "listTasksUpdateItem").Logger()
	taskSynopsisBinding := item.(binding.String)
	taskSyn, err := taskSynopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting task synopsis from binding")
		return
	}
	log = log.With().Str("taskSyn", taskSyn).Logger()
	taskCard, ok := canvasObject.(*widget.Card)
	if !ok {
		log.Error().Msg("error getting card widget")
		return
	}
	td := new(models.TaskData)
	td.Synopsis = taskSyn
	err = td.Load(false)
	if err != nil {
		log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
		return
	}
	log.Trace().Msgf("setting title=%s, subtitle=%s", td.Synopsis, td.Description)
	taskCard.SetTitle(td.Synopsis)
	// TODO: trim subtitle to 64 chars; use ellipsis if >64 chars
	taskCard.SetSubTitle(td.Description)
	//taskCard.Content.(*fyne.Container).Objects[0].(*widget.Label).SetText(td.Description)
}

func (m *manageWindow) toggleEditWidgets(show bool) {
	if show {
		m.TaskEditor.Show()
		return
	}
	m.TaskEditor.Hide()
}
