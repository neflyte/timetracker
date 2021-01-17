# timetracker Makefile

.PHONY: build clean lint

build:
	go build ./cmd/timetracker

clean:
	{ [ -f "./timetracker" ] && rm -f "./timetracker"; } || true

lint:
	{ type -p golangci-lint >/dev/null 2>&1 && golangci-lint run; } || true

