package timesheet

import (
	"github.com/neflyte/timetracker/internal/database"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/spf13/cobra"
)

var (
	DumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dumps all timesheets",
		RunE:  dumpTimesheets,
	}
)

func dumpTimesheets(_ *cobra.Command, _ []string) error {
	log := logger.GetLogger("dumpTimesheets")
	timesheet := new(models.Timesheet)
	result := database.DB.Find(&timesheet)
	if result.Error != nil {
		log.Printf("error reading timesheets: %s\n", result.Error)
		return result.Error
	}
	log.Printf("records=%d", result.RowsAffected)
	rows, err := result.Rows()
	if err != nil {
		log.Printf("error getting result rows: %s\n", err)
		return err
	}
	defer database.CloseRows(rows)
	for rows.Next() {
		sheet := new(models.Timesheet)
		err = result.ScanRows(rows, sheet)
		if err != nil {
			log.Printf("error scanning result row: %s\n", err)
			return err
		}
		log.Println(sheet.String())
	}
	return nil
}
