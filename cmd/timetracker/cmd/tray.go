package cmd

import (
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
	return tray.Run()
}
