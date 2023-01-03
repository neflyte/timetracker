package cmd

import (
	"errors"
	"github.com/neflyte/timetracker/internal/ui/tray"
	"os"
	"path"

	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/nightlyone/lockfile"
	"github.com/spf13/cobra"
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

func postDoTray(_ *cobra.Command, _ []string) error {
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

func doTray(_ *cobra.Command, _ []string) error {
	// log := logger.GetLogger("doTray")
	// Write the AppVersion to the appstate Map so gui components can access it without a direct binding
	appstate.Map().Store(appstate.KeyAppVersion, AppVersion)
	// Start the tray
	//tray.Run(func() {
	//	log.Debug().
	//		Msg("calling postDoTray from tray.Run func")
	//	err := postDoTray(nil, nil)
	//	if err != nil {
	//		log.Err(err).
	//			Msg("error running postDoTray")
	//	}
	//})
	tray.Run(nil)
	return nil
}
