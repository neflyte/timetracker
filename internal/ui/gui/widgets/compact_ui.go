package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/rs/zerolog"
)

var _ fyne.Widget = (*CompactUi)(nil)

type CompactUi struct {
	taskNameBinding    binding.String
	elapsedTimeBinding binding.String
	container          *fyne.Container
	taskSelect         *widget.Select
	startStopButton    *widget.Button
	taskNameLabel      *widget.Label
	elapsedTimeLabel   *widget.Label
	ellipsisButton     *widget.Button
	log                zerolog.Logger
	taskList           []string
	widget.BaseWidget
}

func NewCompactUi() *CompactUi {
	compactui := &CompactUi{
		log:                logger.GetStructLogger("CompactUi"),
		taskList:           make([]string, 0),
		taskNameBinding:    binding.NewString(),
		elapsedTimeBinding: binding.NewString(),
	}
	compactui.ExtendBaseWidget(compactui)
	compactui.initUI()
	return compactui
}

func (c *CompactUi) initUI() {
	c.taskSelect = widget.NewSelect(c.taskList, c.taskWasSelected)
	c.taskSelect.PlaceHolder = "Select a task" // i18n
	c.startStopButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), c.startStopButtonWasTapped)
	c.taskNameLabel = widget.NewLabelWithData(c.taskNameBinding)
	c.elapsedTimeLabel = widget.NewLabelWithData(c.elapsedTimeBinding)
	c.ellipsisButton = widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), c.ellipsisButtonWasTapped)
	c.container = container.NewHBox(
		c.taskSelect,
		c.startStopButton,
		c.taskNameLabel,
		c.elapsedTimeLabel,
		c.ellipsisButton,
	)
}

func (c *CompactUi) taskWasSelected(selection string) {

}

func (c *CompactUi) startStopButtonWasTapped() {

}

func (c *CompactUi) ellipsisButtonWasTapped() {

}

// CreateRenderer returns a new WidgetRenderer for this widget
func (c *CompactUi) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}
