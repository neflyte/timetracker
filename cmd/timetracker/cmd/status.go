package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
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
	noColour        = false
)

func init() {
	statusCmd.Flags().BoolVarP(&trailingNewline, "newline", "n", false, "flag to add a trailing newline to the output")
	statusCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "include extra details in the output")
	statusCmd.Flags().BoolVarP(&synopsis, "synopsis", "s", false, "show the running task synopsis")
	statusCmd.Flags().BoolVarP(&noColour, "no-colour", "o", false, "do not include colour output")
}

func status(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("status")
	timesheets, err := models.NewTimesheet().SearchOpen()
	if err != nil {
		log.Err(err).
			Msg("error getting running timesheet")
		statusString := constants.UnicodeHeavyX
		if !noColour {
			statusString = color.RedString(statusString)
		}
		fmt.Print(statusString)
		if verbose {
			errString := err.Error()
			if !noColour {
				errString = color.WhiteString(errString)
			}
			fmt.Print("Error: ", errString)
		}
		return err
	}
	sb := strings.Builder{}
	if len(timesheets) == 0 {
		// No running task
		statusString := constants.UnicodeHeavyCheckmark
		if !noColour {
			statusString = color.GreenString(statusString)
		}
		sb.WriteString(statusString)
	} else {
		// Running task...
		timesheet := timesheets[0]
		statusString := constants.UnicodeClock
		if !noColour {
			statusString = color.YellowString(statusString)
		}
		sb.WriteString(statusString)
		if synopsis || verbose {
			synString := timesheet.Task.Synopsis
			if !noColour {
				synString = color.HiWhiteString(synString)
			}
			sb.WriteString(" " + synString)
		}
		if verbose {
			timeSince := time.Since(timesheet.StartTime).Truncate(time.Second).String()
			if !noColour {
				timeSince = color.HiBlueString(timeSince)
			}
			sb.WriteString(" " + timeSince)
		}
	}
	if trailingNewline {
		sb.WriteString("\n")
	}
	fmt.Print(sb.String())
	return nil
}
