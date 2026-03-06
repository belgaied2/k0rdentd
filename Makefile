# Makefile for k0rdentd

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# k0s and k0rdent versions for airgap builds
K0S_VERSION ?= v1.32.8+k0s.0
K0RDENT_VERSION ?= 1.2.2
SKOPEO_VERSION ?= v1.17.0

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
build-airgap: download-k0s download-skopeo generate-metadata
	@echo "Building $(AIRGAP_BINARY_NAME) (airgap flavor)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -tags airgap $(AIRGAP_LDFLAGS) -o $(BINARY_DIR)/$(AIRGAP_BINARY_NAME) cmd/k0rdentd/main.go
	@echo "✓ Built $(BINARY_DIR)/$(AIRGAP_BINARY_NAME)"
	@echo "  K0s version: $(K0S_VERSION)"
	@echo "  Skopeo version: $(SKOPEO_VERSION)"
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

## download-skopeo: Download skopeo binary for embedding in airgap build
.PHONY: download-skopeo
download-skopeo:
	@echo "Downloading skopeo binary $(SKOPEO_VERSION) ($(GOARCH))..."
	@mkdir -p $(AIRGAP_ASSETS_DIR)/skopeo
	@curl -sSL -o $(AIRGAP_ASSETS_DIR)/skopeo/skopeo-$(SKOPEO_VERSION)-$(GOARCH) \
		https://github.com/lework/skopeo-binary/releases/download/$(SKOPEO_VERSION)/skopeo-linux-$(GOARCH)
	@chmod +x $(AIRGAP_ASSETS_DIR)/skopeo/skopeo-$(SKOPEO_VERSION)-$(GOARCH)
	@echo "✓ Downloaded skopeo binary to $(AIRGAP_ASSETS_DIR)/skopeo/"

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
	@rm -rf $(AIRGAP_ASSETS_DIR)/skopeo
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

## test-airgap-remote: Test airgap build on remote machines (3 servers)
## Usage: make test-airgap-remote SERVER1=10.0.0.1 SERVER2=10.0.0.2 SERVER3=10.0.0.3"
## Optional: SSH_USER=user SSH_KEY=~/.ssh/id_rsa
.PHONY: test-airgap-remote
test-airgap-remote: test-check-servers push-build remote-install-airgap-init remote-push-join-configs remote-install-airgap-join

test-check-servers:
	@# Validate we have 3 servers
	@if [ -z "$(SERVER1)" ] || [ -z "$(SERVER2)" ] || [ -z "$(SERVER3)" ]; then \
		echo "Error: Exactly 3 servers are required"; \
		echo "Provided: $(SERVER1) $(SERVER2) $(SERVER3)"; \
		exit 1; \
	fi
	@echo "=== Testing airgap build on remote machines ==="
	@echo "Server 1 (controller): $(SERVER1)"
	@echo "Server 2 (controller): $(SERVER2)"
	@echo "Server 3 (controller): $(SERVER3)"
	@echo "SSH User: $(or $(SSH_USER),$${USER})"


push-build:
	@echo ""
	@# Build airgap binary
	@echo ">>> Step 1: Building airgap binary..."
	$(MAKE) build-airgap
	@# Copy binary to all servers
	@echo ">>> Step 2: Copying binary to all servers..."
	@scp $(SSH_OPTS) bin/k0rdentd-airgap $(or $(SSH_USER),$${USER})@$(SERVER1):~/
	@scp $(SSH_OPTS) bin/k0rdentd-airgap $(or $(SSH_USER),$${USER})@$(SERVER2):~/
	@scp $(SSH_OPTS) bin/k0rdentd-airgap $(or $(SSH_USER),$${USER})@$(SERVER3):~/
	@echo ""
	@# Move binary to /usr/local/bin on each server
	@echo ">>> Step 3: Moving binary to /usr/local/bin on all servers..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER1) "sudo mv ~/k0rdentd-airgap /usr/local/bin/"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER2) "sudo mv ~/k0rdentd-airgap /usr/local/bin/"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER3) "sudo mv ~/k0rdentd-airgap /usr/local/bin/"
	@echo ""
	
