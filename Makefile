# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist outdated ensure-fyne-cli

APPVERSION=$(shell cat VERSION)
SHORTAPPVERSION=$(shell sed -E -e "s/v([0-9.]*).*/\1/" VERSION)
OSES=darwin linux
GO_LDFLAGS=-ldflags "-X 'github.com/neflyte/timetracker/cmd/timetracker/cmd.AppVersion=$(APPVERSION)'"
BINPREFIX=timetracker-$(APPVERSION)_

build:
	if [ ! -d dist ]; then mkdir dist; fi
	go build $(GO_LDFLAGS) -o dist/timetracker ./cmd/timetracker

clean-coverage:
	if [ -d coverage ]; then rm -Rf coverage; fi

clean: clean-coverage
	if [ -d dist ]; then rm -Rf dist; fi

lint:
	golangci-lint run --timeout=5m

test: clean-coverage
	if [ ! -d coverage ]; then mkdir coverage; fi
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

outdated:
	hash go-mod-outdated 2>/dev/null || { cd && go install github.com/psampaz/go-mod-outdated@v0.8.0; cd -; }
	go list -json -u -m all | go-mod-outdated -direct -update

ensure-fyne-cli:
	@echo "Checking for fyne CLI tool"
	hash fyne 2>/dev/null || { cd && go install fyne.io/fyne/v2/cmd/fyne@latest; cd -; }
