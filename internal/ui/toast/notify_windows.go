package toast

import (
	_ "embed"
	"os"
	"os/exec"
	"path"

	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/icons"
	"github.com/rs/zerolog"
)

//go:embed toast.ps1
var toastPs1 []byte

type Impl struct {
	tempDir    string
	scriptPath string
	iconPath   string
	logger     zerolog.Logger
}

func NewToast() Toast {
	return &Impl{
		logger: packageLogger.With().Str("struct", "Impl").Logger(),
	}
}

func (t *Impl) Notify(title string, description string) error {
	log := logger.GetFuncLogger(t.logger, "Notify")
	err := t.ensureScript()
	if err != nil {
		log.Err(err).
			Msg("unable to write temp files")
		return err
	}
	defer t.Cleanup()
	toastArgs := []string{
		"-NoLogo", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-File", t.scriptPath,
		"-AppId", "Timetracker",
		"-Title", title,
		"-Description", description,
		"-Icon", t.iconPath,
	}
	toastCmd := exec.Command("powershell", toastArgs...)
	err = toastCmd.Run()
	if err != nil {
		log.Err(err).
			Msg("unable to run powershell script")
		return err
	}
	return nil
}

func (t *Impl) ensureScript() error {
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
	t.scriptPath = path.Join(tempDir, "toast.ps1")
	err = os.WriteFile(t.scriptPath, toastPs1, tempFileMode)
	if err != nil {
		log.Err(err).
			Msg("unable to write script to temp directory")
		return err
	}
	log.Debug().
		Str("scriptPath", t.scriptPath).
		Msg("wrote script to temp directory")
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

func (t *Impl) Cleanup() {
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
	if t.scriptPath != "" {
		err := os.Remove(t.scriptPath)
		if err != nil {
			log.Err(err).
				Str("scriptPath", t.scriptPath).
				Msg("unable to remove temporary script file")
		}
		t.scriptPath = ""
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
