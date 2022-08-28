package cmd

import (
	"errors"
	"os"
	"path"

	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/gui"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/nightlyone/lockfile"
	"github.com/spf13/cobra"
)

const (
	guiPidfile = "timetracker-gui.pid"
)

var (
	guiCmd = &cobra.Command{
		Use:      "gui",
		Short:    "Start the Timetracker GUI app",
		PreRunE:  preDoGUI,
		RunE:     doGUI,
		PostRunE: postDoGUI,
		Args:     cobra.ExactArgs(0),
	}
	guiCmdLockfile                       lockfile.Lockfile
	guiCmdLockfilePath                   string
	guiCmdOptionStopRunningTask          *bool
	guiCmdOptionShowCreateAndStartDialog *bool
	guiCmdOptionShowManageWindow         *bool
	guiCmdOptionShowAboutWindow          *bool
)

func init() {
	guiCmdOptionStopRunningTask = guiCmd.Flags().Bool("stop-running-task", false, "Stops the running task, if any")
	guiCmdOptionShowCreateAndStartDialog = guiCmd.Flags().Bool("create-and-start", false, "Shows the Create and Start New Task dialog")
	guiCmdOptionShowManageWindow = guiCmd.Flags().Bool("manage", false, "Shows the Manage Window")
	guiCmdOptionShowAboutWindow = guiCmd.Flags().Bool("about", false, "Shows the About Window")
}

func preDoGUI(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("preDoGUI")
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
	guiCmdLockfilePath = path.Join(userConfigDir, guiPidfile)
	log.Trace().Msgf("guiCmdLockfilePath=%s", guiCmdLockfilePath)
	pidExists, err := utils.CheckPidfile(guiCmdLockfilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, utils.ErrStalePidfile) {
		log.Err(err).Msgf("error checking pidfile %s", guiCmdLockfilePath)
	}
	log.Trace().Msgf("pidExists=%t", pidExists)
	if pidExists {
		log.Error().Msgf("pidfile %s exists and its process is running; exiting", guiCmdLockfilePath)
		return errors.New("another process is already running")
	}
	if errors.Is(err, utils.ErrStalePidfile) {
		log.Debug().Msgf("attempting to remove stale pidfile %s", guiCmdLockfilePath)
		err = os.Remove(guiCmdLockfilePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Err(err).Msgf("error removing existing pidfile %s", guiCmdLockfilePath)
			return errors.New("unable to remove stale pidfile")
		}
	}
	guiCmdLockfile, err = lockfile.New(guiCmdLockfilePath)
	if err != nil {
		log.Err(err).Msgf("error creating pidfile; pidPath=%s", guiCmdLockfilePath)
		return err
	}
	err = guiCmdLockfile.TryLock()
	if err != nil {
		log.Err(err).Msgf("error locking pidfile; pidPath=%s", guiCmdLockfilePath)
		return err
	}
	log.Debug().Msgf("locked pidfile %s", guiCmdLockfilePath)
	// Write the AppVersion to the appstate Map so gui components can access it without a direct binding
	appstate.Map().Store(appstate.KeyAppVersion, AppVersion)
	return nil
}

func postDoGUI(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("postDoGUI")
	err := guiCmdLockfile.Unlock()
	if err != nil {
		log.Err(err).Msgf("error releasing pidfile %s", guiCmdLockfilePath)
		log.Warn().Msgf("attempting to force-remove pidfile %s", guiCmdLockfilePath)
		fileErr := os.Remove(guiCmdLockfilePath)
		if fileErr != nil {
			log.Err(fileErr).Msgf("error force-removing pidfile %s", guiCmdLockfilePath)
		}
		return err
	}
	log.Debug().Msgf("unlocked pidfile %s", guiCmdLockfilePath)
	return nil
}

func doGUI(_ *cobra.Command, _ []string) error {
	app := gui.InitGUI()
	// Default to showing the main timetracker window
	gui.ShowTimetrackerWindow()
	if *guiCmdOptionStopRunningTask {
		gui.ShowTimetrackerWindowAndStopRunningTask()
	}
	if *guiCmdOptionShowManageWindow {
		gui.ShowTimetrackerWindowWithManageWindow()
	}
	if *guiCmdOptionShowAboutWindow {
		gui.ShowTimetrackerWindowWithAbout()
	}
	if *guiCmdOptionShowCreateAndStartDialog {
		gui.ShowTimetrackerWindowAndShowCreateAndStartDialog()
	}
	// Start the GUI
	gui.StartGUI(app)
	return nil
}
