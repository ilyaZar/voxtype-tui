BINARY := voxtype-tui
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
LDFLAGS ?= -X main.version=$(VERSION)

.PHONY: all build install test fmt tidy clean

all: build

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/voxtype-tui

install: build
	mkdir -p $(BINDIR)
	install -m 0755 bin/$(BINARY) $(BINDIR)/$(BINARY)

test:
	go test ./...

fmt:
	gofmt -w cmd internal

tidy:
	go mod tidy

clean:
	rm -rf bin dist coverage.out
