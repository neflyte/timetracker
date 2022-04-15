package cmd

import (
	"errors"
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/tray"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/nightlyone/lockfile"
	"github.com/spf13/cobra"
	"os"
	"path"
)

const (
	trayPidfile = "timetracker-tray.pid"
)

var (
	trayCmd = &cobra.Command{
		Use:      "tray",
		Short:    "Start the Timetracker system tray app",
		PreRunE:  preDoTray,
		RunE:     doTray,
		PostRunE: postDoTray,
	}
	trayCmdLockFile     lockfile.Lockfile
	trayCmdLockfilePath string
)

func preDoTray(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("preDoTray")
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = "."
	} else {
		userConfigDir = path.Join(userConfigDir, "timetracker")
	}
	if userConfigDir != "." {
		err = os.MkdirAll(userConfigDir, configDirectoryMode)
		if err != nil {
			log.Err(err).Msgf("error creating directories for pidfile; userConfigDir=%s", userConfigDir)
			return err
		}
	}
	trayCmdLockfilePath = path.Join(userConfigDir, trayPidfile)
	log.Trace().Msgf("trayCmdLockfilePath=%s", trayCmdLockfilePath)
	pidExists, err := utils.CheckPidfile(trayCmdLockfilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, utils.ErrStalePidfile) {
		log.Err(err).Msgf("error checking pidfile %s", trayCmdLockfilePath)
	}
	log.Trace().Msgf("pidExists=%t", pidExists)
	if pidExists {
		log.Error().Msgf("pidfile %s exists and its process is running; exiting", trayCmdLockfilePath)
		return errors.New("another process is already running")
	}
	if errors.Is(err, utils.ErrStalePidfile) {
		log.Debug().Msgf("attempting to remove stale pidfile %s", trayCmdLockfilePath)
		err = os.Remove(trayCmdLockfilePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Err(err).Msgf("error removing existing pidfile %s", trayCmdLockfilePath)
			return errors.New("unable to remove stale pidfile")
		}
	}
	trayCmdLockFile, err = lockfile.New(trayCmdLockfilePath)
	if err != nil {
		log.Err(err).Msgf("error creating pidfile; trayCmdLockfilePath=%s", trayCmdLockfilePath)
		return err
	}
	err = trayCmdLockFile.TryLock()
	if err != nil {
		log.Err(err).Msgf("error locking pidfile; trayCmdLockfilePath=%s", trayCmdLockfilePath)
		return err
	}
	log.Debug().Msgf("locked pidfile %s", trayCmdLockfilePath)
	return nil
}

func postDoTray(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("postDoTray")
	err := trayCmdLockFile.Unlock()
	if err != nil {
		log.Err(err).Msgf("error releasing pidfile %s", trayCmdLockfilePath)
		log.Warn().Msgf("attempting to force-remove pidfile %s", trayCmdLockfilePath)
		fileErr := os.Remove(trayCmdLockfilePath)
		if fileErr != nil {
			log.Err(fileErr).Msgf("error force-removing pidfile %s", trayCmdLockfilePath)
		}
		return err
	}
	log.Debug().Msgf("unlocked pidfile %s", trayCmdLockfilePath)
	return nil
}

func doTray(_ *cobra.Command, _ []string) error {
	// Write the AppVersion to the appstate Map so gui components can access it without a direct binding
	appstate.Map().Store(appstate.KeyAppVersion, AppVersion)
	tray.Run()
	return nil
}
