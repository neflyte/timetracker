package windows

import (
	"fyne.io/fyne/v2"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
)

type manageWindowV2 interface {
	windowBase
	Show()
	Hide()
	Close()
}

type manageWindowV2Impl struct {
	fyne.Window
	log        zerolog.Logger
	app        *fyne.App
	container  *fyne.Container
	buttonHBox *fyne.Container
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
	return nil
}
