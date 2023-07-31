package timesheet

import (
	"fmt"
	"strconv"

	"github.com/alexeyco/simpletable"
	"github.com/fatih/color"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/neflyte/timetracker/internal/models"
	"github.com/neflyte/timetracker/internal/ui/cli"
	"github.com/spf13/cobra"
)

const defaultTaskLimit = 5

var (
	// LastStartedCmd represents the command that prints the last x started tasks
	LastStartedCmd = &cobra.Command{
		Use:   "last-started",
		Short: "Shows the last x started tasks",
		RunE:  doLastStarted,
	}
	taskLimit    uint
	outputFormat string
)

func init() {
	LastStartedCmd.Flags().UintVar(&taskLimit, "limit", defaultTaskLimit, "the number of tasks to return; must be greater than zero")
	LastStartedCmd.Flags().StringVar(&outputFormat, "outputFormat", outputFormatText, "output format (text, csv, json, xml; default text)")
}

func doLastStarted(_ *cobra.Command, _ []string) (err error) {
	lastStartedTasks, err := models.NewTimesheet().LastStartedTasks(taskLimit)
	if err != nil {
		return
	}
	printLastStarted(lastStartedTasks, outputFormat)
	return nil
}

func printLastStarted(tasks []models.TaskData, format string) {
	log := logger.GetLogger("printLastStarted")
	// Output using requested format
	switch format {
	case outputFormatText:
		printLastStartedTable(tasks)
	case outputFormatCSV:
		cli.PrintCSV(log, tasks)
	case outputFormatJSON:
		jsonData := struct {
			Tasks []models.TaskData `json:"Tasks"`
		}{
			Tasks: tasks,
		}
		cli.PrintJSON(log, jsonData)
	case outputFormatXML:
		xmlData := struct {
			Tasks []models.TaskData `xml:"Tasks"`
		}{
			Tasks: tasks,
		}
		cli.PrintXML(log, xmlData)
	}
}

func printLastStartedTable(tasks []models.TaskData) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "Task ID"},
			{Text: "Synopsis"},
			{Text: "Description"},
		},
	}
	for _, task := range tasks {
		rec := []*simpletable.Cell{
			{Text: strconv.Itoa(int(task.ID))},
			{Text: task.Synopsis},
			{Text: task.Description},
		}
		table.Body.Cells = append(table.Body.Cells, rec)
	}
	if len(table.Body.Cells) == 0 {
		fmt.Println(color.WhiteString("There are no tasks"))
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
}
