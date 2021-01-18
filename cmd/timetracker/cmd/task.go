package cmd

import (
	"github.com/neflyte/timetracker/cmd/timetracker/cmd/task"
	"github.com/spf13/cobra"
)

var (
	taskCmd = &cobra.Command{
		Use:   "task",
		Short: "Task operations",
		Long:  "Perform various operations on a task",
	}
)

func init() {
	taskCmd.AddCommand(
		task.CreateCmd,
		task.ListCmd,
		task.UpdateCmd,
		task.DeleteCmd,
		task.StartCmd,
		task.StopCmd,
		task.SearchCmd,
	)
}
