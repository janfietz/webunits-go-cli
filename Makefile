BINARY_NAME := webuntis
MODULE := github.com/janfietz/webunits-go-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X $(MODULE)/pkg/cli.Version=$(VERSION)"

.PHONY: build test lint clean install release-dry-run

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/webuntis

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

install:
	go install $(LDFLAGS) ./cmd/webuntis

release-dry-run:
	goreleaser --snapshot --clean
