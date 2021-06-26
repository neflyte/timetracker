package tray

import (
	"errors"
	"fmt"
	"github.com/neflyte/timetracker/internal/logger"
	"github.com/spf13/viper"
	"os"
	"path"
)

const (
	keyStopTaskConfirm = "stop-task-confirm"
)

func init() {
	// Set up config file parameters
	viper.SetConfigName("timetracker-tray")
	viper.SetConfigType("yaml")
	// Set config file paths:
	// 1. User's config directory
	userConfDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("*  error reading user config directory: %s\n", err.Error())
	} else {
		userTimetrackerConfDir := path.Join(userConfDir, "timetracker")
		// Make sure the directory exists before we add it
		mkdirErr := os.MkdirAll(userTimetrackerConfDir, 0755)
		if mkdirErr != nil {
			fmt.Printf("*  error creating configuration directory '%s': %s\n", userTimetrackerConfDir, mkdirErr.Error())
		} else {
			viper.AddConfigPath(userTimetrackerConfDir)
		}
	}
	// 2. Current directory
	viper.AddConfigPath(".")
	// Set default config values
	viper.SetDefault(keyStopTaskConfirm, true) // Confirm when stopping an active task
	// Ensure the config file exists
	err = viper.SafeWriteConfig()
	if err != nil {
		var viperAlreadyExistsError viper.ConfigFileAlreadyExistsError
		isConfigExistsErr := errors.Is(err, viperAlreadyExistsError)
		if !isConfigExistsErr {
			fmt.Printf("*  error creating config file: %s\n", err.Error())
		}
	}
}

func readConfig() error {
	log := logger.GetFuncLogger(trayLogger, "readConfig")
	err := viper.ReadInConfig()
	if err != nil {
		if !errors.Is(err, viper.ConfigFileNotFoundError{}) {
			log.Err(err).Msg("error reading configuration file")
			return err
		}
	}
	return nil
}

func writeConfig() error {
	log := logger.GetFuncLogger(trayLogger, "writeConfig")
	err := viper.WriteConfig()
	if err != nil {
		log.Err(err).Msg("error writing configuration file")
		return err
	}
	return nil
}
