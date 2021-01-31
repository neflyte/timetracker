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
	db := database.DB
	if listDeletedTasks {
		db = database.DB.Unscoped()
	}
	tasks := make([]models.Task, 0)
	err := db.Find(&tasks).Error
	if err != nil {
		fmt.Println(chalk.Red, "Error getting result row:", chalk.White, chalk.Dim.TextStyle(err.Error()))
		log.Err(err).Msg("error getting result row")
		return err
	}
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
	for _, task := range tasks {
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
