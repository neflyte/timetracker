package task

import (
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/errors"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/neflyte/timetracker/lib/ui/cli"
	"github.com/spf13/cobra"
)

var (
	// ListCmd represents the command to list tasks
	ListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List tasks",
		RunE:    listTasks,
	}
	listDeletedTasks = false
)

func init() {
	ListCmd.Flags().BoolVarP(&listDeletedTasks, "deleted", "d", false, "Include deleted tasks")
}

func listTasks(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("listTasks")
	tasks, err := models.NewTask().LoadAll(listDeletedTasks)
	if err != nil {
		cli.PrintAndLogError(log, err, errors.ListTaskError)
		return err
	}
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "ID"},
			{Text: "Synopsis"},
			{Text: "Description"},
			{Text: "Created At"},
			{Text: "Updated At"},
		},
	}
	for _, task := range tasks {
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(task.ID))},
			{Text: task.Synopsis},
			{Text: task.Description},
			{Text: task.CreatedAt.Format(constants.TimestampLayout)},
			{Text: task.UpdatedAt.Format(constants.TimestampLayout)},
		}
		table.Body.Cells = append(table.Body.Cells, rec)
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}
