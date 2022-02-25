# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.2-alpha1] - TBD
### Added
- Add button to main GUI window which creates and starts a new task
- Add last 5 started tasks to top of tasklist widget in main GUI window

### Changed
- Default action on macOS when no CLI arguments are specified is to start the GUI
- Add Makefile target to install the fyne CLI tool before building the Darwin target

## [0.3.1] - 2021-12-12
- Released

## [0.3.1-rc1] - 2021-09-18
### Changed
- Removed incomplete sections from README.md to be completed at a later time

## [0.3.1-beta1] - 2021-09-11
### Changed
- Fixed defect in report where end date was not at the end of the day
- Update github.com/rs/zerolog from v1.23.0 to v1.25.0
- Update gorm.io/gorm from v1.21.11 to v1.21.15
- Update gorm.io/driver/sqlite from v1.1.4 to v1.1.5

## [0.3.1-alpha5] - 2021-08-22
### Added
- CSV export functionality added to GUI report

## [0.3.1-alpha4] - 2021-08-14
### Added
- CLI command to run a basic task report with aggregated task durations
- Use `go vet` as one of the linters
- GUI window to run basic task report

## [0.3.1-alpha3] - 2021-06-27
### Added
- Add CLI command to display a list of recently-started tasks
- Error state handling in system tray (again)
- TaskData object cache in Manage Window for tasks list widget to reduce database calls
- Enabled database transactions for creating or updating entities
- Add github.com/bluele/factory-go to support entity testing
- Add github.com/gofrs/uuid for UUID support

### Changed
- Transition to using interfaces for models with access to the underlying struct
- Increase minimum size of GUI windows to 640x480
- Increase width of Tasklist widget
- Update minimum Golang version from 1.15 to 1.16
- Update github.com/spf13/viper from v1.8.0 to v1.8.1

## [0.3.1-alpha2] - 2021-06-19
### Added
- Add checkbox option to close main window when stopping a task when invoked from a tray
- Add elapsed time counter to main window when a task is running
- Add github.com/gen2brain/beeep for system notifications from tray
- Tray shows notifications for any errors that occur when getting timesheet status
- Add options menu in system tray
- Add option to the tray menu to toggle a confirmation dialog when stopping a running task

### Changed
- Trim the active task text, so it doesn't expand the window size

### Removed
- Removed QUIT button from main window

## [0.3.1-alpha1] - 2021-06-19
### Added
- Add checkbox option to close main window when creating and starting a new task when invoked from the tray
- Send system notifications when starting or stopping a task from the GUI

### Changed
- Start time in main window now uses RFC-1123Z format
- The GUI app will exit when the main window closes
- Upgrade gorm.io/gorm from v1.21.10 to v1.21.11
- Upgrade github.com/rs/zerolog from v1.22.0 to v1.23.0

## [0.3.0] - 2021-06-18
### Added
- Added a GUI to create or update tasks
- Started a changelog

### Changed
- Separated system tray and GUI implementation due to cross-platform limitations
