package timesheet

import (
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	// LastStartedCmd represents the command that prints the last x started tasks
	LastStartedCmd = &cobra.Command{
		Use:   "last-started",
		Short: "Shows the last x started tasks",
		RunE:  doLastStarted,
	}
	taskLimit uint
)

func init() {
	LastStartedCmd.Flags().UintVar(&taskLimit, "limit", 5, "the number of tasks to return; must be greater than zero")
}

func doLastStarted(_ *cobra.Command, _ []string) (err error) {
	tsd := new(models.TimesheetData)
	lastStartedTasks, err := tsd.LastStartedTasks(taskLimit)
	if err != nil {
		return
	}
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "Task ID"},
			{Text: "Synopsis"},
			{Text: "Description"},
		},
	}
	for _, task := range lastStartedTasks {
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(task.ID))},
			{Text: task.Synopsis},
			{Text: task.Description},
		}
		table.Body.Cells = append(table.Body.Cells, rec)
	}
	if len(table.Body.Cells) == 0 {
		fmt.Println(color.WhiteString("There are no timesheets"))
		return nil
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}
