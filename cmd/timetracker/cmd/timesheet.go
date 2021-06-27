package cmd

import (
	"github.com/neflyte/timetracker/cmd/timetracker/cmd/timesheet"
	"github.com/spf13/cobra"
)

var (
	timesheetCmd = &cobra.Command{
		Use:     "timesheet",
		Aliases: []string{"ts"},
		Short:   "Timesheet operations",
		Long:    "Report on task times",
	}
)

func init() {
	timesheetCmd.AddCommand(
		timesheet.DumpCmd,
		timesheet.LastStartedCmd,
		timesheet.ReportCmd,
	)
}
