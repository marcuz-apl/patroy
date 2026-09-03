.PHONY: build test clean fmt vet lint tidy

BIN_DIR := bin
BINARY := $(BIN_DIR)/patroy
GO ?= $(shell which go 2>/dev/null || echo /usr/local/go/bin/go)

GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
FULL_VERSION ?= $(shell (./.githooks/alfazen-version id 2>/dev/null || ([ -f VERSION ] && cat VERSION)) || echo "0.4.0")

LDFLAGS := -s -w \
	-X 'main.version=v$(FULL_VERSION)' \
	-X 'main.commit=$(GIT_SHA)' \
	-X 'main.date=$(BUILD_DATE)'

all: build

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/patroy

build-windows:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/patroy.exe ./cmd/patroy

build-all: build build-windows
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/patroy_darwin_arm64 ./cmd/patroy
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/patroy_darwin_amd64 ./cmd/patroy

test:
	$(GO) test -v -race ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)
