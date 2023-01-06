# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist-darwin dist-windows outdated ensure-fyne-cli generate-icons-darwin generate-icons-windows generate-bundled-icons

ifeq ($(OS),Windows_NT)
SHELL=CMD.EXE
.SHELLFLAGS=/C
APPVERSION=$(shell type VERSION)
SHORTAPPVERSION=$(shell FOR /F "delims=v tokens=1" %%I IN (VERSION) DO @ECHO %%I)
BUILD_FILENAME=timetracker.exe
else
APPVERSION=$(shell cat VERSION)
SHORTAPPVERSION=$(shell sed -E -e "s/v([0-9.]*).*/\1/" VERSION)
BUILD_FILENAME=timetracker
endif
OSES=darwin linux
GO_LDFLAGS=-ldflags "-X 'github.com/neflyte/timetracker/cmd/timetracker/cmd.AppVersion=$(APPVERSION)'"
BINPREFIX=timetracker-$(APPVERSION)_

build:
ifeq ($(OS),Windows_NT)
	CMD /C IF NOT EXIST dist MD dist
else
	if [ ! -d dist ]; then mkdir dist; fi
endif
	go build $(GO_LDFLAGS) -o dist/$(BUILD_FILENAME) ./cmd/timetracker

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

dist-darwin: ensure-fyne-cli lint build
	fyne package -name Timetracker -appVersion $(SHORTAPPVERSION) -appBuild 0 -os darwin -executable dist/$(BUILD_FILENAME)
	mv Timetracker.app dist/

dist-windows: build
	CMD /C COPY dist\\$(BUILD_FILENAME) cmd\\timetracker
	CMD /C "cd cmd\timetracker && fyne package -name Timetracker-pkg -appVersion $(SHORTAPPVERSION) -appBuild 0 -os windows -executable $(BUILD_FILENAME)"
	CMD /C COPY cmd\\timetracker\\Timetracker-pkg.exe dist\\$(BINPREFIX)windows-amd64.exe
	CMD /C DEL cmd\\timetracker\\Timetracker-pkg.exe
	7z a -txz dist\\$(BINPREFIX)windows-amd64.exe.xz dist\\$(BINPREFIX)windows-amd64.exe

outdated:
ifneq ($(OS),Windows_NT)
	hash go-mod-outdated 2>/dev/null || { cd && go install github.com/psampaz/go-mod-outdated@v0.8.0; cd -; }
endif
	go list -json -u -m all | go-mod-outdated -direct -update

ensure-fyne-cli:
	@echo "Checking for fyne CLI tool"
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
