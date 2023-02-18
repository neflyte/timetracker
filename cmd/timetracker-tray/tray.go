package main

import (
	"errors"
	"os"
	"path"

	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/tray"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/nightlyone/lockfile"
)

var (
	trayCmdLockFile     lockfile.Lockfile
	trayCmdLockfilePath string
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

func preDoTray() error {
	log := logger.GetLogger("preDoTray")
	userConfigDir, err := ensureUserHomeDirectory()
	if err != nil {
		log.Err(err).
			Msg("error ensuring user home directory exists")
		return err
	}
	trayCmdLockfilePath = path.Join(userConfigDir, trayPidfile)
	log.Trace().Msgf("trayCmdLockfilePath=%s", trayCmdLockfilePath)
	pidExists, err := utils.CheckPidfile(trayCmdLockfilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, utils.ErrStalePidfile) {
		log.Err(err).
			Str("pidfile", trayCmdLockfilePath).
			Msg("error checking pidfile")
	}
	log.Trace().Msgf("pidExists=%t", pidExists)
	if pidExists {
		log.Error().
			Str("pidfile", trayCmdLockfilePath).
			Msg("pidfile exists and its process is running; exiting")
		return errors.New("another process is already running")
	}
	if errors.Is(err, utils.ErrStalePidfile) {
		removeStalePidfile(trayCmdLockfilePath)
	}
	trayCmdLockFile, err = lockfile.New(trayCmdLockfilePath)
	if err != nil {
		log.Err(err).
			Str("pidfile", trayCmdLockfilePath).
			Msg("error creating pidfile")
		return err
	}
	err = trayCmdLockFile.TryLock()
	if err != nil {
		log.Err(err).
			Str("pidfile", trayCmdLockfilePath).
			Msg("error locking pidfile")
		return err
	}
	log.Debug().
		Str("pidfile", trayCmdLockfilePath).
		Msg("locked pidfile")
	return nil
}

func postDoTray() error {
	log := logger.GetLogger("postDoTray")
	log.Debug().
		Msg("called")
	err := trayCmdLockFile.Unlock()
	if err != nil {
		log.Err(err).
			Str("pidfile", trayCmdLockfilePath).
			Msg("error releasing pidfile")
		log.Warn().
			Str("pidfile", trayCmdLockfilePath).
			Msg("attempting to force-remove pidfile")
		fileErr := os.Remove(trayCmdLockfilePath)
		if fileErr != nil {
			log.Err(fileErr).
				Str("pidfile", trayCmdLockfilePath).
				Msg("error force-removing pidfile")
		}
		return err
	}
	log.Debug().
		Str("pidfile", trayCmdLockfilePath).
		Msg("unlocked pidfile")
	return nil
}

func doTray() {
	// Start the tray
	tray.Run(nil)
}
