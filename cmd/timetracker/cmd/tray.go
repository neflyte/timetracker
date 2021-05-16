package cmd

import (
	"github.com/neflyte/timetracker/internal/appstate"
	"github.com/neflyte/timetracker/internal/ui/tray"
	"github.com/spf13/cobra"
)

var (
	trayCmd = &cobra.Command{
		Use:   "tray",
		Short: "Start the Timetracker system tray app",
		RunE:  doTray,
	}
)

func doTray(_ *cobra.Command, _ []string) error {
	// Write the AppVersion to the appstate Map so gui components can access it without a direct binding
	appstate.Map().Store(appstate.KeyAppVersion, AppVersion)
	return tray.Run()
}
