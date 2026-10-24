.PHONY: help help-all setup build install test types lint check

.DEFAULT_GOAL := help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "  check    - Lint, vet, and test"
	@echo "  build    - Build gerrit-cli binary"
	@echo "  install  - Build and install to ~/.local/bin"
	@echo "  test     - Run tests"
	@echo ""
	@echo "Run 'make help-all' for the full list, including internal targets."

help-all:
	@echo "Public targets:"
	@echo "  check    - Lint, vet, and test"
	@echo "  build    - Build gerrit-cli binary"
	@echo "  install  - Build and install to ~/.local/bin"
	@echo "  test     - Run tests"
	@echo ""
	@echo "Internal targets (typically called via check or build):"
	@echo "  lint     - Run golangci-lint"
	@echo "  types    - Run go vet"
	@echo "  setup    - Download go modules"

# Version reported by --version. Outside a checkout the binary uses "dev".
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w
ifneq ($(VERSION),dev)
  LDFLAGS += -X gerrit-cli/cmd/app/command.version=$(VERSION)
endif

setup:
	go mod download

build:
	@echo "Building gerrit-cli $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o target/gerrit-cli ./cmd/app
	@echo "Build complete: target/gerrit-cli"

install: build
	@install_dir="$${HOME}/.local/bin"; \
	mkdir -p "$$install_dir"; \
	cp target/gerrit-cli "$$install_dir/gerrit-cli"; \
	chmod +x "$$install_dir/gerrit-cli"; \
	echo "Installed: $$install_dir/gerrit-cli"

test:
	go test ./...

types:
	go vet ./...

lint:
	golangci-lint run ./...

check: lint types test
	@echo "ALL CHECKS PASSED"
