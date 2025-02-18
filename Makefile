# eth-beacon-genesis
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
VERSION := $(shell git rev-parse --short HEAD)

GOLDFLAGS += -X 'github.com/ethpandaops/eth-beacon-genesis/utils.BuildVersion="$(VERSION)"'
GOLDFLAGS += -X 'github.com/ethpandaops/eth-beacon-genesis/utils.Buildtime="$(BUILDTIME)"'
GOLDFLAGS += -X 'github.com/ethpandaops/eth-beacon-genesis/utils.BuildRelease="$(RELEASE)"'

.PHONY: all test clean

all: test build

test:
	go test ./...

build:
	@echo version: $(VERSION)
	env CGO_ENABLED=1 go build -v -o bin/ -ldflags="-s -w $(GOLDFLAGS)" ./cmd/*

clean:
	rm -f bin/*
	$(MAKE) -C ui-package clean
