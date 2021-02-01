package cmd

import (
	"fmt"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/utils"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"time"
)

var (
	statusCmd = &cobra.Command{
		Use:     "status",
		Aliases: []string{"s"},
		Short:   "Active task status",
		Long:    "Shows an optionally-verbose status of the active task, if any",
		Args:    cobra.ExactArgs(0),
		RunE:    status,
	}
	trailingNewline = false
	verbose         = false
	synopsis        = false
)

func init() {
	statusCmd.Flags().BoolVarP(&trailingNewline, "newline", "n", false, "flag to add a trailing newline to the output")
	statusCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "include extra details in the output")
	statusCmd.Flags().BoolVarP(&synopsis, "synopsis", "s", false, "show the running task synopsis")
}

func status(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("status")
	defer func() {
		if trailingNewline {
			fmt.Println()
		}
	}()
	timesheet, err := utils.GetRunningTimesheet()
	if err != nil {
		log.Err(err).Msg("error getting running timesheet")
		fmt.Print(chalk.Red, constants.UnicodeHeavyX)
		if verbose {
			fmt.Print("Error:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		}
		return err
	}
	if timesheet == nil {
		// No running task
		fmt.Print(chalk.Green, constants.UnicodeHeavyCheckmark)
	} else {
		// Running task...
		fmt.Print(chalk.Yellow, constants.UnicodeClock)
		if synopsis || verbose {
			fmt.Print(" ", chalk.White, timesheet.Task.Synopsis)
		}
		if verbose {
			fmt.Print(" ", chalk.Blue, time.Since(timesheet.StartTime).Truncate(time.Second).String())
		}
	}
	return nil
}
