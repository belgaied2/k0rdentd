# Implementation Plan: Embed Skopeo Binary

## Overview

This plan describes how to add skopeo to the embedded binaries in k0rdentd for air-gapped installations. Skopeo is required by the registry pusher to copy OCI images from the bundle to the local registry.

## Requirements

- Download skopeo from `lework/skopeo-binary` GitHub releases during `make build-airgap`
- Store in `internal/airgap/assets/skopeo` directory
- Extract to `/usr/bin/skopeo` during the **registry** phase (before pushing images)
- Support both amd64 and arm64 architectures

## Architecture

The skopeo binary is extracted during the `k0rdentd registry` command, not during `k0rdentd install`. This is because skopeo is needed to push images from the bundle to the local registry, which happens before the installation.

```mermaid
flowchart TD
    A[make build-airgap] --> B[download-k0s target]
    A --> C[download-skopeo target]
    B --> D[Store in internal/airgap/assets/k0s]
    C --> E[Store in internal/airgap/assets/skopeo]
    D --> F[Build with -tags airgap]
    E --> F
    F --> G[assets.go embeds both binaries]
    
    subgraph Registry Phase
        G --> H[registry daemon Start]
        H --> I[ExtractSkopeoBinary]
        I --> J[/usr/bin/skopeo]
        J --> K[Push images to local registry]
    end
    
    subgraph Install Phase
        L[installer Install] --> M[ExtractK0sBinary]
        M --> N[/usr/local/bin/k0s]
    end
```

## Implementation Details

### 1. Makefile Changes

Add skopeo version variable and download target:

```makefile
# Skopeo version for airgap builds
SKOPEO_VERSION ?= v1.16.1

## download-skopeo: Download skopeo binary for embedding in airgap build
.PHONY: download-skopeo
download-skopeo:
	@echo "Downloading skopeo binary $(SKOPEO_VERSION) ($(GOARCH))..."
	@mkdir -p $(AIRGAP_ASSETS_DIR)/skopeo
	@curl -sSL -o $(AIRGAP_ASSETS_DIR)/skopeo/skopeo-$(SKOPEO_VERSION)-$(GOARCH) \
		https://github.com/lework/skopeo-binary/releases/download/$(SKOPEO_VERSION)/skopeo-linux-$(GOARCH)
	@chmod +x $(AIRGAP_ASSETS_DIR)/skopeo/skopeo-$(SKOPEO_VERSION)-$(GOARCH)
	@echo "✓ Downloaded skopeo binary to $(AIRGAP_ASSETS_DIR)/skopeo/"
```

Update `build-airgap` target to depend on both downloads:

```makefile
build-airgap: download-k0s download-skopeo generate-metadata
```

Update `clean-airgap` target:

```makefile
clean-airgap:
	@rm -rf $(AIRGAP_ASSETS_DIR)/skopeo
	# ... existing cleanup
```

### 2. assets.go Changes

Add skopeo embedding alongside k0s:

```go
// SkopeoBinary embeds the skopeo binary for the target platform
// The binary is placed in internal/airgap/assets/skopeo/ during build
//
//go:embed skopeo/*
var SkopeoBinary embed.FS
```

### 3. stub.go Changes

Add stub for non-airgap builds:

```go
// SkopeoBinary is a stub for non-airgap builds
var SkopeoBinary fs.FS = emptyFS{}
```

### 4. registry/daemon.go Changes

Add extraction function and call it in the Start method:

```go
// ExtractSkopeoBinary extracts the embedded skopeo binary to /usr/bin/skopeo
func ExtractSkopeoBinary() error {
    if !airgap.IsAirGap() {
        return fmt.Errorf("not an airgap build, cannot extract embedded skopeo binary")
    }

    // Read the skopeo directory from embedded FS
    entries, err := fs.ReadDir(assets.SkopeoBinary, "skopeo")
    if err != nil {
        return fmt.Errorf("failed to read embedded skopeo directory: %w", err)
    }
    // ... extraction logic
}
```

Call it in the Start method before pushing images:

```go
func (r *RegistryDaemon) Start(ctx context.Context) error {
    // Step 1: Extract skopeo binary if this is an airgap build
    if airgap.IsAirGap() {
        if err := ExtractSkopeoBinary(); err != nil {
            return fmt.Errorf("failed to extract skopeo binary: %w", err)
        }
    }
    // ... rest of Start method
}
```

### 5. README.md Updates

Update `internal/airgap/assets/README.md` to document skopeo.

## Files Modified

| File | Change |
|------|--------|
| `Makefile` | Add SKOPEO_VERSION, download-skopeo target, update build-airgap and clean-airgap |
| `internal/airgap/assets/assets.go` | Add SkopeoBinary embed.FS variable |
| `internal/airgap/assets/stub.go` | Add SkopeoBinary stub |
| `internal/airgap/registry/daemon.go` | Add ExtractSkopeoBinary function, call it in Start method |
| `internal/airgap/assets/README.md` | Document skopeo |

## Testing

1. Run `make clean-airgap`
2. Run `make build-airgap`
3. Verify skopeo binary is downloaded to `internal/airgap/assets/skopeo/`
4. Run `sudo ./bin/k0rdentd-airgap registry -b /path/to/bundle.tar.gz`
5. Verify skopeo is extracted to `/usr/bin/skopeo`
6. Run `skopeo --version` to verify functionality

## Notes

- The `lework/skopeo-binary` releases use naming format: `skopeo-linux-{arch}`
- Version example: `v1.16.1`
- Skopeo is extracted during the **registry** phase, not the install phase
- This is because skopeo is needed to push images from the bundle to the local registry
- The extraction path `/usr/bin/skopeo` ensures skopeo is available system-wide for the registry pusher
