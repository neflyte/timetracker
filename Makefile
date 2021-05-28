# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist

APPVERSION=$(shell cat VERSION)
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
	golangci-lint run

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
