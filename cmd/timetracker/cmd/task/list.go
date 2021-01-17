package task

import (
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	ListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE:  listTasks,
	}
	listDeletedTasks = false
)

func init() {
	ListCmd.Flags().BoolVarP(&listDeletedTasks, "deleted", "d", false, "Include deleted tasks")
}

func listTasks(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("listTasks")
	var result *gorm.DB
	if !listDeletedTasks {
		result = database.DB.Find(new(models.Task))
	} else {
		result = database.DB.Unscoped().Find(new(models.Task))
	}
	err := result.Error
	if err != nil {
		log.Printf("error listing tasks: %s\n", err)
		return err
	}
	rows, err := result.Rows()
	if err != nil {
		log.Printf("error getting result row: %s\n", err)
		return err
	}
	defer database.CloseRows(rows)
	for rows.Next() {
		task := new(models.Task)
		err = result.ScanRows(rows, task)
		if err != nil {
			log.Printf("error scanning row into &models.Task: %s\n", err)
			return err
		}
		log.Printf(task.String())
	}
	return nil
}
