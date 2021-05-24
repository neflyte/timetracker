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
	"reflect"
)

const (
	noSelectionIndex              = widget.ListItemID(-1)
	manageWindowContainerRowsCols = 2
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
	taskList     []string

	isEditing      bool
	isDirty        bool
	selectedTaskID widget.ListItemID
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
	m.TaskEditor = widgets.NewTaskEditor()
	m.ListTasks = widget.NewListWithData(m.BindTaskList, m.listTasksCreateItem, m.listTasksUpdateItem)
	m.ListTasks.OnSelected = m.taskWasSelected
	// setup observables
	m.TaskEditor.Observables()[widgets.TaskEditorTaskSavedEventKey].ForEach(
		m.doEditSave,
		func(err error) {
			m.Log.Err(err).Msg("error from TaskSaved observable")
		},
		func() {
			m.Log.Debug().Msgf("observable %s is finished", widgets.TaskEditorTaskSavedEventKey)
		},
	)
	m.TaskEditor.Observables()[widgets.TaskEditorEditCancelledEventKey].ForEach(
		m.doEditCancel,
		func(err error) {
			m.Log.Err(err).Msg("error from EditCancelled observable")
		},
		func() {
			m.Log.Debug().Msgf("observable %s is finished", widgets.TaskEditorEditCancelledEventKey)
		},
	)
	// setup layout
	m.Container = container.NewPadded(
		container.NewAdaptiveGrid(
			manageWindowContainerRowsCols,
			container.NewVScroll(m.ListTasks),
			m.TaskEditor,
		),
	)
	m.Window.SetCloseIntercept(m.Hide)
	m.Window.SetContent(m.Container)
	m.Window.SetFixedSize(true)
	m.Window.Resize(MinimumWindowSize)
}

func (m *manageWindow) Get() manageWindow {
	return *m
}

func (m *manageWindow) Show() {
	m.refreshTasks()
	m.TaskEditor.Disable()
	m.Window.Show()
}

func (m *manageWindow) Hide() {
	log := m.Log.With().Str("func", "Hide").Logger()
	err := m.BindTaskList.Set(make([]string, 0))
	if err != nil {
		log.Err(err).Msg("error resetting task list")
	}
	m.TaskEditor.Disable()
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
	synopses := make([]string, len(tasks))
	for idx, task := range tasks {
		synopses[idx] = task.Synopsis
	}
	err = m.BindTaskList.Set(synopses)
	if err != nil {
		log.Err(err).Msg("error setting tasks")
	}
}

func (m *manageWindow) doEditSave(item interface{}) {
	log := m.Log.With().Str("func", "doEditSave").Logger()
	dirtyTask, ok := item.(*models.TaskData)
	if !ok {
		// TODO: show error dialog
		log.Error().Msgf("could not cast interface{} to *models.TaskData; got type %s", reflect.TypeOf(item).String())
		return
	}
	err := m.saveChanges(dirtyTask)
	if err != nil {
		// TODO: show error dialog
		log.Err(err).Msg("error saving changes to task")
		return
	}
	m.doneEditing()
}

func (m *manageWindow) doEditCancel(item interface{}) {
	log := m.Log.With().Str("func", "doEditCancel").Logger()
	editCancelled, ok := item.(bool)
	if !ok {
		log.Error().Msgf("could not cast interface{} to bool; got type %s", reflect.TypeOf(item).String())
		return
	}
	if !editCancelled {
		// Only do something if we got a true value
		return
	}
	// check if we are dirty and prompt to save if we are
	if m.TaskEditor.IsDirty() {
		dialog.NewConfirm(
			"Save changes?",
			"You have unsaved changes. Would you like to save them?",
			func(saveChanges bool) {
				if saveChanges {
					err := m.saveChanges(m.TaskEditor.GetDirtyTask())
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

func (m *manageWindow) saveChanges(dirtyTask *models.TaskData) error {
	log := m.Log.With().Str("func", "saveChanges").Logger()
	if m.isEditing && m.TaskEditor.IsDirty() {
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
	log := m.Log.With().Str("func", "doneEditing").Logger()
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
	m.TaskEditor.Disable()
}

func (m *manageWindow) taskWasSelected(id widget.ListItemID) {
	log := m.Log.With().
		Str("func", "taskWasSelected").
		Int("listItemID", id).
		Logger()
	if m.isEditing && m.TaskEditor.IsDirty() {
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
	m.TaskEditor.Enable()
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

func (m *manageWindow) listTasksCreateItem() fyne.CanvasObject {
	return widgets.NewTasklistItem()
}

func (m *manageWindow) listTasksUpdateItem(item binding.DataItem, canvasObject fyne.CanvasObject) {
	log := m.Log.With().Str("func", "listTasksUpdateItem").Logger()
	taskSynopsisBinding := item.(binding.String)
	taskSyn, err := taskSynopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting task synopsis from binding")
		return
	}
	log = log.With().Str("taskSyn", taskSyn).Logger()
	td := new(models.TaskData)
	td.Synopsis = taskSyn
	err = td.Load(false)
	if err != nil {
		log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
		return
	}
	tasklistItem, ok := canvasObject.(*widgets.TasklistItem)
	if !ok {
		log.Error().Msgf("error getting tasklistItem widget; got %s", reflect.TypeOf(canvasObject).String())
		return
	}
	log.Trace().Msgf("setting task=%s", td.String())
	// TODO: trim subtitle to 64 chars; use ellipsis if >64 chars
	err = tasklistItem.SetTask(td)
	if err != nil {
		log.Err(err).Msg("error setting task on tasklistItem")
	}
}
