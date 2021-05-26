package gui

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
	"github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	"reflect"
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

	App           *fyne.App
	Window        fyne.Window
	Container     *fyne.Container
	HSplit        *container.Split
	ListTasks     *widget.List
	AddTaskButton *widget.Button
	TaskEditor    *widgets.TaskEditor

	taskListChangedChannel    chan rxgo.Item
	TaskListChangedObservable rxgo.Observable

	BindTaskList binding.ExternalStringList
	taskList     []string

	isEditing      bool
	selectedTaskID widget.ListItemID
}

func NewManageWindow(app fyne.App) ManageWindow {
	mw := &manageWindow{
		App:                    &app,
		Window:                 app.NewWindow("Manage Tasks"),
		Log:                    logger.GetStructLogger("ManageWindow"),
		taskList:               make([]string, 0),
		selectedTaskID:         noSelectionIndex,
		taskListChangedChannel: make(chan rxgo.Item, ManageWindowEventChannelBufferSize),
	}
	mw.Init()
	return mw
}

func (m *manageWindow) Init() {
	// setup observables
	m.TaskListChangedObservable = rxgo.FromEventSource(m.taskListChangedChannel)
	// setup bindings
	m.BindTaskList = binding.BindStringList(&m.taskList)
	// setup widgets
	m.TaskEditor = widgets.NewTaskEditor()
	m.ListTasks = widget.NewListWithData(m.BindTaskList, m.listTasksCreateItem, m.listTasksUpdateItem)
	m.ListTasks.OnSelected = m.taskWasSelected
	m.AddTaskButton = widget.NewButtonWithIcon("NEW", theme.ContentAddIcon(), m.createNewTask)
	// setup observables
	m.TaskEditor.Observables()[widgets.TaskEditorTaskSavedEventKey].ForEach(
		m.doEditSave,
		func(err error) {
			m.Log.Err(err).Msg("error from TaskSaved observable")
		},
		func() {
			m.Log.Trace().Msgf("observable %s is finished", widgets.TaskEditorTaskSavedEventKey)
		},
	)
	m.TaskEditor.Observables()[widgets.TaskEditorEditCancelledEventKey].ForEach(
		m.doEditCancel,
		func(err error) {
			m.Log.Err(err).Msg("error from EditCancelled observable")
		},
		func() {
			m.Log.Trace().Msgf("observable %s is finished", widgets.TaskEditorEditCancelledEventKey)
		},
	)
	// setup layout
	m.HSplit = container.NewHSplit(
		container.NewPadded(
			container.NewBorder(
				nil,
				container.NewCenter(m.AddTaskButton),
				nil,
				nil,
				container.NewVScroll(m.ListTasks),
			),
		),
		container.NewPadded(m.TaskEditor),
	)
	m.Container = container.NewMax(m.HSplit)
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
	m.TaskEditor.Hide()
	m.Window.Show()
	m.jiggleHSplit()
}

func (m *manageWindow) Hide() {
	log := logger.GetFuncLogger(m.Log, "Hide")
	err := m.BindTaskList.Set(make([]string, 0))
	if err != nil {
		log.Err(err).Msg("error resetting task list")
	}
	m.TaskEditor.Hide()
	m.Window.Hide()
}

func (m *manageWindow) Close() {
	m.Window.Close()
}

func (m *manageWindow) refreshTasks() {
	log := logger.GetFuncLogger(m.Log, "refreshTasks")
	tasks, err := models.Task(new(models.TaskData)).LoadAll(false)
	if err != nil {
		log.Err(err).Msg("error reading all tasks")
		return
	}
	log.Trace().Msgf("read %d tasks", len(tasks))
	synopses := make([]string, len(tasks))
	for idx, task := range tasks {
		synopses[idx] = task.Synopsis
	}
	err = m.BindTaskList.Set(synopses)
	if err != nil {
		log.Err(err).Msg("error setting tasks")
	}
	m.jiggleHSplit()
}

// jiggleHSplit moves the HSplit component to the left and back again so each side
// is forced to redraw with the correct sizing
func (m *manageWindow) jiggleHSplit() {
	oldOffset := m.HSplit.Offset
	newOffset := oldOffset - 1.0
	m.HSplit.SetOffset(newOffset)
	m.HSplit.SetOffset(oldOffset)
}

func (m *manageWindow) doEditSave(item interface{}) {
	log := logger.GetFuncLogger(m.Log, "doEditSave")
	dirtyTask, ok := item.(models.TaskData)
	if !ok {
		err := fmt.Errorf("could not cast interface{} to models.TaskData; got type %s", reflect.TypeOf(item).String())
		dialog.NewError(err, m.Window).Show()
		log.Error().Msg(err.Error())
		return
	}
	log.Debug().Msgf("saving dirtyTask %s", dirtyTask.String())
	err := m.saveChanges(dirtyTask)
	if err != nil {
		dialog.NewError(err, m.Window).Show()
		log.Err(err).Msg("error saving changes to task")
		return
	}
	m.doneEditing()
	m.refreshTasks()
	m.taskListChangedChannel <- rxgo.Of(true)
}

