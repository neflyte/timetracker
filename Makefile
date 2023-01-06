# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist dist-darwin dist-windows outdated ensure-fyne-cli generate-icons-darwin generate-icons-windows

ifeq ($(OS),Windows_NT)
APPVERSION=$(shell cmd /C type VERSION)
else
APPVERSION=$(shell cat VERSION)
endif
ifeq ($(OS),Darwin)
SHORTAPPVERSION=$(shell sed -E -e "s/v([0-9.]*).*/\1/" VERSION)
endif
OSES=darwin linux
GO_LDFLAGS=-ldflags "-X 'github.com/neflyte/timetracker/cmd/timetracker/cmd.AppVersion=$(APPVERSION)'"
BINPREFIX=timetracker-$(APPVERSION)_

build:
ifeq ($(OS),Windows_NT)
	CMD /C IF NOT EXIST dist MD dist
	go build $(GO_LDFLAGS) -o dist/timetracker.exe ./cmd/timetracker
else
	if [ ! -d dist ]; then mkdir dist; fi
	go build $(GO_LDFLAGS) -o dist/timetracker ./cmd/timetracker
endif

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

dist: lint
	@if [ ! -d dist ]; then mkdir dist; fi
	@for os in $(OSES); do \
		echo "Building for $$os" && \
		GOARCH=amd64 GOOS=$$os go build $(GO_LDFLAGS) -o dist/$(BINPREFIX)$$os-amd64 ./cmd/timetracker && \
		cd dist && \
		tar cfJ $(BINPREFIX)$$os-amd64.tar.xz $(BINPREFIX)$$os-amd64 && \
		sha512sum $(BINPREFIX)$$os-amd64.tar.xz > $(BINPREFIX)$$os-amd64.tar.xz.sha512 && \
		cd ..; \
	done

dist-darwin: ensure-fyne-cli lint
	GOOS=darwin GOARCH=amd64 go build $(GO_LDFLAGS) -o dist/$(BINPREFIX)darwin-amd64 ./cmd/timetracker
	fyne package -name Timetracker -os darwin -appID cc.ethereal.timetracker -appVersion "$(SHORTAPPVERSION)" -icon assets/images/Apps-Anydo-icon.png -executable dist/$(BINPREFIX)darwin-amd64
	mv Timetracker.app dist/

dist-windows: lint
	go build $(GO_LDFLAGS) -o dist/$(BINPREFIX)windows-amd64.exe ./cmd/timetracker
	7z a -txz dist/$(BINPREFIX)windows-amd64.exe.xz dist/$(BINPREFIX)windows-amd64.exe

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

