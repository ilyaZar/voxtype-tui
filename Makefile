BINARY := voxtype-tui
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

.PHONY: all build install test fmt tidy clean

all: build

build:
	go build -o bin/$(BINARY) ./cmd/voxtype-tui

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