remote-install-airgap-init:
	@# Install on first server and wait for completion
	@echo ">>> Step 4: Installing on server 1 (first controller)..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER1) "sudo /usr/local/bin/k0rdentd-airgap install"
	@echo ""
	@# Export join config from first server
	@echo ">>> Step 5: Exporting join config from server 1..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER1) "sudo /usr/local/bin/k0rdentd-airgap export-join-config --overwrite"
	@echo ""

remote-push-join-configs:
	@# Copy join config from server 1 to servers 2 and 3
	@echo ">>> Step 6: Copying join config to servers 2 and 3..."
	@# First copy to local temp, then distribute
	@scp $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER1):~/join-configs/controller-join.yaml /tmp/controller-join.yaml
	@scp $(SSH_OPTS) /tmp/controller-join.yaml $(or $(SSH_USER),$${USER})@$(SERVER2):~/
	@scp $(SSH_OPTS) /tmp/controller-join.yaml $(or $(SSH_USER),$${USER})@$(SERVER3):~/
	@rm -f /tmp/controller-join.yaml
	@echo ""
	@# Move join config to /etc/k0rdentd/ on servers 2 and 3
	@echo ">>> Step 7: Configuring join on servers 2 and 3..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER2) "sudo mkdir -p /etc/k0rdentd && sudo cp ~/controller-join.yaml /etc/k0rdentd/k0rdentd.yaml"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER3) "sudo mkdir -p /etc/k0rdentd && sudo cp ~/controller-join.yaml /etc/k0rdentd/k0rdentd.yaml"
	@echo ""

remote-install-airgap-join:
	@# Install on servers 2 and 3
	@echo ">>> Step 8: Installing on server 2 (joining cluster)..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER2) "sudo /usr/local/bin/k0rdentd-airgap install"
	@echo ""
	@echo ">>> Step 9: Installing on server 3 (joining cluster)..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER3) "sudo /usr/local/bin/k0rdentd-airgap install"
	@echo ""
	@echo ""
	@echo "=== Multi-node airgap deployment complete ==="
	@echo "Cluster should now have 3 controllers"
	@echo "Verify with: ssh $(or $(SSH_USER),$${USER})@$(SERVER1) 'sudo k0s kubectl get nodes'"

## test-airgap-remote-quick: Quick test on a single remote machine
## Usage: make test-airgap-remote-quick SERVER=10.0.0.1
.PHONY: test-airgap-remote-quick
test-airgap-remote-quick:
	@if [ -z "$(SERVER)" ]; then \
		echo "Error: SERVER parameter is required"; \
		echo "Usage: make test-airgap-remote-quick SERVER=\"ip\""; \
		exit 1; \
	fi
	@echo "=== Quick testing airgap build on $(SERVER) ==="
	@echo ">>> Building airgap binary..."
	$(MAKE) build-airgap
	@echo ">>> Copying binary to $(SERVER)..."
	@scp $(SSH_OPTS) bin/k0rdentd-airgap $(or $(SSH_USER),$${USER})@$(SERVER):~/
	@echo ">>> Moving binary to /usr/local/bin..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "sudo mv ~/k0rdentd-airgap /usr/local/bin/"
	@echo ">>> Running install..."
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "sudo /usr/local/bin/k0rdentd-airgap install"
	@echo "✓ Single-node airgap deployment complete"

.PHONY: test-clean-remote
test-clean-remote:
	@if [ -z "$(SERVER)" ]; then \
		echo "Error: SERVER parameter is required"; \
		exit 1; \
	fi
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "sudo /usr/local/bin/k0rdentd-airgap uninstall --force"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "sudo rm -rf /etc/k0s"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "sudo rm -rf /etc/k0rdentd"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "sudo rm -rf /usr/local/bin/k0*"
	@ssh $(SSH_OPTS) $(or $(SSH_USER),$${USER})@$(SERVER) "rm controller-join.yaml"

## help: Show this help message
.PHONY: help
help:
	@echo "k0rdentd Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

