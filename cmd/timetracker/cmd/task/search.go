package task

import (
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/neflyte/timetracker/internal/constants"
	"github.com/neflyte/timetracker/internal/errors"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
	"strconv"
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
