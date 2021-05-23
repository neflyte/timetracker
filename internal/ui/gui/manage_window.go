package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
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
	App           *fyne.App
	Window        fyne.Window
	Container     *fyne.Container
	ListTasks     *widget.List
	Log           zerolog.Logger
	BtnEditSave   *widget.Button
	BtnEditCancel *widget.Button

	BindTaskList            binding.ExternalStringList
	BindTaskEditSynopsis    binding.ExternalString
	BindTaskEditDescription binding.ExternalString

	isEditing           bool
	isDirty             bool
	selectedTaskID      widget.ListItemID
	taskList            []string
	taskEditSynopsis    string
	taskEditDescription string
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
	m.BindTaskEditSynopsis = binding.BindString(&m.taskEditSynopsis)
	m.BindTaskEditDescription = binding.BindString(&m.taskEditDescription)
	// setup widgets
	m.ListTasks = widget.NewListWithData(m.BindTaskList, m.listTasksCreateItem, m.listTasksUpdateItem)
	m.ListTasks.OnSelected = m.taskWasSelected
	m.BtnEditSave = widget.NewButtonWithIcon("Save", theme.ConfirmIcon(), m.doEditSave)
	m.BtnEditCancel = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), m.doEditCancel)
	// setup layout
	m.Container = container.NewPadded(
		container.NewHSplit(
			m.ListTasks,
			container.NewPadded(container.NewVBox(
				// TODO: entry components
				container.NewHBox(m.BtnEditCancel, m.BtnEditSave),
			)),
		),
	)
	m.Window.SetCloseIntercept(m.Hide)
	m.Window.SetContent(m.Container)
	m.Window.SetFixedSize(false)
	m.Window.Resize(MinimumWindowSize)
	// Load window data in a goroutine
	go m.InitWindowData()
}

func (m *manageWindow) InitWindowData() {
	log := m.Log.With().Str("func", "InitWindowData").Logger()
	log.Trace().Msg("started")
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
	log.Trace().Msg("finished")
}

func (m *manageWindow) Get() manageWindow {
	return *m
}

func (m *manageWindow) Show() {
	m.Window.Show()
}

func (m *manageWindow) Hide() {
	m.Window.Hide()
}

func (m *manageWindow) Close() {
	m.Window.Close()
}

func (m *manageWindow) doEditSave() {
	if !m.isDirty {
		return
	}
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
	if m.isDirty {
		dialog.NewConfirm(
			"Save changes?",
			"You have unsaved changes. Would you like to save them?",
			func(saveChanges bool) {
				if saveChanges {
					err := m.saveChanges()
					if err != nil {
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
	if m.isEditing && m.isDirty {
		newTaskSynopsis, err := m.BindTaskEditSynopsis.Get()
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msg("error getting synopsis")
			return err
		}
		newTaskDescription, err := m.BindTaskEditDescription.Get()
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msg("error getting description")
			return err
		}
		taskSyn := m.taskList[m.selectedTaskID]
		td := new(models.TaskData)
		td.Synopsis = taskSyn
		err = td.Load(false)
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
			return err
		}
		td.Synopsis = newTaskSynopsis
		td.Description = newTaskDescription
		err = td.Update(false)
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msgf("error updating task %s", td.String())
			return err
		}
		err = m.BindTaskList.SetValue(m.selectedTaskID, td.Synopsis)
		if err != nil {
			// TODO: show error dialog
			log.Err(err).Msgf("error updating task list")
			return err
		}
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
	m.isDirty = false
	m.ListTasks.Unselect(m.selectedTaskID)
	m.selectedTaskID = noSelectionIndex
	err := m.BindTaskEditSynopsis.Set("")
	if err != nil {
		log.Err(err).Msg("error clearing synopsis")
	}
	err = m.BindTaskEditDescription.Set("")
	if err != nil {
		log.Err(err).Msg("error clearing description")
	}
	m.BtnEditSave.Disable()
	m.BtnEditCancel.Disable()
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
		m.ListTasks.Unselect(id) // FIXME: do we need this?
		return
	}
	m.selectedTaskID = id
	m.isEditing = true
	m.isDirty = false
	taskSyn := m.taskList[id]
	td := new(models.TaskData)
	td.Synopsis = taskSyn
	err := td.Load(false)
	if err != nil {
		// TODO: show error dialog
		log.Err(err).Msgf("error loading task with synopsis %s", taskSyn)
		return
	}
	err = m.BindTaskEditSynopsis.Set(td.Synopsis)
	if err != nil {
		log.Err(err).Msg("error binding synopsis")
	}
	err = m.BindTaskEditDescription.Set(td.Description)
	if err != nil {
		log.Err(err).Msg("error binding description")
	}
	m.BtnEditCancel.Enable()
	m.BtnEditSave.Disable()
}

func (m *manageWindow) listTasksCreateItem() fyne.CanvasObject {
	return widget.NewCard("new task", "task description", nil)
}
func (m *manageWindow) listTasksUpdateItem(item binding.DataItem, canvasObject fyne.CanvasObject) {
	log := m.Log.With().Str("function", "listTasksUpdateItem").Logger()
	taskSynopsisBinding := item.(binding.String)
	taskSyn, err := taskSynopsisBinding.Get()
	if err != nil {
		log.Err(err).Msg("error getting task synopsis from binding")
		return
	}
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
	taskCard.SetTitle(td.Synopsis)
	// TODO: trim subtitle to 64 chars; use ellipsis if >64 chars
	taskCard.SetSubTitle(td.Description)
}
