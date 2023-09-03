//go:build !windows
// +build !windows

package toast

import (
	"os"
	"path"

	"github.com/gen2brain/beeep"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/ui/icons"
	"github.com/rs/zerolog"
)

type impl struct {
	logger   zerolog.Logger
	tempDir  string
	iconPath string
}

func NewToast() Toast {
	t := &impl{
		logger: packageLogger.With().Str("struct", "impl").Logger(),
	}
	err := t.ensureIcon()
	if err != nil {
		t.logger.Err(err).
			Msg("unable to write temp file")
	}
	return t
}

func (t *impl) Notify(title string, description string) error {
	log := logger.GetFuncLogger(t.logger, "Notify")
	err := t.ensureIcon()
	if err != nil {
		log.Err(err).
			Msg("unable to write temp file")
	}
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
	if t.tempDir == "" {
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
	}
	if t.iconPath == "" {
		t.iconPath = path.Join(t.tempDir, "icon-v2.ico")
		if _, err := os.Stat(t.iconPath); err != nil {
			log.Debug().
				Err(err).
				Str("iconPath", t.iconPath).
				Msg("got error running stat()")
			err = os.WriteFile(t.iconPath, icons.IconV2.StaticContent, tempFileMode)
			if err != nil {
				log.Err(err).
					Msg("unable to write icon to temp directory")
				return err
			}
			log.Debug().
				Str("iconPath", t.iconPath).
				Msg("wrote icon to temp directory")
		}
	}
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
