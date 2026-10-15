.PHONY: setup build install test types lint check

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
