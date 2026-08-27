# Knowledge Forge — build, test, release.
#
# Two lanes, and the difference between them is the whole reason this file is not three
# lines. `portable` is CGO_ENABLED=0 and cross-compiles to every target from any host;
# `full` is CGO_ENABLED=1 and carries go-tree-sitter, so it can only be built where a C
# toolchain for the target exists. pkg/codeindex is behind a build tag precisely so the
# portable binary still runs — it reports the code index as unavailable rather than
# claiming an undocumented codebase is fully documented.

BIN      := forge
PKG      := ./cmd/forge
DIST     := dist
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS  := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: all build full test bench vet fmt lint dist checksums install-hook clean help

all: fmt vet test build

## build: the portable binary for this host (no cgo, no code index)
build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BIN) $(PKG)

## full: this host's binary with the tree-sitter code index compiled in
full:
	CGO_ENABLED=1 go build -ldflags '$(LDFLAGS)' -o $(BIN) $(PKG)

## test: the whole suite, cgo on so pkg/codeindex's parser is actually exercised
test:
	CGO_ENABLED=1 go test ./...
	CGO_ENABLED=0 go build ./...

## bench: parse, similarity and drift — the three the phase brief names
bench:
	go test ./pkg/vault ./pkg/similarity ./pkg/drift ./pkg/codeindex ./pkg/gitsig \
		./pkg/linkcheck -run '^$$' -bench . -benchmem

vet:
	go vet ./...

fmt:
	gofmt -l -w pkg cmd

## lint: gofmt as a gate rather than a rewrite, for CI
lint:
	@out=$$(gofmt -l pkg cmd); if [ -n "$$out" ]; then echo "gofmt:"; echo "$$out"; exit 1; fi
	go vet ./...

## dist: the portable matrix — four targets, reproducible from any host
dist: clean
	@mkdir -p $(DIST)
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		echo "  $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
			go build -ldflags '$(LDFLAGS)' -o $(DIST)/$(BIN)-$$os-$$arch $(PKG) || exit 1; \
	done
	@$(MAKE) --no-print-directory checksums

## checksums: what bin/forge verifies before it executes anything
checksums:
	@cd $(DIST) && shasum -a 256 $(BIN)-* > checksums.txt && cat checksums.txt

## install-hook: copy this build to where the vault's post-commit hook looks for it
install-hook: full
	@mkdir -p $$HOME/.forge/bin
	@cp $(BIN) $$HOME/.forge/bin/$(BIN)
	@shasum -a 256 $$HOME/.forge/bin/$(BIN) | awk '{print $$1}' > $$HOME/.forge/bin/$(BIN).sha256
	@echo "installed $$HOME/.forge/bin/$(BIN) — pin written to $(BIN).sha256"

clean:
	@rm -rf $(DIST) $(BIN)

help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
