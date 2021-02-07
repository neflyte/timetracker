# timetracker Makefile

.PHONY: build clean clean-coverage lint test dist

APPVERSION=$(shell cat VERSION)
OSES=darwin linux
GO_LDFLAGS=-ldflags "-X cmd.root.AppVersion=$(APPVERSION)"
BINPREFIX=timetracker-$(APPVERSION)_

build:
	go build -i -installsuffix $(shell go env GOOS) -pkgdir $(shell go env GOPATH)/pkg $(GO_LDFLAGS) -o timetracker ./cmd/timetracker

clean-coverage:
	{ [ -f cover.out ] && rm -f cover.out; } || true
	{ [ -f coverage.html ] && rm -f coverage.html; } || true

clean: clean-coverage
	{ [ -f timetracker ] && rm -f timetracker; } || true
	{ [ -d dist ] && rm -Rf dist; } || true

lint:
	{ type -p golangci-lint >/dev/null 2>&1 && golangci-lint run; } || true

test: clean-coverage
	go test -covermode=count -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o coverage.html

dist: lint
	[ -d dist ] || mkdir dist
	@for os in $(OSES); do \
		echo "Building for $$os" && \
  		GOARCH=amd64 GOOS=$$os go build -i -installsuffix $$os -pkgdir $(shell go env GOPATH)/pkg -o dist/$(BINPREFIX)$$os-amd64 ./cmd/timetracker && \
  		cd dist && \
  		tar cfJ $(BINPREFIX)$$os-amd64.tar.xz $(BINPREFIX)$$os-amd64 && \
        sha512sum $(BINPREFIX)$$os-amd64.tar.xz > $(BINPREFIX)$$os-amd64.tar.xz.sha512 && \
        cd ..; \
    done
