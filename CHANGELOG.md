# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
