package autostart

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
)

const (
	tempFileMode = 0644
)

//go:embed autostart.ps1
var autostartPs1 []byte

var (
	tempDir    string
	scriptPath string
	nonce      = 1
)

// Enable enables tray autostart
func Enable() error {
	log := logger.GetFuncLogger(packageLogger, "Enable")
	err := ensureTempDirectory()
	if err != nil {
		log.Err(err).
			Msg("unable to create temp directory")
		return err
	}
	err = ensureScript()
	if err != nil {
		log.Err(err).
			Msg("unable to write script")
		return err
	}
	defer cleanup()
	// Get the path to the tray
	trayPath, err := os.Executable()
	if err != nil {
		return err
	}
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
	powershellPath := fmt.Sprintf("%s\\%s", systemRoot, constants.DefaultWindowsPowershellPath)
	autostartArgs := []string{
		"-ExecutionPolicy", "Bypass",
		"-NoLogo", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-File", scriptPath,
		"-Enable", "-TrayPath", trayPath,
	}
	var out, stderr strings.Builder
	autostartCmd := exec.Command(powershellPath, autostartArgs...)
	autostartCmd.Stdout = &out
	autostartCmd.Stderr = &stderr
	autostartCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err = autostartCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Disable disables tray autostart
func Disable() error {
	log := logger.GetFuncLogger(packageLogger, "Disable")
	err := ensureTempDirectory()
	if err != nil {
		log.Err(err).
			Msg("unable to create temp directory")
		return err
	}
	err = ensureScript()
	if err != nil {
		log.Err(err).
			Msg("unable to write script")
		return err
	}
	defer cleanup()
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
	powershellPath := fmt.Sprintf("%s\\%s", systemRoot, constants.DefaultWindowsPowershellPath)
	autostartArgs := []string{
		"-ExecutionPolicy", "Bypass",
		"-NoLogo", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-File", scriptPath,
		"-Disable",
	}
	var out, stderr strings.Builder
	autostartCmd := exec.Command(powershellPath, autostartArgs...)
	autostartCmd.Stdout = &out
	autostartCmd.Stderr = &stderr
	autostartCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err = autostartCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// IsEnabled determines if tray autostart is enabled
func IsEnabled() bool {
	log := logger.GetFuncLogger(packageLogger, "IsEnabled")
	err := ensureTempDirectory()
	if err != nil {
		log.Err(err).
			Msg("unable to create temp directory")
		return false
	}
	err = ensureScript()
	if err != nil {
		log.Err(err).
			Msg("unable to write script")
		return false
	}
	defer cleanup()
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
	powershellPath := fmt.Sprintf("%s\\%s", systemRoot, constants.DefaultWindowsPowershellPath)
	autostartArgs := []string{
		"-ExecutionPolicy", "Bypass",
		"-NoLogo", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-File", scriptPath,
		"-IsEnabled",
	}
	var out, stderr strings.Builder
	autostartCmd := exec.Command(powershellPath, autostartArgs...)
	autostartCmd.Stdout = &out
	autostartCmd.Stderr = &stderr
	autostartCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err = autostartCmd.Run()
	if err != nil {
		if errors.Is(err, new(exec.ExitError)) {
			return false
		}
		log.Err(err).
			Msg("unable to check if autostart is enabled")
		return false
	}
	return true
}

func ensureTempDirectory() error {
	if tempDir == "" {
		newTmpdir, err := os.MkdirTemp("", "timetracker-tray")
		if err != nil {
			return err
		}
		tempDir = newTmpdir
	}
	return nil
}

func ensureScript() error {
	if scriptPath == "" {
		scriptPath = path.Join(tempDir, fmt.Sprintf("autostart-%d.ps1", nonce))
		nonce++
	}
	err := os.WriteFile(scriptPath, autostartPs1, tempFileMode)
	if err != nil {
		scriptPath = ""
		return err
	}
	return nil
}

func cleanup() {
	log := logger.GetFuncLogger(packageLogger, "cleanup")
	if scriptPath != "" {
		err := os.Remove(scriptPath)
		if err != nil {
			log.Err(err).
				Str("scriptPath", scriptPath).
				Msg("unable to remove temporary script file")
		}
		scriptPath = ""
	}
	if tempDir != "" {
		err := os.RemoveAll(tempDir)
		if err != nil {
			log.Err(err).
				Str("tempDir", tempDir).
				Msg("unable to remove temporary directory")
		}
		tempDir = ""
	}
}
