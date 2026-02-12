# Makefile for k0rdentd

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# k0s and k0rdent versions for airgap builds
K0S_VERSION ?= v1.32.8+k0s.0
K0RDENT_VERSION ?= 1.2.2

# Build directories
BINARY_DIR := bin
AIRGAP_DIR := build/airgap
AIRGAP_ASSETS_DIR := internal/airgap/assets

# Architecture detection
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
GOARCH := amd64
else ifeq ($(ARCH),aarch64)
GOARCH := arm64
else
GOARCH := $(ARCH)
endif

# Build flags
LDFLAGS := -ldflags "-X github.com/belgaied2/k0rdentd/internal/airgap.Version=$(VERSION) \
                      -X github.com/belgaied2/k0rdentd/internal/airgap.Flavor=online \
                      -X github.com/belgaied2/k0rdentd/internal/airgap.BuildTime=$(BUILD_TIME)"

AIRGAP_LDFLAGS := -ldflags "-X github.com/belgaied2/k0rdentd/internal/airgap.Version=$(VERSION) \
                            -X github.com/belgaied2/k0rdentd/internal/airgap.Flavor=airgap \
                            -X github.com/belgaied2/k0rdentd/internal/airgap.K0sVersion=$(K0S_VERSION) \
                            -X github.com/belgaied2/k0rdentd/internal/airgap.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Main binary name
BINARY_NAME := k0rdentd
AIRGAP_BINARY_NAME := k0rdentd-airgap

.PHONY: all
all: build

## build: Build the online version of k0rdentd
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) (online flavor)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) cmd/k0rdentd/main.go
	@echo "✓ Built $(BINARY_DIR)/$(BINARY_NAME)"

## build-airgap: Build the airgap version of k0rdentd
.PHONY: build-airgap
build-airgap: download-k0s
	@echo "Building $(AIRGAP_BINARY_NAME) (airgap flavor)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -tags airgap $(AIRGAP_LDFLAGS) -o $(BINARY_DIR)/$(AIRGAP_BINARY_NAME) cmd/k0rdentd/main.go
	@echo "✓ Built $(BINARY_DIR)/$(AIRGAP_BINARY_NAME)"
	@echo "  K0s version: $(K0S_VERSION)"
	@echo "  Architecture: $(GOARCH)"

## download-k0s: Download k0s binary for embedding in airgap build
.PHONY: download-k0s
download-k0s:
	@echo "Downloading k0s binary $(K0S_VERSION) ($(GOARCH))..."
	@mkdir -p $(AIRGAP_ASSETS_DIR)/k0s
	@curl -sSL -o $(AIRGAP_ASSETS_DIR)/k0s/k0s-$(K0S_VERSION)-$(GOARCH) \
		https://github.com/k0sproject/k0s/releases/download/$(K0S_VERSION)/k0s-$(K0S_VERSION)-$(GOARCH)
	@chmod +x $(AIRGAP_ASSETS_DIR)/k0s/k0s-$(K0S_VERSION)-$(GOARCH)
	@echo "✓ Downloaded k0s binary to $(AIRGAP_ASSETS_DIR)/k0s/"

## generate-metadata: Generate metadata.json for airgap build
.PHONY: generate-metadata
generate-metadata:
	@echo "Generating airgap metadata..."
	@mkdir -p $(AIRGAP_ASSETS_DIR)
	@echo '{"flavor":"airgap","k0sVersion":"$(K0S_VERSION)","buildTime":"$(BUILD_TIME)"}' \
		> $(AIRGAP_ASSETS_DIR)/metadata.json
	@echo "✓ Generated metadata.json"

## prepare-airgap-assets: Prepare all assets for airgap build
.PHONY: prepare-airgap-assets
prepare-airgap-assets: download-k0s generate-metadata
	@echo "✓ All airgap assets prepared"

## download-bundles: Download k0rdent enterprise airgap bundles (user helper)
.PHONY: download-bundles
download-bundles:
	@echo "Note: Download k0rdent enterprise bundles from:"
	@echo "  https://get.mirantis.com/k0rdent-enterprise/$(K0RDENT_VERSION)/"
	@echo ""
	@echo "Example:"
	@echo "  wget https://get.mirantis.com/k0rdent-enterprise/$(K0RDENT_VERSION)/airgap-bundle-$(K0RDENT_VERSION).tar.gz"
	@echo "  wget https://get.mirantis.com/k0rdent-enterprise/$(K0RDENT_VERSION)/airgap-bundle-$(K0RDENT_VERSION).tar.gz.sig"
	@echo ""
	@echo "Verify with cosign:"
	@echo "  cosign verify-blob --key https://get.mirantis.com/cosign.pub \\"
	@echo "    --signature airgap-bundle-$(K0RDENT_VERSION).tar.gz.sig \\"
	@echo "    airgap-bundle-$(K0RDENT_VERSION).tar.gz"

## clean: Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@rm -rf build
	@echo "✓ Cleaned"

## clean-airgap: Clean only airgap build artifacts
.PHONY: clean-airgap
clean-airgap:
	@echo "Cleaning airgap artifacts..."
	@rm -rf $(AIRGAP_DIR)
	@rm -rf $(AIRGAP_ASSETS_DIR)/k0s
	@rm -f $(AIRGAP_ASSETS_DIR)/metadata.json
	@rm -f $(BINARY_DIR)/$(AIRGAP_BINARY_NAME)
	@echo "✓ Cleaned airgap artifacts"

## test: Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## deps: Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## fmt: Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## lint: Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi

## help: Show this help message
.PHONY: help
help:
	@echo "k0rdentd Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
