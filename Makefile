.PHONY: build test clean fmt vet lint tidy

BIN_DIR := bin
BINARY := $(BIN_DIR)/patroy
GO := /usr/local/go/bin/go

all: build

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/patroy

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
