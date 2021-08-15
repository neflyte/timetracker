package cmd

import (
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/tray"
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
		return err
	}
	log.Debug().Msgf("unlocked pidfile %s", trayCmdLockfilePath)
	return nil
}

func doTray(_ *cobra.Command, _ []string) error {
	// Write the appVersion to the appstate Map so gui components can access it without a direct binding
	appstate.Map().Store(appstate.KeyAppVersion, appVersion)
	tray.Run()
	return nil
}
