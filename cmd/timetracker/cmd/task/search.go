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
	// SearchCmd represents the command that searches for tasks
	SearchCmd = &cobra.Command{
		Use:     "search [search terms]",
		Aliases: []string{"find"},
		Short:   "Search for tasks",
		Args:    cobra.ExactArgs(1),
		RunE:    searchTask,
	}
)

func searchTask(_ *cobra.Command, args []string) error {
	log := logger.GetLogger("searchTask")
	tasks, err := models.NewTask().Search(args[0])
	if err != nil {
		cli.PrintAndLogError(log, err, errors.SearchTaskError)
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
