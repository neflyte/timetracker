//go:build !windows

package toast

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/ui/icons"
	"github.com/rs/zerolog"
)

type impl struct {
	logger   zerolog.Logger
	tempDir  string
	iconPath string
	nonce    int
}

// NewToast creates a new instance of the Toast interface
func NewToast() Toast {
	return &impl{
		logger: packageLogger.With().Str("struct", "impl").Logger(),
		nonce:  1,
	}
}

// Notify sends a notification
func (t *impl) Notify(title string, description string) error {
	log := logger.GetFuncLogger(t.logger, "Notify")
	err := t.ensureTempDirectory()
	if err != nil {
		log.Err(err).
			Msg("unable to create temp directory")
	}
	err = t.ensureIcon()
	if err != nil {
		log.Err(err).
			Msg("unable to write temp icon file")
	}
	err = beeep.Alert(title, description, t.iconPath)
	defer func() {
		<-time.After(time.Second)
		t.Cleanup()
	}()
	if err != nil {
		log.Err(err).
			Msg("unable to send notification")
		return err
	}
	return nil
}

func (t *impl) ensureTempDirectory() error {
	log := logger.GetFuncLogger(t.logger, "ensureTempDirectory")
	if t.tempDir == "" {
		tempDir, err := os.MkdirTemp("", "timetracker-toast")
		if err != nil {
			log.Err(err).
				Msg("unable to create temp directory")
			return err
		}
		t.tempDir = tempDir
		log.Debug().
			Str("tempDir", tempDir).
			Msg("created temp directory")
	}
	return nil
}

func (t *impl) ensureIcon() error {
	log := logger.GetFuncLogger(t.logger, "ensureIcon")
	if t.iconPath == "" {
		t.iconPath = path.Join(t.tempDir, fmt.Sprintf("icon-v2-%d.ico", t.nonce))
		t.nonce++
		err := os.WriteFile(t.iconPath, icons.IconV2.StaticContent, tempFileMode)
		if err != nil {
			log.Err(err).
				Msg("unable to write icon to temp directory")
			t.iconPath = ""
			return err
		}
		log.Debug().
			Str("iconPath", t.iconPath).
			Msg("wrote icon to temp directory")
	}
	return nil
}

// Cleanup cleans up temporary files and directories
func (t *impl) Cleanup() {
	log := logger.GetFuncLogger(t.logger, "Cleanup")
	if t.iconPath != "" {
		err := os.Remove(t.iconPath)
		if err != nil {
			log.Err(err).
				Str("iconPath", t.iconPath).
				Msg("unable to remove temporary icon file")
		}
		log.Debug().
			Str("iconPath", t.iconPath).
			Msg("removed temporary icon file")
		t.iconPath = ""
	}
	if t.tempDir != "" {
		err := os.RemoveAll(t.tempDir)
		if err != nil {
			log.Err(err).
				Str("tempDir", t.tempDir).
				Msg("unable to remove temporary directory")
		}
		log.Debug().
			Str("tempDir", t.tempDir).
			Msg("removed temporary directory")
		t.tempDir = ""
	}
}
