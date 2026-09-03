.PHONY: build test clean fmt vet lint tidy

BIN_DIR := bin
BINARY := $(BIN_DIR)/patroy
GO := /usr/local/go/bin/go

GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
FULL_VERSION ?= $(shell ./scripts/build-id.sh 2>/dev/null || echo "0.1.0-dev")

LDFLAGS := -s -w \
	-X 'main.version=v$(FULL_VERSION)' \
	-X 'main.commit=$(GIT_SHA)' \
	-X 'main.date=$(BUILD_DATE)'

all: build

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/patroy

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
