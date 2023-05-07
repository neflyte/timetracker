//go:build !windows
// +build !windows

package toast

type ToastImpl struct {
	logger   zerolog.Logger
	tempDir  string
	iconPath string
}

func NewToast() Toast {
	return &ToastImpl{
		logger: packageLogger.With().Str("struct", "ToastImpl").Logger(),
	}
}

func (t *ToastImpl) Notify(title string, description string) error {
	log := logger.GetFuncLogger(t.logger, "Notify")
	err := t.ensureScript()
	if err != nil {
		log.Err(err).
			Msg("unable to write temp files")
		return err
	}
	defer t.Cleanup()
	err = beeep.Alert(title, description, t.iconPath)
	if err != nil {
		log.Err(err).
			Msg("unable to send notification")
		return err
	}
	return nil
}

func (t *ToastImpl) ensureScript() error {
	log := logger.GetFuncLogger(t.logger, "ensureScript")
	tempDir, err := os.MkdirTemp("", "timetracker-toast")
	if err != nil {
		log.Err(err).
			Msg("unable to create temp directory")
		return err
	}
	log.Debug().
		Str("tempDir", tempDir).
		Msg("created temp directory")
	t.tempDir = tempDir
	t.iconPath = path.Join(tempDir, "icon-v2.ico")
	err = os.WriteFile(t.iconPath, icons.IconV2.StaticContent, tempFileMode)
	if err != nil {
		log.Err(err).
			Msg("unable to write icon to temp directory")
		return err
	}
	log.Debug().
		Str("iconPath", t.iconPath).
		Msg("wrote icon to temp directory")
	return nil
}

func (t *ToastImpl) Cleanup() {
	log := logger.GetFuncLogger(t.logger, "Cleanup")
	if t.iconPath != "" {
		err := os.Remove(t.iconPath)
		if err != nil {
			log.Err(err).
				Str("iconPath", t.iconPath).
				Msg("unable to remove temporary icon file")
		}
		t.iconPath = ""
	}
	if t.tempDir != "" {
		err := os.RemoveAll(t.tempDir)
		if err != nil {
			log.Err(err).
				Str("tempDir", t.tempDir).
				Msg("unable to remove temporary directory")
		}
		t.tempDir = ""
	}
}
