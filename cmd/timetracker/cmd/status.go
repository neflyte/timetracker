package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
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
	timesheets, err := models.Timesheet(new(models.TimesheetData)).SearchOpen()
	if err != nil {
		log.Err(err).Msg("error getting running timesheet")
		fmt.Print(color.RedString(constants.UnicodeHeavyX))
		if verbose {
			fmt.Print("Error:", color.WhiteString(err.Error()))
		}
		return err
	}
	if len(timesheets) == 0 {
		// No running task
		fmt.Print(color.GreenString(constants.UnicodeHeavyCheckmark))
	} else {
		// Running task...
		timesheet := timesheets[0]
		fmt.Print(color.YellowString(constants.UnicodeClock))
		if synopsis || verbose {
			fmt.Print(" ", color.WhiteString(timesheet.Task.Synopsis))
		}
		if verbose {
			fmt.Print(" ", color.HiBlueString(time.Since(timesheet.StartTime).Truncate(time.Second).String()))
		}
	}
	return nil
}
