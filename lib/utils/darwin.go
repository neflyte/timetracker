package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

const (
	launchdServiceName = "cc.ethereal.timetracker"
)

func generateLaunchdPlist(appPath string) (string, error) {
	if appPath == "" {
		return "", errors.New("appPath is empty")
	}
	plistMap := map[string]any{
		"Label": launchdServiceName,
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
	log := utilsLogger.With().Str("func", "writePlist").Logger()
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
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s.*.plist", launchdServiceName))
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

// EnableLaunchd sets Timetracker Tray to start at login
func EnableLaunchd() error {
	log := utilsLogger.With().Str("func", "EnableLaunchd").Logger()
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

// DisableLaunchd stops Timetracker Tray from starting at login
func DisableLaunchd() error {
	log := utilsLogger.With().Str("func", "DisableLaunchd").Logger()
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	launchdService := fmt.Sprintf("gui/%s/%s", currentUser.Uid, launchdServiceName)
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
