# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist-linux dist-darwin dist-windows outdated ensure-fyne-cli
.PHONY: generate-icons-darwin generate-icons-windows generate-bundled-icons ensure-dist-directory build-cli build-gui build-tray

# Set platform-specific build variables
ifeq ($(OS),Windows_NT)
SHELL=C:\\Windows\\system32\\cmd.exe
.SHELLFLAGS=/C
APPVERSION=$(shell type VERSION)
SHORTAPPVERSION=$(shell FOR /F "delims=v tokens=1" %%I IN (VERSION) DO @ECHO %%I)
BUILD_FILENAME=timetracker.exe
GUI_BUILD_FILENAME=timetracker-gui.exe
TRAY_BUILD_FILENAME=timetracker-tray.exe
GO_LDFLAGS_EXTRA=
GUI_GO_LDFLAGS_EXTRA=-H windowsgui
TRAY_GO_LDFLAGS_EXTRA=-H windowsgui
else
APPVERSION=$(shell cat VERSION)
SHORTAPPVERSION=$(shell sed -E -e "s/v([0-9.]*).*/\1/" VERSION)
BUILD_FILENAME=timetracker
GUI_BUILD_FILENAME=timetracker-gui
TRAY_BUILD_FILENAME=timetracker-tray
GO_LDFLAGS_EXTRA=
GUI_GO_LDFLAGS_EXTRA=
TRAY_GO_LDFLAGS_EXTRA=
endif

# Set platform-independent build variables
GO_LDFLAGS=-ldflags "-s -X 'github.com/neflyte/timetracker/cmd/timetracker/cmd.AppVersion=$(APPVERSION)' $(GO_LDFLAGS_EXTRA)"
GUI_GO_LDFLAGS=-ldflags "-s -X 'github.com/neflyte/timetracker/cmd/timetracker-gui/cmd.AppVersion=$(APPVERSION)' $(GUI_GO_LDFLAGS_EXTRA)"
TRAY_GO_LDFLAGS=-ldflags "-s -X 'github.com/neflyte/timetracker/cmd/timetracker-tray/cmd.AppVersion=$(APPVERSION)' $(TRAY_GO_LDFLAGS_EXTRA)"
BINPREFIX=timetracker-$(APPVERSION)_
GUI_BINPREFIX=timetracker-gui-$(APPVERSION)_
TRAY_BINPREFIX=timetracker-tray-$(APPVERSION)_

build: ensure-dist-directory build-cli build-gui build-tray

ensure-dist-directory:
ifeq ($(OS),Windows_NT)
	CMD /C IF NOT EXIST dist MD dist
else
	if [ ! -d dist ]; then mkdir dist; fi
endif

build-cli:
	go build $(GO_LDFLAGS) -o dist/$(BUILD_FILENAME) ./cmd/timetracker

build-gui:
	go build $(GUI_GO_LDFLAGS) -o dist/$(GUI_BUILD_FILENAME) ./cmd/timetracker-gui

build-tray:
	go build $(TRAY_GO_LDFLAGS) -o dist/$(TRAY_BUILD_FILENAME) ./cmd/timetracker-tray

clean-coverage:
ifeq ($(OS),Windows_NT)
	CMD /C IF EXIST coverage RD /S /Q coverage
else
	if [ -d coverage ]; then rm -Rf coverage; fi
endif

clean: clean-coverage
ifeq ($(OS),Windows_NT)
	CMD /C IF EXIST dist RD /S /Q dist
else
	if [ -d dist ]; then rm -Rf dist; fi
endif

lint:
	golangci-lint run --timeout=5m

test: clean-coverage
ifeq ($(OS),Windows_NT)
	CMD /C IF NOT EXIST coverage MD coverage
else
	if [ ! -d coverage ]; then mkdir coverage; fi
endif
	go test -covermode=count -coverprofile=coverage/cover.out ./...
	go tool cover -html=coverage/cover.out -o coverage/coverage.html

dist-linux: lint build
	cp dist/$(BUILD_FILENAME) dist/$(BINPREFIX)linux-amd64
	cp dist/$(GUI_BUILD_FILENAME) dist/$(GUI_BINPREFIX)linux-amd64
	cp dist/$(TRAY_BUILD_FILENAME) dist/$(TRAY_BINPREFIX)linux-amd64
	cd dist && 7z a -mx9 $(BINPREFIX)linux-amd64.7z $(BINPREFIX)linux-amd64 $(GUI_BINPREFIX)linux-amd64 $(TRAY_BINPREFIX)linux-amd64

dist-darwin: ensure-fyne-cli lint build
	if [[ ! -d dist/darwin ]]; then mkdir -p dist/darwin; fi
	fyne package -name Timetracker -appID cc.ethereal.Timetracker -appVersion $(SHORTAPPVERSION) -appBuild 1 -os darwin -executable dist/$(GUI_BUILD_FILENAME) -icon assets/icons/icon-v2.png
	mv Timetracker.app dist/darwin
	cp dist/$(BUILD_FILENAME) dist/$(TRAY_BUILD_FILENAME) dist/darwin/Timetracker.app/Contents/MacOS/
	hdiutil create -srcfolder dist/darwin -volname "$(BINPREFIX)darwin-amd64" -imagekey zlib-level=9 dist/$(BINPREFIX)darwin-amd64.dmg

