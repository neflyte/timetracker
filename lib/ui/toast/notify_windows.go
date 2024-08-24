package toast

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/ui/icons"
	"github.com/rs/zerolog"
)

//go:embed toast.ps1
var toastPs1 []byte

type impl struct {
	tempDir    string
	scriptPath string
	iconPath   string
	nonce      int
	logger     zerolog.Logger
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
	err = t.ensureScript()
	if err != nil {
		log.Err(err).
			Msg("unable to write script file")
	}
	err = t.ensureIcon()
	if err != nil {
		log.Err(err).
			Msg("unable to write icon file")
	}
	defer func() {
		// Sleep for 1 second before cleaning up to give the OS a chance to use the resources
		<-time.After(time.Second)
		t.Cleanup()
	}()
	// construct powershell path using environment variable data
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemDrive := os.Getenv("SystemDrive")
		if systemDrive == "" {
			systemRoot = "C:\\Windows"
		} else {
			systemRoot = fmt.Sprintf("%s\\Windows", systemDrive)
		}
	}
	powershellPath := fmt.Sprintf("%s\\%s", systemRoot, DefaultWindowsPowershellPath)
	toastArgs := []string{
		"-ExecutionPolicy", "Bypass",
		"-NoLogo", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-File", t.scriptPath,
		"-AppId", "Timetracker",
		"-Title", title,
		"-Description", description,
		"-Icon", t.iconPath,
	}
	log.Debug().
		Strs("args", toastArgs).
		Str("path", powershellPath).
		Msg("powershell args")
	var out, stderr strings.Builder
	toastCmd := exec.Command(powershellPath, toastArgs...)
	toastCmd.Stdout = &out
	toastCmd.Stderr = &stderr
	toastCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err = toastCmd.Run()
	if err != nil {
		log.Err(err).
			Str("script", t.scriptPath).
			Str("stderr", stderr.String()).
			Str("stdout", out.String()).
			Msg("unable to run powershell script")
		return err
	}
	log.Debug().
		Str("stdout", out.String()).
		Msg("successfully ran powershell script")
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
	}
	err := os.WriteFile(t.iconPath, icons.IconV2.StaticContent, tempFileMode)
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

func (t *impl) ensureScript() error {
	log := logger.GetFuncLogger(t.logger, "ensureScript")
	if t.scriptPath == "" {
		t.scriptPath = path.Join(t.tempDir, fmt.Sprintf("toast-%d.ps1", t.nonce))
		t.nonce++
	}
	err := os.WriteFile(t.scriptPath, toastPs1, tempFileMode)
	if err != nil {
		log.Err(err).
			Msg("unable to write script to temp directory")
		return err
	}
	log.Debug().
		Str("scriptPath", t.scriptPath).
		Msg("wrote script to temp directory")
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
