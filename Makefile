# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist-linux dist-darwin dist-windows outdated ensure-fyne-cli
.PHONY: generate-icons-darwin generate-icons-windows generate-bundled-icons ensure-dist-directory build-cli build-gui

ifeq ($(OS),Windows_NT)
SHELL=CMD.EXE
.SHELLFLAGS=/C
APPVERSION=$(shell type VERSION)
SHORTAPPVERSION=$(shell FOR /F "delims=v tokens=1" %%I IN (VERSION) DO @ECHO %%I)
BUILD_FILENAME=timetracker.exe
GUI_BUILD_FILENAME=timetracker-gui.exe
GO_LDFLAGS_EXTRA=
GUI_GO_LDFLAGS_EXTRA=-H windowsgui
else
APPVERSION=$(shell cat VERSION)
SHORTAPPVERSION=$(shell sed -E -e "s/v([0-9.]*).*/\1/" VERSION)
BUILD_FILENAME=timetracker
GUI_BUILD_FILENAME=timetracker-gui
GO_LDFLAGS_EXTRA=
GUI_GO_LDFLAGS_EXTRA=
endif
GO_LDFLAGS=-ldflags "-s -X 'github.com/neflyte/timetracker/cmd/timetracker/cmd.AppVersion=$(APPVERSION)' $(GO_LDFLAGS_EXTRA)"
GUI_GO_LDFLAGS=-ldflags "-s -X 'github.com/neflyte/timetracker/cmd/timetracker-gui/main.AppVersion=$(APPVERSION)' $(GUI_GO_LDFLAGS_EXTRA)"
BINPREFIX=timetracker-$(APPVERSION)_
GUI_BINPREFIX=timetracker-gui-$(APPVERSION)_

build: ensure-dist-directory build-cli build-gui

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
	cd dist && tar cvJf $(BINPREFIX)linux-amd64.tar.xz $(BINPREFIX)linux-amd64 $(GUI_BINPREFIX)linux-amd64

dist-darwin: ensure-fyne-cli lint build
	if [[ ! -d dist/darwin ]]; then mkdir -p dist/darwin; fi
	fyne package -name Timetracker -appVersion $(SHORTAPPVERSION) -appBuild 0 -os darwin -executable dist/$(BUILD_FILENAME) -icon assets/icons/icon-v2.png
	mv Timetracker.app dist/darwin
	hdiutil create -srcfolder dist/darwin -volname "$(BINPREFIX)darwin-amd64" -imagekey zlib-level=9 dist/$(BINPREFIX)darwin-amd64.dmg

dist-windows: lint build
	CMD /C COPY dist\\$(BUILD_FILENAME) cmd\\timetracker\\timetracker-build.exe
	CMD /C "cd cmd\timetracker && fyne package -name Timetracker -appVersion $(SHORTAPPVERSION) -appBuild 0 -os windows -executable timetracker-build.exe"
	CMD /C COPY cmd\\timetracker\\Timetracker.exe dist\\$(BINPREFIX)windows-amd64.exe
	CMD /C DEL cmd\\timetracker\\Timetracker.exe
	7z a -txz dist\\$(BINPREFIX)windows-amd64.exe.xz dist\\$(BINPREFIX)windows-amd64.exe

outdated:
ifneq ($(OS),Windows_NT)
	hash go-mod-outdated 2>/dev/null || { cd && go install github.com/psampaz/go-mod-outdated@v0.8.0; cd -; }
endif
	go list -json -u -m all | go-mod-outdated -direct -update

ensure-fyne-cli:
	hash fyne 2>/dev/null || { cd && go install fyne.io/fyne/v2/cmd/fyne@latest; cd -; }

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
