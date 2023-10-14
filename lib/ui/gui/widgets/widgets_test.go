package widgets

import (
	"errors"
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/bluele/factory-go/factory"
	"github.com/gofrs/uuid"
	"github.com/neflyte/timetracker/lib/logger"
	"github.com/neflyte/timetracker/lib/models"
)

var (
	testApp     fyne.App
	TaskFactory = factory.NewFactory(new(models.TaskData)).
			Attr("Synopsis", func(_ factory.Args) (interface{}, error) {
			// Use a v4 UUID as the task synopsis
			synuuid, err := uuid.NewV4()
			if err != nil {
				return nil, err
			}
			return synuuid.String(), nil
		}).
		Attr("Description", func(args factory.Args) (interface{}, error) {
			taskData, ok := args.Instance().(*models.TaskData)
			if !ok {
				return nil, errors.New("args for Description was not *TaskData; this is unexpected")
			}
			return fmt.Sprintf("description for task %s", taskData.Synopsis), nil
		})
)

func TestMain(m *testing.M) {
	logger.InitLogger("debug", true)
	defer logger.CleanupLogger()
	testApp = test.NewApp()
	defer testApp.Quit()
	go testApp.Run()
	ret := m.Run()
	if ret > 0 {
		panic(
			fmt.Sprintf("Test failed with exit code %d", ret),
		)
	}
}
