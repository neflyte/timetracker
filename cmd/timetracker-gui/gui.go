package main

import (
	"errors"
	"os"
	"path"

	"github.com/neflyte/timetracker/cmd/timetracker-gui/cmd"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/ui/gui"
	"github.com/neflyte/timetracker/lib/utils"
	"github.com/nightlyone/lockfile"
)

var (
	guiCmdLockfile     lockfile.Lockfile
	guiCmdLockfilePath string
)

func ensureUserHomeDirectory() (string, error) {
	log := logger.GetLogger("ensureUserHomeDirectory")
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = "."
	} else {
		userConfigDir = path.Join(userConfigDir, "timetracker")
	}
	if userConfigDir != "." {
		err = os.MkdirAll(userConfigDir, constants.ConfigDirectoryMode)
		if err != nil {
			log.Err(err).
				Str("userConfigDir", userConfigDir).
				Msg("error creating directories for pidfile")
			return "", err
		}
	}
	return userConfigDir, nil
}

func removeStalePidfile(pidfile string) {
	log := logger.GetLogger("removeStalePidfile")
	log.Debug().
		Str("pidfile", pidfile).
		Msg("attempting to remove stale pidfile")
	err := os.Remove(pidfile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Err(err).
			Str("pidfile", pidfile).
			Msg("error removing existing pidfile")
	}
}

func preDoGUI() error {
	log := logger.GetLogger("preDoGUI")
	userConfigDir, err := ensureUserHomeDirectory()
	if err != nil {
		log.Err(err).
			Msg("error ensuring user home directory exists")
		return err
	}
	guiCmdLockfilePath = path.Join(userConfigDir, guiPidfile)
	log.Trace().
		Msgf("guiCmdLockfilePath=%s", guiCmdLockfilePath)
	pidExists, err := utils.CheckPidfile(guiCmdLockfilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, utils.ErrStalePidfile) {
		log.Err(err).
			Str("pidfile", guiCmdLockfilePath).
			Msg("error checking pidfile %s")
	}
	log.Trace().Msgf("pidExists=%t", pidExists)
	if pidExists {
		log.Error().
			Str("pidfile", guiCmdLockfilePath).
			Msgf("pidfile exists and its process is running; exiting")
		return errors.New("another process is already running")
	}
	if errors.Is(err, utils.ErrStalePidfile) {
		removeStalePidfile(guiCmdLockfilePath)
	}
	guiCmdLockfile, err = lockfile.New(guiCmdLockfilePath)
	if err != nil {
		log.Err(err).
			Str("pidfile", guiCmdLockfilePath).
			Msg("error creating pidfile")
		return err
	}
	err = guiCmdLockfile.TryLock()
	if err != nil {
		log.Err(err).
			Str("pidfile", guiCmdLockfilePath).
			Msg("error locking pidfile")
		return err
	}
	log.Debug().
		Str("pidfile", guiCmdLockfilePath).
		Msg("locked pidfile")
	return nil
}

func postDoGUI() error {
	log := logger.GetLogger("postDoGUI")
	err := guiCmdLockfile.Unlock()
	if err != nil {
		log.Err(err).
			Str("pidfile", guiCmdLockfilePath).
			Msg("error releasing pidfile")
		log.Warn().
			Str("pidfile", guiCmdLockfilePath).
			Msg("attempting to force-remove pidfile")
		fileErr := os.Remove(guiCmdLockfilePath)
		if fileErr != nil {
			log.Err(fileErr).
				Str("pidfile", guiCmdLockfilePath).
				Msg("error force-removing pidfile")
		}
		return err
	}
	log.Debug().
		Str("pidfile", guiCmdLockfilePath).
		Msg("unlocked pidfile")
	return nil
}

func doGUI() {
	app := gui.InitGUI(cmd.AppVersion)
	switch {
	case guiCmdOptionStopRunningTask:
		gui.ShowTimetrackerWindowAndStopRunningTask()
	case guiCmdOptionShowManageWindow:
		gui.ShowTimetrackerWindowWithManageWindow()
	case guiCmdOptionShowAboutWindow:
		gui.ShowTimetrackerWindowWithAbout()
	case guiCmdOptionShowCreateAndStartDialog:
		gui.ShowTimetrackerWindowAndShowCreateAndStartDialog()
	default:
		gui.ShowTimetrackerWindow()
	}
	// Start the GUI
	gui.StartGUI(app)
}
