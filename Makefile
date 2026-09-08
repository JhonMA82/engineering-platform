.PHONY: build test vet cover release

# Release line stamped into internal/version via ldflags. Override for a
# real release: make release VERSION=1.0.0 COMMIT=$(git rev-parse HEAD).
VERSION ?= dev
COMMIT ?= unknown
BUILD_DATE ?= unknown
LDFLAGS = -X github.com/jhonma82/engineering-platform/internal/version.CoreVersion=$(VERSION) -X github.com/jhonma82/engineering-platform/internal/version.Commit=$(COMMIT) -X github.com/jhonma82/engineering-platform/internal/version.BuildDate=$(BUILD_DATE)

build:
	go build ./...

release:
	go build -ldflags "$(LDFLAGS)" ./cmd/eng

test:
	go test ./...

vet:
	go vet ./...

cover:
	go test -coverprofile=cover.out ./... && go tool cover -func=cover.out
