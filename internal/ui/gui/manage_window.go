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
	TaskMap       map[string]models.TaskData
	BtnEditSave   *widget.Button
	BtnEditCancel *widget.Button

	BindIsEditing           binding.ExternalBool
	BindTaskList            binding.ExternalStringList
	BindTaskEditSynopsis    binding.ExternalString
	BindTaskEditDescription binding.ExternalString

	isEditing           bool
	taskList            []string
	taskEditSynopsis    string
	taskEditDescription string
}

func NewManageWindow(app fyne.App) ManageWindow {
	mw := &manageWindow{
		App:      &app,
		Window:   app.NewWindow("Manage Tasks"),
		Log:      logger.GetStructLogger("ManageWindow"),
		TaskMap:  make(map[string]models.TaskData),
		taskList: make([]string, 0),
	}
	mw.Init()
	return mw
}

func (m *manageWindow) Init() {
	// setup bindings
	m.BindTaskList = binding.BindStringList(&m.taskList)
	m.BindTaskEditSynopsis = binding.BindString(&m.taskEditSynopsis)
	m.BindTaskEditDescription = binding.BindString(&m.taskEditDescription)
	m.BindIsEditing = binding.BindBool(&m.isEditing)
	// setup widgets
	m.ListTasks = widget.NewListWithData(
		m.BindTaskList,
		func() fyne.CanvasObject {
			return widget.NewCard("new task", "task description", nil)
		},
		func(item binding.DataItem, canvasObject fyne.CanvasObject) {
			taskStringBinding := item.(binding.String)
			taskString, err := taskStringBinding.Get()
			if err != nil {
				m.Log.Err(err).Msg("error getting task string from binding")
				return
			}
			taskCard, ok := canvasObject.(*widget.Card)
			if ok {
				taskCard.SetTitle(taskString)
				task, found := m.TaskMap[taskString]
				if found {
					taskCard.SetSubTitle(task.Description)
				}
			}
		},
	)
	m.ListTasks.OnSelected = func(id widget.ListItemID) {
		log := m.Log.With().Str("function", "OnSelected").Logger()
		isEditing, err := m.BindIsEditing.Get()
		if err != nil {
			log.Err(err).Msg("error getting IsEditing")
			return
		}
		if isEditing {
			dialog.NewInformation(
				"Unsaved Changes",
				"You have unsaved changes. Save them or cancel editing before selecting a different task",
				m.Window,
			).Show()
			return
		}
		err = m.BindIsEditing.Set(true)
		if err != nil {
			log.Err(err).Msg("error setting IsEditing")
			return
		}
		task, ok := m.TaskMap[m.taskList[id]]
		if ok {
			err = m.BindTaskEditSynopsis.Set(task.Synopsis)
			if err != nil {
				log.Err(err).Msg("error binding synopsis")
			}
			err = m.BindTaskEditDescription.Set(task.Description)
			if err != nil {
				log.Err(err).Msg("error binding description")
			}
		}
		m.BtnEditCancel.Enable()
		m.BtnEditSave.Disable()
	}
	m.BtnEditSave = widget.NewButtonWithIcon("Save", theme.ConfirmIcon(), func() {

	})
	m.BtnEditCancel = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		log := m.Log.With().Str("function", "BtnEditCancel").Logger()
		isEditing, err := m.BindIsEditing.Get()
		if err != nil {
			log.Err(err).Msg("error getting IsEditing")
			return
		}
		if !isEditing {
			dialog.NewError(
				errors.New("a task is not being edited; please select a task to edit"),
				m.Window,
			).Show()
			return
		}
		// TODO: check if we are dirty and prompt to save if we are
		err = m.BindIsEditing.Set(false)
		if err != nil {
			log.Err(err).Msg("error setting IsEditing")
		}
		err = m.BindTaskEditSynopsis.Set("")
		if err != nil {
			log.Err(err).Msg("error clearing synopsis")
		}
		err = m.BindTaskEditDescription.Set("")
		if err != nil {
			log.Err(err).Msg("error clearing description")
		}
		m.BtnEditSave.Disable()
		m.BtnEditCancel.Disable()
	})
	// setup layout
	m.Container = container.NewPadded(
		container.NewHSplit(
			m.ListTasks,
			container.NewPadded(container.NewVBox()),
		),
	)
	m.Window.SetCloseIntercept(func() {
		m.Window.Hide()
	})
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
			continue
		}
		m.TaskMap[task.Synopsis] = task
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
