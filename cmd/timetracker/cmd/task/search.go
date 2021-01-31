package task

import (
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"strconv"
)

var (
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
	rows, err := database.DB.
		Model(new(models.Task)).
		Where("synopsis LIKE ? OR description LIKE ?", args[0], args[0]).
		Rows()
	if err != nil {
		fmt.Println(chalk.Red, "Error searching for tasks:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("error searching for tasks")
		return err
	}
	defer database.CloseRows(rows)
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "#"},
			{Text: "Synopsis"},
			{Text: "Description"},
			{Text: "Created At"},
			{Text: "Updated At"},
		},
	}
	var task models.Task
	for rows.Next() {
		err = database.DB.ScanRows(rows, &task)
		if err != nil {
			fmt.Println(chalk.Red, "Error scanning result row:", chalk.White, chalk.Dim.TextStyle(err.Error()))
			log.Err(err).Msg("error scanning row into &models.Task")
			return err
		}
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(task.ID))},
			{Text: task.Synopsis},
			{Text: task.Description},
			{Text: task.CreatedAt.Format(`2006-01-02 15:04:05 PM`)},
			{Text: task.UpdatedAt.Format(`2006-01-02 15:04:05 PM`)},
		}
		table.Body.Cells = append(table.Body.Cells, rec)
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
	return nil
}
