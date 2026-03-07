VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/nownow-labs/nownow/cmd.Version=$(VERSION)
BINARY := nownow

SWIFT_HELPER := nowplaying-helper
SWIFT_SRC := internal/detect/nowplaying/main.swift
UNAME_S := $(shell uname -s)

.PHONY: build test lint clean build-all release-local checksums swift-helper

ifeq ($(UNAME_S),Darwin)
build: swift-helper
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .
else
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .
endif

swift-helper:
	swiftc -O -o $(SWIFT_HELPER) $(SWIFT_SRC)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf dist/ $(BINARY) $(SWIFT_HELPER)

build-all: clean
	mkdir -p dist
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_darwin_amd64/$(BINARY) .
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_darwin_arm64/$(BINARY) .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_linux_amd64/$(BINARY) .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_linux_arm64/$(BINARY) .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_windows_amd64/$(BINARY).exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_windows_arm64/$(BINARY).exe .

release-local:
	goreleaser release --snapshot --clean

checksums:
	cd dist && shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt || true
