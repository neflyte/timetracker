package task

import (
	"github.com/neflyte/timetracker/lib/ui/cli"
	"github.com/spf13/cobra"
)

var (
	// StopCmd represents the command that stops the running task
	StopCmd = &cobra.Command{
		Use:     "stop",
		Aliases: []string{"st"},
		Short:   "Stop the running task",
		Args:    cobra.ExactArgs(0),
		RunE:    stopTask,
	}
)

func stopTask(_ *cobra.Command, _ []string) error {
	return cli.StopRunningTimesheet()
}
