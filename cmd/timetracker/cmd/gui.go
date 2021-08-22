package cmd

import (
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/ui/gui"
	"github.com/nightlyone/lockfile"
	"github.com/spf13/cobra"
	"os"
	"path"
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
