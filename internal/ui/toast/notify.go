//go:build !windows
// +build !windows

package toast

import (
	"os"
	"path"

	"github.com/gen2brain/beeep"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/rs/zerolog"
)

type impl struct {
	logger   zerolog.Logger
	tempDir  string
	iconPath string
}

func NewToast() Toast {
	return &impl{
		logger: packageLogger.With().Str("struct", "impl").Logger(),
	}
}

func (t *impl) Notify(title string, description string) error {
	log := logger.GetFuncLogger(t.logger, "Notify")
	err := t.ensureIcon()
	if err != nil {
		log.Err(err).
			Msg("unable to write temp file")
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

func (t *impl) ensureIcon() error {
	log := logger.GetFuncLogger(t.logger, "ensureIcon")
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

func (t *impl) Cleanup() {
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
