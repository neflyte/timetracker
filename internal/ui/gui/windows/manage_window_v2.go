package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/gui/widgets"
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
	m.createButton = widget.NewButton("New", func() {})
	m.editButton = widget.NewButton("Edit", func() {})
	m.deleteButton = widget.NewButton("Delete", func() {})
	m.buttonHBox = container.NewHBox(m.createButton, widget.NewSeparator(), m.editButton, m.deleteButton)
	m.taskSelector = widgets.NewTaskSelector()
	m.taskSelector.Observable().ForEach(
		func(item interface{}) {

		},
		func(err error) {

		},
		func() {

		},
	)
	m.container = container.NewVBox(m.buttonHBox, m.taskSelector)
	m.Window.SetCloseIntercept(m.Hide)
	m.Window.SetContent(m.container)
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
}
