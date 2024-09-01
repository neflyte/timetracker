package autostart

import "github.com/neflyte/timetracker/lib/logger"

const (
	iconFileMode = 0644
	tempFileMode = 0644
)

var (
	packageLogger = logger.GetPackageLogger("ui/autostart")
	iconPath      = ""
)
