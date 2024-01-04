# timetracker
A personal time tracker for simple tasks
-

![Golang v1.18](https://img.shields.io/badge/Golang-v1.18-blue?style=for-the-badge&logo=go&color=00add8&link=https://golang.org)

[Features](#features) | [Installation](#installation) | [Usage](#usage) | [Changelog](https://github.com/neflyte/timetracker/CHANGELOG.md)

### What is it?

`timetracker` is a personal time tracker for simple tasks. It is loosely based on a time tracking program I used in the early 2000s.

### What does it do?

`timetracker` tracks the date and time that a task starts at and stops at in a database. It can report on how long was spent on each task in a given time period.

### Features

- Cross-platform; supporting Linux, macOS, and Windows
  - Tested on Ubuntu 22.04, macOS Monterey + Sonoma, Windows 10 and 11
- GUI app to start, stop, and manage tasks
- System tray app
  - Convenient access to start, stop, and create tasks
  - Task status (idle, running)
- CLI app for scripting or other command-line tasks

### Installation

#### Build Dependencies

To build `timetracker` from source, install the following tools followed by the dependencies that your Operating System requires below:

- [Golang v1.18](https://golang.org) or newer
- [GNU Make v3.82](https://www.gnu.org/software/make/) or newer
- [Git](https://git-scm.com/)

##### Linux

Run one of the commands below to install dependency packages:

###### Ubuntu/Debian

```shell
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev
```

###### Fedora

```shell
sudo dnf install golang gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel libXxf86vm-devel
```

##### macOS

Install the Xcode tools from the Terminal using the following command:

```shell
xcode-select --install
```

##### Other Operating Systems

Please consult the following sites for information on dependencies for other platforms:

- [fyne.io - Getting Started](https://developer.fyne.io/started/)
- [fyne-io/systray](https://github.com/fyne-io/systray)

#### Building

- Clone the repository to your machine using GitHub's download feature or by using the following `git` command:

```shell
git clone https://github.com/neflyte/timetracker
```

- Build the `timetracker` app:

```shell
make
```
  - The app will be placed in the `dist` subdirectory

#### Installing

##### Linux

- Copy the apps from the `dist` subdirectory to a directory on the system path, for example `/usr/local/bin` or `$HOME/bin`:

```shell
cp dist/timetracker* $HOME/bin
```

##### macOS

- Copy the `Timetracker.app` bundle from the `dist` subdirectory into the `Applications` folder.

##### Windows

- Copy the apps from the `dist` subdirectory to a new directory, for example `C:\Program Files\Timetracker`, and add it to the system `PATH` environment variable:

```powershell
New-Item 'C:\Program Files\Timetracker' -Type Directory -Force
Copy-Item dist\timetracker*.exe 'C:\Program Files\Timetracker'
```

### Usage

#### GUI app

##### Linux

To start the GUI app, run the following command:

```shell
timetracker-gui
```

##### macOS / Windows

To start the GUI app, double-click on the app icon.

#### System tray app

To start the system tray app as a background process, run one the following commands:

##### Linux

```shell
nohup timetracker-tray &
```

##### macOS

```shell
nohup /Applications/Timetracker.app/Contents/MacOS/timetracker-tray &
```

##### Windows

```powershell
timetracker-tray
```

- The system tray app can also be launched by double-clicking the `timetracker-tray.exe` app icon.
