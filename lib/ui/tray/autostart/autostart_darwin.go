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
				Msg("Error removing launchd plist file")
		}
	}()
	currentUser, err := user.Current()
	if err != nil {
		log.Err(err).
			Msg("Error getting current user ID")
		return err
	}
	launchdDomain := fmt.Sprintf("gui/%s/", currentUser.Uid)
	stdout, stderr, err := runLaunchctl("bootstrap", []string{launchdDomain, plistFile})
	if err != nil {
		log.Err(err).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Msg("error bootstrapping plist")
		return err
	}
	launchdService := fmt.Sprintf("%s/%s", launchdDomain, constants.AppID)
	stdout, stderr, err = runLaunchctl("enable", []string{launchdService})
	if err != nil {
		log.Err(err).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Msg("error enabling service")
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
	stdout, stderr, err := runLaunchctl("disable", []string{launchdService})
	if err != nil {
		log.Err(err).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Msg("error disabling service")
		return err
	}
	stdout, stderr, err = runLaunchctl("remove", []string{constants.AppID})
	if err != nil {
		log.Err(err).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Msg("error removing service")
		return err
	}
	return nil
}

// IsEnabled determines if tray autostart is enabled
func IsEnabled() bool {
	log := logger.GetFuncLogger(packageLogger, "IsEnabled")
	stdout, stderr, err := runLaunchctl("list", []string{constants.AppID})
	if err != nil {
		log.Err(err).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Msg("error checking for service")
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

func runLaunchctl(command string, args []string) (string, string, error) {
	log := logger.GetFuncLogger(packageLogger, "runLaunchctl")
	var stdout, stderr strings.Builder
	launchctlArgs := []string{command}
	launchctlArgs = append(launchctlArgs, args...)
	launchctlCommand := exec.Command("launchctl", launchctlArgs...)
	launchctlCommand.Stdout = &stdout
	launchctlCommand.Stderr = &stderr
	log.Debug().
		Str("command", command).
		Strs("args", launchctlArgs).
		Msg("run launchctl")
	err := launchctlCommand.Run()
	log.Debug().
		Str("stdout", stdout.String()).
		Str("stderr", stderr.String()).
		Msg("launchctl command output")
	if err != nil {
		log.Err(err).
			Str("command", command).
			Strs("args", launchctlArgs).
			Msg("error running launchctl")
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}