func (m *manageWindow) doEditCancel(item interface{}) {
	log := logger.GetFuncLogger(m.Log, "doEditCancel")
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
					err := m.saveChanges(*m.TaskEditor.GetDirtyTask())
					if err != nil {
						dialog.NewError(err, m.Window).Show()
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

func (m *manageWindow) saveChanges(dirtyTask models.TaskData) error {
	log := logger.GetFuncLogger(m.Log, "saveChanges")
	if m.isEditing && m.TaskEditor.IsDirty() {
		log.Debug().Msgf("dirtyTask=%s", dirtyTask.String())
		td := new(models.TaskData)
		// Re-load the task record first if it exists
		if dirtyTask.ID > 0 {
			td.Synopsis = dirtyTask.Synopsis
			log.Debug().Msgf("re-loading record for task ID %d (%s)", dirtyTask.ID, dirtyTask.Synopsis)
			err := td.Load(false)
			if err != nil {
				dialog.NewError(err, m.Window).Show()
				log.Err(err).Msgf("error loading task with synopsis %s", dirtyTask.Synopsis)
				return err
			}
		}
		td.Synopsis = dirtyTask.Synopsis
		td.Description = dirtyTask.Description
		if td.ID > 0 {
			log.Trace().Msgf("updating record for task ID %d (%s)", td.ID, td.Synopsis)
			err := td.Update(false)
			if err != nil {
				dialog.NewError(err, m.Window).Show()
				log.Err(err).Msgf("error updating task %s", td.String())
				return err
			}
			log.Debug().Msgf("record for task ID %d (%s) updated successfully", td.ID, td.Synopsis)
		} else {
			log.Trace().Msgf("creating new task record (%s)", td.Synopsis)
			err := td.Create()
			if err != nil {
				dialog.NewError(err, m.Window).Show()
				log.Err(err).Msgf("error creating task %s", td.String())
				return err
			}
			log.Debug().Msgf("new task record (%s) created successfully", td.Synopsis)
		}
		m.refreshTasks()
		return nil
	}
	return errors.New("task is not being edited or task is not dirty")
}

func (m *manageWindow) doneEditing() {
	log := logger.GetFuncLogger(m.Log, "doneEditing")
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
	m.TaskEditor.Hide()
	m.jiggleHSplit()
}

func (m *manageWindow) taskWasSelected(id widget.ListItemID) {
	log := logger.GetFuncLogger(m.Log, "taskWasSelected").
		With().Int("listItemID", id).Logger()
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
	m.TaskEditor.Show()
	taskSyn := m.taskList[id]
	td := new(models.TaskData)
	td.Synopsis = taskSyn
	err := td.Load(false)
	if err != nil {
		dialog.NewError(err, m.Window).Show()
		log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
		return
	}
	err = m.TaskEditor.SetTask(td)
	if err != nil {
		log.Err(err).Msg("error setting task")
	}
	m.jiggleHSplit()
}

func (m *manageWindow) createNewTask() {
	// basically the same as taskWasSelected but load an empty task record into the editor instead
	log := logger.GetFuncLogger(m.Log, "createNewTask")
	if m.isEditing && m.TaskEditor.IsDirty() {
		dialog.NewInformation(
			"Unsaved Changes",
			"You have unsaved changes.\nSave them or cancel editing before creating a new task.",
			m.Window,
		).Show()
		return
	}
	m.selectedTaskID = -1
	m.isEditing = true
	m.TaskEditor.Show()
	err := m.TaskEditor.SetTask(new(models.TaskData))
	if err != nil {
		log.Err(err).Msg("error setting empty task")
	}
	m.jiggleHSplit()
}

func (m *manageWindow) listTasksCreateItem() fyne.CanvasObject {
	return widgets.NewTasklistItem()
}

func (m *manageWindow) listTasksUpdateItem(item binding.DataItem, canvasObject fyne.CanvasObject) {
	log := logger.GetFuncLogger(m.Log, "listTasksUpdateItem")
	taskSynopsisBinding, ok := item.(binding.String)
	if !ok {
		log.Error().Msgf("did not get binding.String; got %s instead", reflect.TypeOf(item).String())
		return
	}
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
	err = tasklistItem.SetTask(td)
	if err != nil {
		log.Err(err).Msg("error setting task on tasklistItem")
	}
}
