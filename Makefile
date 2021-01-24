# timetracker Makefile

.PHONY: build clean lint dist dist7z

APPVERSION=$(shell cat VERSION)
OSES=darwin linux
GO_LDFLAGS=-ldflags "-X cmd.root.AppVersion=$(APPVERSION)"
BINPREFIX=timetracker-$(APPVERSION)-

build:
	go build -i -installsuffix $(shell go env GOOS) -pkgdir $(shell go env GOPATH)/pkg $(GO_LDFLAGS) -o timetracker ./cmd/timetracker

clean:
	{ [ -f timetracker ] && rm -f timetracker; } || true
	{ [ -d dist ] && rm -Rf dist; } || true

lint:
	{ type -p golangci-lint >/dev/null 2>&1 && golangci-lint run; } || true

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
