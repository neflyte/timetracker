# timetracker
A personal time tracker for simple tasks
-

![Golang v1.16](https://img.shields.io/badge/Golang-v1.16-blue?style=for-the-badge&logo=go&color=00add8&link=https://golang.org)

[Features](#features) | [Installation](#installation) | [Usage](#usage)

### What is it?

`timetracker` is a personal time tracker for simple tasks. It is loosely based on a time tracking program I used in the early 2000s.

### What does it do?

`timetracker` tracks the date and time that a task starts at and stops at in a simple database. It can report on how long was spent on each task in a given time period.

### Features

- (Theoretically) cross-platform; supporting Linux, macOS, and Windows
  - Tested on Ubuntu 20.04 + macOS Catalina
- GUI app to start, stop, and manage tasks
- System tray app
  - Convenient access to start, stop, and create tasks
  - Task status (idle, running)

### Installation

#### System Requirements

To build `timetracker` from source, install the following tools followed by the dependencies that your Operating System requires below:

- [Golang v1.16+](https://golang.org)
- [GNU Make](https://www.gnu.org/software/make/)
- [Git](https://git-scm.com/) (for downloading on the command line)

##### Linux

Run one of the commands below to install dependency packages:

###### Ubuntu/Debian

`sudo apt-get install gcc libgl1-mesa-dev xorg-dev libgtk-3-dev libappindicator3-dev`

###### Linux Mint

`sudo apt-get install gcc libgl1-mesa-dev xorg-dev libgtk-3-dev libappindicator3-dev libxapp-dev`

###### Fedora

`sudo dnf install gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel libXxf86vm-devel`

###### Arch Linux

`sudo pacman -S xorg-server-devel`

##### macOS

Install the Xcode tools from the Terminal using the following command:

`xcode-select --install`

##### Other Operating Systems

Please consult the following sites for information on dependencies for other platforms:

- [fyne.io - Getting Started](https://developer.fyne.io/started/)
- [getlantern/systray](https://github.com/getlantern/systray)

#### Building

- Clone the repository to your machine using GitHub's download feature or by using the following `git` command:

  `git clone https://github.com/neflyte/timetracker`

- Build the `timetracker` app:

  `make`

  - The app will be placed in the `dist` subdirectory

#### Installing
- Copy the app to a directory on the system path, for example `/usr/local/bin` or `$HOME/bin`:

  `cp dist/timetracker $HOME/bin`

### Usage

#### System tray app

##### Starting

To start the system tray app as a background process, run the following command:

`nohup timetracker tray &`

On Linux and macOS this command will start a new `timetracker` process and detach it from the terminal.

#### GUI app

##### Starting

To start the GUI app, run the following command:

`timetracker ui`
