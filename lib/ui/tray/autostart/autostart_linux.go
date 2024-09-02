package autostart

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/neflyte/timetracker/lib/constants"
	"github.com/neflyte/timetracker/lib/logger"
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
	log := logger.GetFuncLogger(packageLogger, "IsEnabled")
	desktopFilePath := path.Join(xdg.ConfigHome, "autostart", fmt.Sprintf("%s.desktop", constants.AppID))
	log.Debug().
		Str("desktopFilePath", desktopFilePath).
		Msg("stat file")
	_, err := os.Stat(desktopFilePath)
	if err != nil {
		log.Err(err).
			Msg("error stating desktop file; assume autostart is not enabled")
		return false
	}
	log.Debug().
		Msg("autostart is enabled")
	if desktopFile == "" {
		desktopFile = desktopFilePath
	}
	return true
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
	log := logger.GetFuncLogger(packageLogger, "IsEnabled")
	if iconPath == "" {
		iconPath = path.Join(xdg.DataHome, "icons", fmt.Sprintf("%s.png", constants.AppID))
		log.Debug().
			Str("iconPath", iconPath).
			Msg("write icon file")
		err := os.WriteFile(iconPath, icons.IconV2.StaticContent, iconFileMode)
		if err != nil {
			log.Err(err).
				Str("iconPath", iconPath).
				Msg("unable to write icon file")
			iconPath = ""
			return err
		}
		log.Debug().
			Str("iconPath", iconPath).
			Msg("wrote icon file successfully")
	}
	return nil
}

func removeIcon() error {
	log := logger.GetFuncLogger(packageLogger, "removeIcon").
		With().
		Str("iconPath", iconPath).
		Logger()
	if iconPath != "" {
		log.Debug().
			Msg("stat icon file")
		_, err := os.Stat(iconPath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug().
					Msg("icon file does not exist")
				return nil
			}
			log.Err(err).
				Msg("unexpected error when stating icon file")
			return err
		}
		log.Debug().
			Msg("remove icon file")
		err = os.Remove(iconPath)
		if err != nil {
			log.Err(err).
				Msg("error removing icon file")
			return err
		}
		log.Debug().
			Msg("removed icon file successfully")
		iconPath = ""
	}
	return nil
}

func writeDesktopFile() error {
	log := logger.GetFuncLogger(packageLogger, "writeDesktopFile")
	if desktopFile == "" {
		log.Debug().
			Msg("generate desktop file content")
		desktopFileContent, err := generateXdgDesktopSpec()
		if err != nil {
			log.Err(err).
				Msg("unable to generate desktop file content")
			return err
		}
		log.Debug().
			Msg(desktopFileContent)
		desktopFile = path.Join(xdg.ConfigHome, "autostart", fmt.Sprintf("%s.desktop", constants.AppID))
		log.Debug().
			Str("desktopFile", desktopFile).
			Msg("write desktop file")
		err = os.WriteFile(desktopFile, []byte(desktopFileContent), desktopFileMode)
		if err != nil {
			log.Err(err).
				Str("desktopFile", desktopFile).
				Msg("unable to write desktop file")
			return err
		}
		log.Debug().
			Str("desktopFile", desktopFile).
			Msg("wrote desktop file successfully")
	}
	return nil
}

func removeDesktopFile() error {
	log := logger.GetFuncLogger(packageLogger, "removeDesktopFile").
		With().
		Str("desktopFile", desktopFile).
		Logger()
	if desktopFile != "" {
		log.Debug().
			Msg("stating desktop file")
		_, err := os.Stat(desktopFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug().
					Msg("desktop file does not exist")
				return nil
			}
			log.Err(err).
				Msg("unexpected error when stating desktop file")
			return err
		}
		log.Debug().
			Msg("remove desktop file")
		err = os.Remove(desktopFile)
		if err != nil {
			log.Err(err).
				Msg("error removing desktop file")
			return err
		}
		log.Debug().
			Msg("removed desktop file successfully")
		_, err = os.Stat(desktopFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Debug().
					Msg("validated that desktop file was removed")
			} else {
				log.Err(err).
					Msg("unexpected error when stating desktop file after removing it")
			}
		} else {
			log.Error().
				Str("desktopFile", desktopFile).
				Msg("desktopFile still exists after removing it")
		}
		desktopFile = ""
	}
	return nil
}
