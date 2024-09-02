package autostart

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"howett.net/plist"
)

// Enable enables tray autostart
func Enable() error {
	log := logger.GetFuncLogger(packageLogger, "Enable")
	plistFile, err := writePlist()
	if err != nil {
		return err
	}
	defer func() {
		err = os.Remove(plistFile)
		if err != nil {
			log.Err(err).
				Msg("Error removing launchd plist")
		}
	}()
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	launchdDomain := fmt.Sprintf("gui/%s/", currentUser.Uid)
	var stdout, stderr strings.Builder
	launchctlArgs := []string{"bootstrap", launchdDomain, plistFile}
	launchctlCommand := exec.Command("launchctl", launchctlArgs...)
	launchctlCommand.Stdout = &stdout
	launchctlCommand.Stderr = &stderr
	err = launchctlCommand.Run()
	log.Debug().
		Str("stdout", stdout.String()).
		Str("stderr", stderr.String()).
		Msg("command output")
	if err != nil {
		log.Err(err).
			Strs("launchctlArgs", launchctlArgs).
			Msg("error running launchctl")
		return err
	}
	return nil
}

// Disable disables tray autostart
func Disable() error {
	log := logger.GetFuncLogger(packageLogger, "Disable")
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	launchdService := fmt.Sprintf("gui/%s/%s", currentUser.Uid, constants.AppID)
	var stdout, stderr strings.Builder
	launchctlArgs := []string{"bootout", launchdService}
	launchctlCommand := exec.Command("launchctl", launchctlArgs...)
	launchctlCommand.Stdout = &stdout
	launchctlCommand.Stderr = &stderr
	err = launchctlCommand.Run()
	log.Debug().
		Str("stdout", stdout.String()).
		Str("stderr", stderr.String()).
		Msg("command output")
	if err != nil {
		log.Err(err).
			Strs("launchctlArgs", launchctlArgs).
			Msg("error running launchctl")
		return err
	}
	return nil
}

// IsEnabled determines if tray autostart is enabled
func IsEnabled() bool {
	log := logger.GetFuncLogger(packageLogger, "IsEnabled")
	var stdout, stderr strings.Builder
	launchctlArgs := []string{"list", constants.AppID}
	launchctlCommand := exec.Command("launchctl", launchctlArgs...)
	launchctlCommand.Stdout = &stdout
	launchctlCommand.Stderr = &stderr
	err := launchctlCommand.Run()
	log.Debug().
		Str("stdout", stdout.String()).
		Str("stderr", stderr.String()).
		Msg("command output")
	if err != nil {
		log.Err(err).
			Strs("launchctlArgs", launchctlArgs).
			Msg("error running launchctl")
		return false
	}
	return true
}

func generateLaunchdPlist(appPath string) (string, error) {
	if appPath == "" {
		return "", errors.New("appPath is empty")
	}
	plistMap := map[string]any{
		"Label": constants.AppID,
		"ProgramArguments": []string{
			fmt.Sprintf("%s/timetracker-tray", appPath),
		},
		"ProcessType": "Interactive",
		"RunAtLoad":   true,
		"KeepAlive":   false,
	}
	plistBytes, err := plist.Marshal(&plistMap, plist.XMLFormat)
	if err != nil {
		return "", err
	}
	return string(plistBytes), nil
}

func writePlist() (string, error) {
	log := logger.GetFuncLogger(packageLogger, "writePlist")
	trayPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	trayBaseDir := filepath.Dir(trayPath)
	plistData, err := generateLaunchdPlist(trayBaseDir)
	if err != nil {
		return "", err
	}
	log.Debug().
		Str("plistData", plistData).
		Msg("plist data")
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s.*.plist", constants.AppID))
	if err != nil {
		return "", err
	}
	log.Debug().
		Str("tempFile", tempFile.Name()).
		Msg("created tempfile")
	defer func() {
		err = tempFile.Close()
		if err != nil {
			log.Err(err).
				Msg("Error closing temporary file")
		}
	}()
	_, err = tempFile.WriteString(plistData)
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}