dist-windows: ensure-fyne-cli lint build
	CMD /C COPY /Y cmd\\timetracker-gui\\FyneApp.toml cmd\\timetracker-tray\\FyneApp.toml
	CMD /C COPY dist\\$(GUI_BUILD_FILENAME) cmd\\timetracker-gui\\timetracker-build.exe
	CMD /C "cd cmd\timetracker-gui && fyne package -name Timetracker -appVersion $(SHORTAPPVERSION) -appBuild 0 -os windows -executable timetracker-build.exe"
	CMD /C COPY cmd\\timetracker-gui\\Timetracker.exe dist\\$(GUI_BINPREFIX)windows-amd64.exe
	CMD /C DEL cmd\\timetracker-gui\\Timetracker.exe
	CMD /C COPY dist\\$(TRAY_BUILD_FILENAME) cmd\\timetracker-tray\\timetracker-build.exe
	CMD /C "cd cmd\timetracker-tray && fyne package -name Timetracker -appVersion $(SHORTAPPVERSION) -appBuild 0 -os windows -executable timetracker-build.exe"
	CMD /C COPY cmd\\timetracker-tray\\Timetracker.exe dist\\$(TRAY_BINPREFIX)windows-amd64.exe
	CMD /C DEL cmd\\timetracker-tray\\Timetracker.exe
	CMD /C DEL cmd\\timetracker-tray\\FyneApp.toml
	CMD /C COPY dist\\$(BUILD_FILENAME) dist\\$(BINPREFIX)windows-amd64.exe
	CMD /C "cd dist && 7z a -mx9 $(BINPREFIX)windows-amd64.7z $(BINPREFIX)windows-amd64.exe $(GUI_BINPREFIX)windows-amd64.exe $(TRAY_BINPREFIX)windows-amd64.exe"

outdated:
ifeq ($(OS),Windows_NT)
	CMD /C "pushd %HOMEDRIVE%%HOMEPATH% && go install github.com/psampaz/go-mod-outdated@v0.8.0 && popd"
else
	hash go-mod-outdated 2>/dev/null || { cd && go install github.com/psampaz/go-mod-outdated@v0.8.0; cd -; }
endif
	go list -json -u -m all | go-mod-outdated -direct -update

ensure-fyne-cli:
ifeq ($(OS),Windows_NT)
	CMD /C "pushd %HOMEDRIVE%%HOMEPATH% && go install fyne.io/fyne/v2/cmd/fyne@latest && popd"
else
	hash fyne 2>/dev/null || { cd && go install fyne.io/fyne/v2/cmd/fyne@latest; cd -; }
endif

generate-icons-darwin:
	bash scripts/generate_icns.sh assets/images/icon-v2.svg assets/icons
	bash scripts/generate_icns.sh assets/images/icon-v2-error.svg assets/icons
	bash scripts/generate_icns.sh assets/images/icon-v2-notrunning.svg assets/icons
	bash scripts/generate_icns.sh assets/images/icon-v2-running.svg assets/icons

generate-icons-windows:
	convert assets/images/icon-v2.svg -resize 256x256 assets/icons/icon-v2.ico
	convert assets/images/icon-v2-error.svg -resize 256x256 assets/icons/icon-v2-error.ico
	convert assets/images/icon-v2-notrunning.svg -resize 256x256 assets/icons/icon-v2-notrunning.ico
	convert assets/images/icon-v2-running.svg -resize 256x256 assets/icons/icon-v2-running.ico

generate-bundled-icons:
# macOS
	fyne bundle --name IconV2 -o internal/ui/icons/icon_v2_darwin.go --pkg icons assets/icons/icon-v2.png
	fyne bundle --name IconV2Error -o internal/ui/icons/icon_v2_error_darwin.go --pkg icons assets/icons/icon-v2-error.icns
	fyne bundle --name IconV2NotRunning -o internal/ui/icons/icon_v2_notrunning_darwin.go --pkg icons assets/icons/icon-v2-notrunning.icns
	fyne bundle --name IconV2Running -o internal/ui/icons/icon_v2_running_darwin.go --pkg icons assets/icons/icon-v2-running.icns
# windows
	fyne bundle --name IconV2 -o internal/ui/icons/icon_v2_windows.go --pkg icons assets/icons/icon-v2.ico
	fyne bundle --name IconV2Error -o internal/ui/icons/icon_v2_error_windows.go --pkg icons assets/icons/icon-v2-error.ico
	fyne bundle --name IconV2NotRunning -o internal/ui/icons/icon_v2_notrunning_windows.go --pkg icons assets/icons/icon-v2-notrunning.ico
	fyne bundle --name IconV2Running -o internal/ui/icons/icon_v2_running_windows.go --pkg icons assets/icons/icon-v2-running.ico
# linux
	fyne bundle --name IconV2 -o internal/ui/icons/icon_v2_linux.go --pkg icons assets/icons/icon-v2.png
	fyne bundle --name IconV2Error -o internal/ui/icons/icon_v2_error_linux.go --pkg icons assets/icons/icon-v2-error.png
	fyne bundle --name IconV2NotRunning -o internal/ui/icons/icon_v2_notrunning_linux.go --pkg icons assets/icons/icon-v2-notrunning.png
	fyne bundle --name IconV2Running -o internal/ui/icons/icon_v2_running_linux.go --pkg icons assets/icons/icon-v2-running.png
