package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
)

type ManageWindow interface {
	Show()
	Hide()
	Close()
	Get() manageWindow
}

type manageWindow struct {
	App       *fyne.App
	Window    fyne.Window
	Container *fyne.Container
	ListTasks *widget.List
	Log       zerolog.Logger

	BindTaskList binding.ExternalStringList

	taskList []string
}

func NewManageWindow(app fyne.App) ManageWindow {
	mw := &manageWindow{
		App:    &app,
		Window: app.NewWindow("Manage Tasks"),
		Log:    logger.GetStructLogger("ManageWindow"),
	}
	mw.init()
	return mw
}

func (m *manageWindow) init() {
	m.taskList = make([]string, 0)
	m.BindTaskList = binding.BindStringList(&m.taskList)
	m.ListTasks = widget.NewListWithData(
		m.BindTaskList,
		func() fyne.CanvasObject {
			return widget.NewCard("new task", "task description", nil)
		},
		func(item binding.DataItem, canvasObject fyne.CanvasObject) {
			taskStringBinding := item.(binding.String)
			taskString, err := taskStringBinding.Get()
			if err != nil {
				// TODO: log this error
				return
			}
			taskCard, ok := canvasObject.(*fyne.Container).Objects[0].(*widget.Card)
			if ok {
				taskCard.SetTitle(taskString)
			}
		},
	)
	m.Container = container.NewPadded(
		container.NewHSplit(
			m.ListTasks,
			container.NewPadded(container.NewVBox()),
		),
	)
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
