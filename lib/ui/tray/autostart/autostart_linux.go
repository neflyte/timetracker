package autostart

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/ui/icons"
)

const (
	desktopFileMode = 0644
)

var (
	desktopFile = ""
)

// Enable enables tray autostart
func Enable() error {
	err := writeIcon()
	if err != nil {
		return err
	}
	err = writeDesktopFile()
	if err != nil {
		return err
	}
	return nil
}

// Disable disables tray autostart
func Disable() error {
	err := removeDesktopFile()
	if err != nil {
		return err
	}
	err = removeIcon()
	if err != nil {
		return err
	}
	return nil
}

// IsEnabled determines if tray autostart is enabled
func IsEnabled() bool {
	desktopFilePath := path.Join(xdg.DataHome, "applications", fmt.Sprintf("%s.desktop", constants.AppId))
	_, err := os.Stat(desktopFilePath)
	return err == nil
}

func generateXdgDesktopSpec() (string, error) {
	trayPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	lines := []string{
		"[Desktop Entry]",
		"Version = 1.5",
		"Type = Application",
		"Name = Timetracker Tray",
		"Comment = Timetracker tray icon",
		fmt.Sprintf("Icon = %s", iconPath),
		fmt.Sprintf("Exec = %s", trayPath),
		"Terminal = false",
		"Categories = Office;Utility;",
	}
	return strings.Join(lines, "\n"), nil
}

func writeIcon() error {
	if iconPath == "" {
		iconPath = path.Join(xdg.DataHome, "icons", fmt.Sprintf("%s.png", constants.AppId))
		err := os.WriteFile(iconPath, icons.IconV2.StaticContent, iconFileMode)
		if err != nil {
			iconPath = ""
			return err
		}
	}
	return nil
}

func removeIcon() error {
	if iconPath != "" {
		_, err := os.Stat(iconPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		err = os.Remove(iconPath)
		if err != nil {
			return err
		}
		iconPath = ""
	}
	return nil
}

func writeDesktopFile() error {
	if desktopFile == "" {
		desktopFileContent, err := generateXdgDesktopSpec()
		if err != nil {
			return err
		}
		desktopFile = path.Join(xdg.DataHome, "applications", fmt.Sprintf("%s.desktop", constants.AppId))
		err = os.WriteFile(desktopFile, []byte(desktopFileContent), desktopFileMode)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeDesktopFile() error {
	if desktopFile != "" {
		_, err := os.Stat(desktopFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		err = os.Remove(desktopFile)
		if err != nil {
			return err
		}
		desktopFile = ""
	}
	return nil
}
