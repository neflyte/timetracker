# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.4] - 2023-01-04
### Added
- Improved notification support for Windows
- Unique index on the `timesheet.StopTime` database column to ensure only one running task at a time
- First unit tests for UI widgets

### Changed
- Updated minimum Golang version to v1.18
- Task selector defaults to reverse (latest-first) sort
- Split single-file binary into 3 binaries - CLI, GUI, and Tray - for more efficient usage on macOS and Windows platforms
- Default main window size changed from 640x480 to 800x600
- Updated fyne.io/fyne/v2 from v2.3.0 to v2.4.3
- Updated fyne.io/systray from v1.10.1-0.20230722100817-88df1e0ffa9a to v1.10.1-0.20231115130155-104f5ef7839e
- Updated github.com/fatih/color from v1.13.0 to v1.15.0
- Updated github.com/gen2brain/beeep from v0.0.0-20220909211152-5a9ec94374f6 to v0.0.0-20230907135156-1a38885a97fc
- Updated github.com/gofrs/uuid from v4.3.1 to v4.4.0
- Updated github.com/rs/zerolog from v1.28.0 to v1.31.0
- Updated github.com/spf13/cobra from v1.6.1 to v1.7.0
- Updated github.com/spf13/viper from v1.14.0 to v1.17.0
- Updated github.com/stretchr/testify from v1.8.1 to v1.8.2
- Updated golang.org/x/sys from v0.5.0 to v0.15.0
- Updated golang.org/x/net from v0.4.0 to v0.17.0
- Updated golang.org/x/exp from v0.0.0-20231006140011-7918f672742d to v0.0.0-20231226003508-02704c960a9b
- Updated gorm.io/driver/sqlite from v1.4.4 to v1.5.4
- Updated gorm.io/gorm from v1.24.3 to v1.25.5

### Removed

## [0.3.3] - 2023-02-12
### Added
- New task widget
- New task selector widget
- New task editor widget
- New task management window

### Changed
- Simplified main window UI
- All app icons are updated; new icons created from (mostly) scratch
- Windows builds support GUI and CLI operation with a single binary
- Package macOS build in a disk image (DMG)
- Updated fyne.io/fyne/v2 from v2.2.4 to v2.3.0
- Updated gorm.io/gorm from v1.24.2 to v1.24.3
- Updated gorm.io/drive/sqlite from v1.4.3 to v1.4.4

### Removed
- Task selector pick list

## [0.3.2] - 2022-12-03
### Added
- Add preliminary support for building on Win32

### Changed
- Updated fyne.io/fyne/v2 from v2.2.1 to v2.2.4
- Updated github.com/gen2brain/beeep from v0.0.0-20220518085355-d7852edf42fc to v0.0.0-20220909211152-5a9ec94374f6
- Updated github.com/gofrs/uuid from v4.2.0+incompatible to v4.3.1+incompatible
- Updated github.com/rs/zerolog from v1.27.0 to v1.28.0
- Updated github.com/spf13/cobra from v1.5.0 to v1.6.1
- Updated github.com/spf13/viper from v1.12.0 to v1.14.0
- Updated github.com/stretchr/testify from v1.7.5 to v1.8.1
- Updated gorm.io/gorm from v1.23.6 to v1.24.2
- Updated gorm.io/drivers/sqlite from v1.3.4 to v1.4.3
- Replaced github.com/getlantern/systray with fyne.io/systray

## [0.3.2-alpha1] - 2022-06-03
### Added
- Add button to main GUI window which creates and starts a new task
- Add last 5 started tasks to top of tasklist widget in main GUI window

### Changed
- Default action on macOS when no CLI arguments are specified is to start the GUI
- Add Makefile target to install the fyne CLI tool before building the Darwin target
- If a PID file exists, validate that the PID is valid at startup and force-remove the PID file if it is not
- Updated fyne.io/fyne/v2 from v2.1.4 to v2.2.1
- Updated github.com/getlantern/systray from v1.1.0 to v1.2.1
- Updated github.com/rs/zerolog from v1.26.1 to v1.27.0
- Updated github.com/spf13/cobra from v1.4.0 to v1.5.0
- Updated github.com/spf13/viper from v1.11.0 to v1.12.0
- Updated github.com/stretchr/testify from v1.7.1 to v1.7.5
- Updated gorm.io/gorm from v1.23.5 to v1.23.6
- Updated gotm.io/driver/sqlite from v1.3.2 to v1.3.4

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
