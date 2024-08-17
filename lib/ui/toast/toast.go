package toast

import "github.com/neflyte/timetracker/lib/logger"

const (
	DefaultWindowsPowershellPath = "System32\\WindowsPowerShell\\v1.0\\powershell.exe"

	tempFileMode = 0600
)

var (
	packageLogger = logger.GetPackageLogger("toast")
)

type Toast interface {
	Notify(title string, description string) error
	Cleanup()
}
