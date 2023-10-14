package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/neflyte/timetracker/lib/models"
	"github.com/stretchr/testify/assert"
)

func TestUnit_NewCompactUI(t *testing.T) {
	compactui := NewCompactUI()
	testWindow := testApp.NewWindow("CompactUI Test")
	testWindow.SetContent(compactui)
	testWindow.Show()
	defer testWindow.Hide()
	assert.Equal(t, make([]string, 0), compactui.taskList)
	assert.Equal(t, make(models.TaskList, 0), compactui.taskModels)
	assert.NotNil(t, compactui.taskNameBinding)
	assert.NotNil(t, compactui.elapsedTimeBinding)
	assert.NotNil(t, compactui.commandChan)
	assert.Equal(t, "idle", compactui.taskNameLabel.Text)
	assert.Equal(t, "", compactui.elapsedTimeLabel.Text)
	assert.True(t, compactui.startStopButton.Disabled())
}

func TestUnit_SetTaskList(t *testing.T) {
	numTasks := 3
	compactui := NewCompactUI()
	testWindow := testApp.NewWindow("CompactUI Test")
	testWindow.SetContent(compactui)
	testWindow.Show()
	defer testWindow.Hide()
	taskList := make(models.TaskList, numTasks)
	for idx := range taskList {
		taskData, ok := TaskFactory.MustCreate().(*models.TaskData)
		if !ok {
			t.Error("cannot create new task")
		}
		taskList[idx] = models.NewTaskWithData(*taskData)
	}
	compactui.SetTaskList(taskList)
	assert.Equal(t, taskList, compactui.taskModels)
	assert.Equal(t, taskList.Names(), compactui.taskList)
	assert.Len(t, compactui.taskSelect.Options, len(taskList)+1)
}

func TestUnit_SetRunning(t *testing.T) {
	compactui := NewCompactUI()
	testWindow := testApp.NewWindow("CompactUI Test")
	testWindow.SetContent(compactui)
	testWindow.Show()
	defer testWindow.Hide()

	assert.False(t, compactui.IsRunning())
	assert.Equal(t, "START", compactui.startStopButton.Text)
	assert.Equal(t, theme.MediaPlayIcon(), compactui.startStopButton.Icon)
	assert.Equal(t, fyne.TextStyle{Italic: true}, compactui.taskNameLabel.TextStyle)

	compactui.SetRunning(true)

	assert.True(t, compactui.IsRunning())
	assert.Equal(t, "STOP", compactui.startStopButton.Text)
	assert.Equal(t, theme.MediaStopIcon(), compactui.startStopButton.Icon)
	assert.Equal(t, fyne.TextStyle{Bold: true}, compactui.taskNameLabel.TextStyle)

	compactui.SetRunning(false)

	assert.False(t, compactui.IsRunning())
	assert.Equal(t, "START", compactui.startStopButton.Text)
	assert.Equal(t, theme.MediaPlayIcon(), compactui.startStopButton.Icon)
	assert.Equal(t, fyne.TextStyle{Italic: true}, compactui.taskNameLabel.TextStyle)
}
