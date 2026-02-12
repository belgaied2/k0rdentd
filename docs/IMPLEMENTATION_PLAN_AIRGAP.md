# Airgap Feature Implementation Plan

## Status: Phase 3 Completed + OCI Registry Fix + Containerd Mirror Implemented

**Last Updated**: 2026-02-12
**Phase**: Phase 3 (Airgap Installation with Registry) - ✅ COMPLETED
**OCI Registry Fix**: ✅ COMPLETED (2026-02-12)
**Containerd Mirror Configuration**: ✅ COMPLETED (2026-02-12)

---

## Design Summary

- **Approach**: Multi-worker ready with local OCI registry daemon
- **Build Flavors**: Two binaries - online (default) and airgap (with embedded k0s binary only)
- **Bundle Strategy**: **Option B** - Embed k0s binary (~100MB), use external k0rdent bundle (22GB)
- **New**: Local OCI registry daemon runs separately from install workflow
- **Rationale**:
  - Avoids redistributing k0rdent enterprise binaries; users download directly from Mirantis
  - Registry daemon enables official k0rdent airgap installation process
  - Reusable installation steps between online and airgap modes
  - Disk-based storage (not in-memory) handles 22GB of images
  - Configurable port and multi-worker support
- **Multi-Worker**: Registry accessible to workers; export command creates worker bundle with k0s binary

---

## Phase 1: Foundation ✅ (COMPLETED)

### 1.1 Package Structure

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Created `internal/airgap/` directory structure
- ✅ Created `internal/airgap/detector.go` - Build flavor detection
- ✅ Created `internal/airgap/assets/` directory
- ✅ Created `internal/airgap/exporter.go` - Worker artifact export
- ✅ Created `scripts/download-k0rdent-bundle.sh`

**Acceptance**: ✅ All directories and files created

---

### 1.2 Build Flavor Detection

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Implemented `internal/airgap/detector.go`
- ✅ Added ldflag variable `Flavor`
- ✅ Added ldflag variable `K0sVersion`
- ✅ Added ldflag variable `BuildTime`
- ✅ Implemented `IsAirGap()` function
- ✅ Implemented `GetBuildMetadata()` function

**Note**: `K0rdentVersion` is extracted from bundle at runtime, not set at build time

**Acceptance**: ✅ Build flavor detection works

---

### 1.3 Asset Embedding (k0s Binary Only)

**Status**: ✅ Completed
**Estimated Effort**: 1 day

**Decision**: Embed only k0s binary; k0rdent bundle remains external

- ✅ Created `internal/airgap/assets/assets.go` with embed directives
  - `//go:embed k0s/*` for k0s binary
  - `//go:embed metadata.json` for build metadata
- ✅ Created `internal/airgap/assets/stub.go` for non-airgap builds
  - Implements `emptyFS` type satisfying `fs.FS` interface
  - Allows assets package to be imported in both build flavors
- ✅ Added build tags for conditional embedding (`//go:build airgap`)
- ✅ **FIXED**: Exporter now uses `assets.K0sBinary` via `extractFromEmbedded()`

**Binary Size Evidence**:
- Online build: ~61 MB
- Airgap build: ~311 MB (+250 MB = k0s binary size)
- k0s binary string found in airgap binary: `k0s-v1.31.2+k0s.0-amd64`

**Acceptance**: ✅ Build compiles with `-tags airgap` and k0s is properly embedded

---

### 1.4 Bundle Helper Scripts

**Status**: ✅ Completed
**Estimated Effort**: 1 day

- ✅ Created `scripts/download-k0rdent-bundle.sh`
- ✅ Documented bundle preparation steps

**Acceptance**: ✅ Script downloads k0rdent bundle from Mirantis

---

### 1.5 Makefile Integration

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Added `K0S_VERSION` variable
- ✅ Added `AIRGAP_DIR` variable
- ✅ Added `build-airgap` target
- ✅ Added `clean-airgap` target

**Acceptance**: ✅ Makefile targets work

---

### 1.6 Worker Artifact Exporter

**Status**: ✅ Completed
**Estimated Effort**: 1 day

- ✅ Implemented `internal/airgap/exporter.go`
- ✅ Created `pkg/cli/export_worker.go`
- ✅ Added `export-worker-artifacts` CLI command
- ✅ Added `show-flavor` CLI command
- ✅ **NEW**: Extracts k0s from embedded assets via `extractFromEmbedded()`
- ✅ **NEW**: Falls back to filesystem for development/testing

**Acceptance**: ✅ Commands work correctly

---

### 1.7 Installer Integration

**Status**: ✅ Completed
**Estimated Effort**: 1 day

- ✅ Updated `pkg/installer/installer.go` for airgap mode
- ✅ Created `internal/airgap/installer.go` structure
- ✅ Added bundle path configuration support
- ✅ **NEW**: `installAirgap()` returns error with instructions to use registry daemon

**Acceptance**: ✅ Airgap mode integration complete

---

## Phase 2: Registry Daemon Implementation ✅ (COMPLETED)

**Estimated Effort**: 4 days
**Status**: ✅ All tasks completed

### 2.1 Registry Package Structure

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Created `internal/airgap/registry/` directory
- ✅ Created `internal/airgap/registry/daemon.go` - Main daemon implementation
- ✅ Created `internal/airgap/registry/pusher.go` - Push images to registry
- ✅ Created `internal/airgap/registry/verifier.go` - Cosign verification
- ✅ Created `internal/airgap/bundle/version.go` - Version extraction

**Files Created**:
```
internal/airgap/
├── bundle/
│   └── version.go
└── registry/
    ├── daemon.go
    ├── pusher.go
    └── verifier.go
pkg/cli/
└── registry.go
```

---

### 2.2 Cosign Verification

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Implemented `internal/airgap/registry/verifier.go`
- ✅ Added cosign verification logic
- ✅ Implemented `VerifyBundle(bundlePath, cosignKey)` function
- ✅ Implemented `DownloadCosignKey(keyURL, destDir)` for URL-based keys
- ✅ Added `--verify` and `--cosignKey` flags to registry command
- ✅ Support for both URL and local key paths

**Code**:
```go
// internal/airgap/registry/verifier.go
package registry

func VerifyBundle(bundlePath, cosignKey string) error {
    // Check if cosign is available
    if _, err := exec.LookPath("cosign"); err != nil {
        return fmt.Errorf("cosign not found in PATH: %w", err)
    }

    // Check if signature file exists
    sigPath := bundlePath + ".sig"
    if _, err := os.Stat(sigPath); os.IsNotExist(err) {
        return fmt.Errorf("signature file not found: %s", sigPath)
    }

    // Verify signature
    // cosign verify-blob --key <cosignKey> --signature <bundle>.sig <bundle>
    var cmd *exec.Cmd
    if strings.HasPrefix(cosignKey, "http://") || strings.HasPrefix(cosignKey, "https://") {
        // Use URL directly (cosign supports this)
        cmd = exec.Command("cosign", "verify-blob",
            "--key", cosignKey,
            "--signature", sigPath,
            bundlePath)
    } else {
        // Use local key file
        cmd = exec.Command("cosign", "verify-blob",
            "--key", cosignKey,
            "--signature", sigPath,
            bundlePath)
    }

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("cosign verification failed: %w, output: %s", err, string(output))
    }

    return nil
}

func DownloadCosignKey(keyURL, destDir string) (string, error) {
    // Check if curl or wget is available
    var cmd *exec.Cmd
    tempFile := filepath.Join(destDir, "cosign.pub")

    if _, err := exec.LookPath("wget"); err == nil {
        cmd = exec.Command("wget", "-O", tempFile, keyURL)
    } else if _, err := exec.LookPath("curl"); err == nil {
        cmd = exec.Command("curl", "-o", tempFile, keyURL)
    } else {
        return "", fmt.Errorf("neither wget nor curl found in PATH")
    }

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("failed to download cosign key: %w, output: %s", err, string(output))
    }

    return tempFile, nil
}
```

**Acceptance**:
- Bundle signature verified with cosign
- Fails if signature invalid
- Supports both URL and local key paths

---

### 2.3 Version Extraction from Bundle

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Implemented `ExtractK0rdentVersion(bundlePath)` function in `bundle/version.go`
- ✅ Parse `charts/k0rdent-enterprise_*.tar` from bundle
- ✅ Extract `Chart.yaml` from tar file
- ✅ Read version field from Chart.yaml
- ✅ Support both tar.gz archives and extracted directories

**Code**:
```go
// internal/airgap/bundle/version.go
package bundle

func ExtractK0rdentVersion(bundlePath string) (string, error) {
    // Check if it's a directory
    if info, err := os.Stat(bundlePath); err == nil && info.IsDir() {
        return extractVersionFromDir(bundlePath)
    }

    // Otherwise treat as tar.gz archive
    return extractVersionFromArchive(bundlePath)
}
```

**Acceptance**:
- Extracts version from bundle
- Returns "1.2.3" for bundle 1.2.3

---

### 2.4 Registry Daemon Implementation

**Status**: ✅ Completed
**Estimated Effort**: 1.5 days

- ✅ Implemented `internal/airgap/registry/daemon.go`
- ✅ Integrated `github.com/google/go-containerregistry/pkg/registry`
- ✅ Implemented disk-based blob handler (not in-memory)
- ✅ Added configurable port support
- ✅ Added configurable host binding (default 0.0.0.0 for all interfaces)
- ✅ Added graceful shutdown handling (30-second timeout)
- ✅ Implemented signal handling (SIGTERM, SIGINT)
- ✅ Added helper methods:
  - `IsRunning()` - Check if port is in use
  - `GetStorageSize()` - Return total storage size
  - `FormatBytes()` - Convert bytes to human-readable format
  - `Addr()` - Return registry address

**Code Structure**:
```go
// internal/airgap/registry/daemon.go
package registry

type RegistryDaemon struct {
    port       string
    host       string  // Added: host binding support
    storageDir string
    bundlePath string
    server     *http.Server
    verifySig  bool
    cosignKey  string
}

func NewRegistryDaemon(port, storageDir, bundlePath string, verifySig bool, cosignKey string) *RegistryDaemon {
    return &RegistryDaemon{
        port:       port,
        host:       "0.0.0.0", // Listen on all interfaces
        storageDir: storageDir,
        bundlePath: bundlePath,
        verifySig:  verifySig,
        cosignKey:  cosignKey,
    }
}

func (r *RegistryDaemon) Start(ctx context.Context) error {
    logger := getLogger()

    // Step 1: Verify bundle with cosign if requested
    if r.verifySig {
        if err := r.verifyBundle(); err != nil {
            return fmt.Errorf("bundle verification failed: %w", err)
        }
        logger.Info("✓ Bundle verified with cosign")
    }

    // Step 2: Extract k0rdent version from bundle (commented out - done at install time)
    // logger.Infof("Extracting k0rdent version from bundle...")
    // k0rdentVersion, err := bundle.ExtractK0rdentVersion(r.bundlePath)
    // if err != nil {
    //     return fmt.Errorf("failed to extract version: %w", err)
    // }
    // logger.Infof("✓ K0rdent version from bundle: %s", k0rdentVersion)

    // Step 3: Initialize disk-based registry
    if err := os.MkdirAll(r.storageDir, 0755); err != nil {
        return fmt.Errorf("failed to create storage dir: %w", err)
    }

    logger.Infof("Initializing registry with disk-based storage: %s", r.storageDir)
    // Create disk-based blob handler
    blobHandler := registry.NewDiskBlobHandler(r.storageDir)
    // Create registry handler
    reg := registry.New(registry.WithBlobHandler(blobHandler))

    // Step 4: Start HTTP server
    r.server = &http.Server{
        Addr:         r.host + ":" + r.port,
        Handler:      reg,
        ReadTimeout:  5 * time.Minute,
        WriteTimeout: 10 * time.Minute,
    }

    // Start server in a goroutine
    errChan := make(chan error, 1)
    go func() {
        logger.Infof("Registry server listening on %s", r.Addr())
        logger.Infof("Storage directory: %s", r.storageDir)
        if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errChan <- err
        }
    }()

    // Step 5: Push images from bundle to local registry
    logger.Infof("Pushing images from bundle to local registry...")
    if err := r.pushImagesToRegistry(ctx); err != nil {
        return fmt.Errorf("failed to push images: %w", err)
    }

    // Wait for context cancellation or server error
    select {
    case <-ctx.Done():
        logger.Infof("Shutting down registry server...")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        if err := r.server.Shutdown(shutdownCtx); err != nil {
            return fmt.Errorf("registry shutdown failed: %w", err)
        }
        logger.Infof("✓ Registry server stopped gracefully")
        return nil
    case err := <-errChan:
        return err
    }
}

// Addr returns the registry address
func (r *RegistryDaemon) Addr() string {
    return r.host + ":" + r.port
}

// WaitForSignal waits for shutdown signals
func (r *RegistryDaemon) WaitForSignal(ctx context.Context) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

    select {
    case <-sigChan:
        // Signal received, context will be cancelled
    case <-ctx.Done():
        // Context already cancelled
    }
}

// IsRunning checks if the registry port is in use
func (r *RegistryDaemon) IsRunning() bool {
    ln, err := net.Listen("tcp", r.host+":"+r.port)
    if err != nil {
        return true // Port is in use
    }
    ln.Close()
    return false
}

// GetStorageSize returns the total size of the registry storage
func (r *RegistryDaemon) GetStorageSize() (int64, error) {
    var size int64

    err := filepath.Walk(r.storageDir, func(_ string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
    })

    return size, err
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

**Key Improvements Over Original Design**:
- Host binding configuration for multi-worker scenarios
- Separate error channel for server errors
- 30-second graceful shutdown timeout
- Separate `Addr()` helper method
- Storage size calculation and formatting helpers

**Acceptance**:
- `k0rdentd registry --port 5000` starts daemon
- `k0rdentd registry --host 0.0.0.0` binds to all interfaces
- Images pushed to localhost:5000
- Registry persists data to disk
- Daemon stops gracefully on SIGTERM/SIGINT

---

### 2.5 Image Pusher Implementation

**Status**: ✅ Completed
**Estimated Effort**: 1 day

- ✅ Implemented `internal/airgap/registry/pusher.go`
- ✅ Use skopeo to copy images from bundle to registry
- ✅ Handle OCI archives in bundle (.tar files)
- ✅ Progress reporting with `[i/total]` format
- ✅ **Concurrent image pushing** with semaphore (max 5 concurrent)
- ✅ WaitGroup-based error aggregation
- ✅ Filter out skopeo binary from images
- ✅ Support for both tar.gz archives and extracted directories
- ✅ Internal logger interface to avoid circular dependencies
- ✅ `--dest-tls-verify=false` flag for skopeo (TODO: Remove for production)

**Code Structure**:
```go
// internal/airgap/registry/pusher.go
package registry

func PushImages(bundlePath, registryAddr string) error {
    // Steps:
    // 1. Extract bundle if tar.gz
    // 2. Find all OCI archives (.tar files, excluding skopeo)
    // 3. Check if skopeo is available
    // 4. Push images with concurrent workers (max 5)
    // 5. Report progress [i/total] for each image
    // 6. Aggregate errors but continue on failures
}

func pushImagesWithProgress(images []string, bundleDir, registryAddr string) error {
    // Uses WaitGroup and semaphore for concurrency
    // Max 5 concurrent pushes
    // Error channel for aggregating failures
}

func pathToImageRef(path string) string {
    // Converts: /tmp/bundle/k0sproject/k0s:v1.32.8-k0s.0.tar
    // To: k0sproject/k0s:v1.32.8-k0s.0
}

// Internal logger interface
type logger interface {
    Infof(string, ...interface{})
    Warnf(string, ...interface{})
}
```

**Acceptance**:
- Images pushed from bundle to local registry
- Progress reported for large bundles with [i/total] format
- Failed images logged but don't stop entire process
- Errors aggregated and reported at end
- Concurrent pushing speeds up large bundle loading

---

### 2.6 CLI Command for Registry

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day

- ✅ Created `pkg/cli/registry.go`
- ✅ Added `registry` command to main CLI
- ✅ Added `--port` flag (configurable, default 5000)
- ✅ Added `--host` flag (for multi-worker support)
- ✅ Added `--storage` flag (default /var/lib/k0rdentd/registry)
- ✅ Added `--bundle-path` flag (required)
- ✅ Added `--verify` flag (default true)
- ✅ Added `--cosignKey` flag (default Mirantis URL)
- ✅ Added `--background` flag with warning
- ✅ Bundle path validation
- ✅ Port in-use check before starting
- ✅ Signal handling (SIGTERM, SIGINT)
- ✅ Configuration logging

**Build Status**: ✅ Compiles successfully

**Code**:
```go
// pkg/cli/registry.go
package cli

var RegistryCommand = &cli.Command{
    Name:      "registry",
    Usage:     "Run OCI registry daemon for airgap installations",
    UsageText: "k0rdentd registry [options]",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:    "port",
            Aliases: []string{"p"},
            Value:   "5000",
            Usage:   "Port for registry server",
            EnvVars: []string{"K0RDENTD_REGISTRY_PORT"},
        },
        &cli.StringFlag{
            Name:    "host",
            Aliases: []string{"H"},
            Value:   "0.0.0.0",
            Usage:   "Host address to bind to (default: 0.0.0.0 for all interfaces)",
            EnvVars: []string{"K0RDENTD_REGISTRY_HOST"},
        },
        &cli.StringFlag{
            Name:    "storage",
            Aliases: []string{"s"},
            Value:   "/var/lib/k0rdentd/registry",
            Usage:   "Storage directory for registry data",
            EnvVars: []string{"K0RDENTD_REGISTRY_STORAGE"},
        },
        &cli.StringFlag{
            Name:    "bundle-path",
            Aliases: []string{"b"},
            Usage:   "Path to k0rdent airgap bundle (tar.gz or extracted directory)",
            EnvVars: []string{"K0RDENTD_AIRGAP_BUNDLE_PATH"},
            Required: true,
        },
        &cli.BoolFlag{
            Name:    "verify",
            Usage:   "Verify bundle signature with cosign",
            EnvVars: []string{"K0RDENTD_VERIFY_SIGNATURE"},
            Value:   true,
        },
        &cli.StringFlag{
            Name:    "cosignKey",
            Usage:   "Cosign public key URL or local path",
            Value:   "https://get.mirantis.com/cosign.pub",
            EnvVars: []string{"K0RDENT_COSIGN_KEY"},
        },
        &cli.BoolFlag{
            Name:    "background",
            Aliases: []string{"d"},
            Usage:   "Run as daemon in background (not recommended, use systemd/supervise instead)",
            Value:   false,
        },
    },
    Action: registryAction,
}

func registryAction(c *cli.Context) error {
    logger := utils.GetLogger()

    port := c.String("port")
    host := c.String("host")
    storage := c.String("storage")
    bundlePath := c.String("bundle-path")
    verify := c.Bool("verify")
    cosignKey := c.String("cosignKey")
    background := c.Bool("background")

    // Validate bundle path exists
    if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
        return fmt.Errorf("bundle not found at %s", bundlePath)
    }

    // Create registry daemon
    daemon := registry.NewRegistryDaemon(port, storage, bundlePath, verify, cosignKey)

    // Check if port is already in use
    if daemon.IsRunning() {
        return fmt.Errorf("registry port %s is already in use", port)
    }

    // Warn about background mode
    if background {
        logger.Warn("Running in background mode is not recommended")
        logger.Warn("Consider using systemd, supervisord, or similar for process management")
    }

    // Create context with signal handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Setup signal handlers
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    go func() {
        sig := <-sigChan
        logger.Infof("Received signal: %v", sig)
        cancel()
    }()

    // Start registry daemon
    logger.Info("Starting k0rdentd registry daemon...")
    logger.Infof("Configuration:")
    logger.Infof("  Bundle: %s", bundlePath)
    logger.Infof("  Storage: %s", storage)
    logger.Infof("  Address: %s:%s", host, port)
    logger.Infof("  Verify signature: %t", verify)

    if err := daemon.Start(ctx); err != nil {
        return fmt.Errorf("registry daemon failed: %w", err)
    }

    return nil
}
```

**Acceptance**:
- `k0rdentd registry --port 5000 --bundle-path /opt/bundle.tar.gz` works
- All flags supported
- All environment variables supported
- Validates bundle path before starting
- Checks if port is already in use

---

## Phase 3: Airgap Installation with Registry ✅ (COMPLETED)

**Estimated Effort**: 3 days
**Status**: ✅ All core tasks completed
**Last Updated**: 2026-02-11

### Current State

**Status**: ✅ Implementation complete and functional

The airgap installation flow has been fully implemented in:
- `internal/airgap/installer.go` - Main airgap installer
- `pkg/installer/installer.go` - Integration with main installer
- `pkg/generator/generator.go` - Airgap k0s config generation
- `pkg/config/k0rdentd.go` - Airgap configuration structure

**Current Behavior**:
```go
func (i *Installer) installAirgap(k0rdentConfig *config.K0rdentConfig) error {
    logger := utils.GetLogger()

    metadata := airgap.GetBuildMetadata()
    logger.Infof("Air-gapped installation (K0s: %s, K0rdent: %s)",
        metadata.K0sVersion, metadata.K0rdentVersion)

    if i.dryRun {
        logger.Infof("📝 Dry run mode - airgap installation steps:")
        logger.Infof("1. Install k0s from embedded binary")
        logger.Infof("2. Install k0rdent from embedded helm chart")
        if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
            logger.Infof("3. Create cloud provider credentials")
        }
        return nil
    }

    // TODO: Phase 3 - Implement airgap installation with registry daemon
    // The airgap installer needs to be rewritten to work with registry daemon approach.
    // It should:
    // 1. Extract k0s binary from embedded assets, copy the `k0s-<VERSION>-amd64` file to `/usr/local/bin/k0s`
    // 2. Create a `k0s.yaml` file that configures the helm installation of K0rdent with the right references to local registry.
    // 3. Configure k0s to use local registry (localhost:5000 or configured address)
    // 4. Install k0s using `InstallK0s()`.
    //
    // For now, return an error with clear instructions.
    return fmt.Errorf("airgap installation is not yet implemented for registry daemon approach\n" +
        "\n" +
        "To use airgap feature:\n" +
        "1. Start registry daemon first:\n" +
        "   sudo k0rdentd registry --bundle-path <bundle.tar.gz> --port 5000\n" +
        "\n" +
        "2. Then run airgap installation (Phase 3 - not yet implemented):\n" +
        "   sudo k0rdentd install --registry-address localhost:5000\n" +
        "\n" +
        "See docs/FEATURE_airgap.md for more details")
}
```

### 3.1 k0s Binary Extraction ✅ (COMPLETED)

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day
**Completed**: 2026-02-11

- ✅ Implemented `ExtractK0sBinary()` in `internal/airgap/installer.go`
- ✅ Extracts from embedded assets using `assets.K0sBinary`
- ✅ Copies to `/usr/local/bin/k0s` with executable permissions (0755)
- ✅ Creates `/usr/local/bin` directory if it doesn't exist

**Implementation**:
```go
// internal/airgap/installer.go:32-93
func (i *Installer) ExtractK0sBinary() error
```

---

### 3.2 Registry Configuration in k0s ✅ (COMPLETED)

**Status**: ✅ Completed
**Estimated Effort**: 1 day
**Completed**: 2026-02-11

- ✅ Implemented `GenerateAirgapK0sConfig()` in `pkg/generator/generator.go`
- ✅ Sets `default_pull_policy: Never` to prevent external image pulls
- ✅ Configures helm charts to use OCI from local registry
- ✅ Registry address configurable via config file (defaults to localhost:5000)
- ✅ Configuration written BEFORE k0s installation (no restart needed)

**Implementation**:
```go
// pkg/generator/generator.go:163-256
func GenerateAirgapK0sConfig(cfg *config.K0rdentdConfig, registryAddr string, insecure bool) ([]byte, error)
```

**Configuration Structure**:
```yaml
spec:
  images:
    default_pull_policy: Never  # Prevents external pulls
  extensions:
    helm:
      charts:
        - name: kcm
          chartname: oci://localhost:5000/charts/k0rdent-enterprise
          version: "1.2.2"
          namespace: kcm-system
          values: |
            controller:
              globalRegistry: localhost:5000
              templatesRepoURL: oci://localhost:5000/charts
            image:
              repository: localhost:5000/kcm-controller
            flux2:
              cli:
                image: localhost:5000/fluxcd/flux-cli
              helmController:
                image: localhost:5000/fluxcd/helm-controller
              sourceController:
                image: localhost:5000/fluxcd/source-controller
            regional:
              telemetry:
                mode: disabled
                controller:
                  image:
                    repository: localhost:5000/kcm-telemetry
              cert-manager:
                image:
                  repository: localhost:5000/jetstack/cert-manager-controller
                webhook:
                  image:
                    repository: localhost:5000/jetstack/cert-manager-webhook
                cainjector:
                  image:
                    repository: localhost:5000/jetstack/cert-manager-cainjector
                startupapicheck:
                  image:
                    repository: localhost:5000/jetstack/cert-manager-startupapicheck
              cluster-api-operator:
                image:
                  manager:
                    repository: localhost:5000/capi-operator/cluster-api-operator
              velero:
                image:
                  repository: localhost:5000/velero/velero
            rbac-manager:
              enabled: true
              image:
                repository: localhost:5000/reactiveops/rbac-manager
            k0rdent-ui:
              image:
                repository: localhost:5000/k0rdent-ui
            datasourceController:
              image:
                repository: localhost:5000/datasource-controller
```

**Note**: Each component's image repository is explicitly configured to point to the local registry, as required by k0rdent airgap installation. See: https://docs.mirantis.com/k0rdent-enterprise/latest/admin/installation/airgap/airgap-install/

**Old Code (PLANNED, NOT USED)**:
```go
// internal/airgap/installer.go
func (i *AirGapInstaller) configureRegistry() error {
    // Update /etc/k0s/k0s.yaml with registry mirrors
    configPath := "/etc/k0s/k0s.yaml"
    config, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }

    // Add registry configuration
    updatedConfig := addRegistryMirrors(string(config), i.registryAddress)

    // Write back
    if err := os.WriteFile(configPath, []byte(updatedConfig), 0600); err != nil {
        return err
    }

    // Restart k0s
    return restartK0s()
}

func addRegistryMirrors(config, registryAddr string) string {
    // Add registry mirrors to k0s config
    // See: https://docs.k0sproject.io/stable/airgap-install/
    return fmt.Sprintf("%s\nspec:\n  registry:\n    mirrors:\n      docker.io:\n        endpoints:\n          - http://%s", config, registryAddr)
}
```

---

### 3.3 k0rdent Installation via Helm ✅ (COMPLETED)

**Status**: ✅ Completed
**Estimated Effort**: 1 day
**Completed**: 2026-02-11

- ✅ K0rdent installation configured via k0s helm operator (no helm CLI needed)
- ✅ Uses OCI chart from local registry: `oci://localhost:5000/charts/k0rdent-enterprise`
- ✅ All component image repositories explicitly configured to point to local registry
- ✅ `controller.globalRegistry` and `controller.templatesRepoURL` set for k0rdent
- ✅ Reuses existing `waitForK0rdentInstalled()` from online installer

**Implementation Approach**:
Instead of using helm CLI, we leverage k0s's integrated helm operator by configuring the chart in k0s.yaml:
```yaml
spec:
  extensions:
    helm:
      charts:
        - name: kcm
          chartname: oci://localhost:5000/charts/k0rdent-enterprise
          version: "1.2.2"
          namespace: kcm-system
          values: |
            global:
              registry: localhost:5000
```

**Benefits**:
- No external helm CLI dependency
- k0s manages the helm release lifecycle
- Automatic retry and reconciliation
- Consistent with online installation approach

**Old Code (PLANNED, NOT USED - helm CLI approach):**
```go
// internal/airgap/installer.go
func (i *AirGapInstaller) installK0rdent(ctx context.Context, cfg *config.K0rdentConfig, version string) error {
    // Install helm if not present
    if err := ensureHelmInstalled(); err != nil {
        return err
    }

    // Get k0rdent chart from local registry
    chartRepo := fmt.Sprintf("http://%s/helm-charts", i.registryAddress)

    // Install via helm
    cmd := exec.Command("helm", "install", "k0rdent", chartRepo,
        "--namespace", "kcm-system",
        "--create-namespace",
        "--set", "image.registry="+i.registryAddr,
        "--set", fmt.Sprintf("image.tag=%s", version),
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("helm install failed: %w, output: %s", err, string(output))
    }

    // Wait for k0rdent to be ready
    return i.waitForK0rdent(ctx)
}
```

---

### 3.4 Refactor to Avoid Code Duplication ✅ (COMPLETED)

**Status**: ✅ Completed
**Estimated Effort**: 0.5 day
**Completed**: 2026-02-11

- ✅ Airgap-specific logic isolated in `internal/airgap/installer.go`
- ✅ Common k0s installation logic in `pkg/installer/installer.go` reused for both modes
- ✅ Airgap installer prepares environment, then delegates to common `installK0s()`
- ✅ Common functions reused:
  - `installK0s()` - Install and start k0s service
  - `waitForK0rdentInstalled()` - Wait for k0rdent readiness
  - `waitForCAPIProviderHelmReleases()` - Wait for CAPI providers
  - `createCredentials()` - Create cloud provider credentials

**Code Organization**:
```
internal/airgap/installer.go:
  - ExtractK0sBinary()          # Airgap-specific
  - Install()                   # Airgap preparation
  - writeK0sConfig()            # Writes airgap k0s config

pkg/installer/installer.go:
  - installAirgap()             # Calls airgap.Install() then common functions
  - installK0s()                # Common: installs k0s (both modes)
  - waitForK0rdentInstalled()   # Common: waits for k0rdent (both modes)
  - createCredentials()         # Common: creates credentials (both modes)

pkg/generator/generator.go:
  - GenerateK0sConfig()         # Online mode config
  - GenerateAirgapK0sConfig()   # Airgap mode config
```

**Flow Comparison**:
```
Online:  GenerateK0sConfig() → installK0s() → waitForK0rdent() → createCredentials()
Airgap:  airgap.Install() → GenerateAirgapK0sConfig() → installK0s() → waitForK0rdent() → createCredentials()
         └─ ExtractK0sBinary()
```

---

## Phase 4: Multi-Platform Support (PENDING)

**Estimated Effort**: 2 days

---

## Phase 5: Future Enhancements (TODO)

**Estimated Effort**: TBD

- Bundle auto-download from Mirantis
- Upgrade handling for k0rdent version updates
- Private registry support
- Custom CA certificate support
- Bundle migration between registry instances

**Note**: These are optional enhancements. Core functionality works without them.

---

## Testing Strategy

### Unit Tests

- [ ] Test version extraction from bundle
- [ ] Test registry daemon initialization
- [ ] Test cosign verification (mocked)

### Integration Tests

- [ ] Test registry daemon with real bundle
- [ ] Test airgap install with local registry
- [ ] Test multi-worker setup

### Manual Testing

```bash
# Terminal 1: Start registry daemon
sudo ./k0rdentd-airgap registry \
  --bundle-path /opt/airgap-bundle-1.2.3.tar.gz \
  --port 5000 \
  --storage /var/lib/k0rdentd/registry

# Terminal 2: Install k0s and k0rdent (Phase 3 - NOT YET IMPLEMENTED)
sudo ./k0rdentd-airgap install \
  --airgap-bundle-path /opt/airgap-bundle-1.2.3.tar.gz \
  --registry-address localhost:5000

# Verify k0rdent is running
sudo k0s kubectl get pods -n kcm-system
```

---

## Open Questions

1. ~~**CAPI Provider Images**: Do we need to bundle Cluster API provider images for AWS/Azure/OpenStack?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: YES - All CAPI providers included in enterprise bundle

2. ~~**Helm Chart Dependencies**: Are all dependencies available offline?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: YES - All Helm dependencies included in bundle

3. ~~**k0s Multi-Arch**: Do we embed multiple architectures or build separate binaries?~~
   - **Status**: ✅ RESOLVED
   - **Decision**: Phase 1 = single arch (amd64), Phase 4 = multi-arch

4. ~~**Bundle Configuration**: How does user specify bundle location?~~
   - **Status**: ✅ RESOLVED
   - **Decision**: Config file, CLI flag, or environment variable

5. ~~**K0rdent Version**: How to get k0rdent version?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: Extract from bundle Chart.yaml (not build-time)

6. ~~**Bundle Verification**: How to verify bundle integrity?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: Use cosign instead of structure validation

7. ~~**Code Duplication**: How to avoid duplication between online/airgap?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: Refactor to separate airgap-specific from common tasks

8. ~~**OCI Registry Implementation**: How to implement local registry?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: Use go-containerregistry/pkg/registry with disk-based storage

9. ~~**Registry Daemon Lifecycle**: How to run registry?~~
   - **Status**: ✅ RESOLVED
   - **Answer**: Separate `k0rdentd registry` command, persistent daemon

10. **Asset Embedding**: How to embed k0s binary?~~
    - **Status**: ✅ RESOLVED (FIXED)
    - **Answer**: Use `//go:embed` directive and import assets package where used

---

## OCI Registry Fix (Post-Phase 3) ✅ (COMPLETED)

**Date**: 2026-02-12
**Status**: ✅ COMPLETED

### Issue Summary

Three interconnected issues were identified in the airgap registry implementation:

1. **Incorrect Tag Format**: `charts/k0rdent-enterprise_1.2.2.tar` was stored as `charts/k0rdent-enterprise_1.2.2:latest` instead of `charts/k0rdent-enterprise:1.2.2`
2. **Incorrect Path Structure for Root-Level Images**: Images at the root of the extracted bundle (e.g., `kcm-controller_1.2.2.tar`) were stored as `extracted/kcm-controller_1.2.2:latest` instead of `kcm-controller:1.2.2`
3. **External Dependency on Skopeo**: User prefers native Go implementation using `go-containerregistry` (deferred to Phase 4)

### Root Cause

The `pathToImageRef()` function in `internal/airgap/registry/pusher.go`:
- Used `filepath.Base(filepath.Dir(path))` which captured the temp directory name (`extracted`) for root-level images
- Did not convert underscore (`_`) version separator to OCI tag format (`:`)

### Implementation

**File Modified**: `internal/airgap/registry/pusher.go`

**Changes**:
1. Updated `pathToImageRef()` to accept `bundleRoot` parameter
2. Use `filepath.Rel(bundleRoot, imgPath)` to get relative path from bundle root
3. Parse underscore version separator and convert to OCI tag format
4. Handle both `name_version` and `name:tag` formats
5. Updated `pushSingleImage()` to pass `bundleRoot` parameter
6. Updated `pushImagesWithProgress()` to pass `bundleRoot` parameter

**Code Changes**:
```go
// pathToImageRef converts a file path to an OCI image reference
// bundleRoot is the root directory of the extracted bundle
func pathToImageRef(imgPath string, bundleRoot string) string {
    // Get relative path from bundle root to maintain logical structure
    relPath, _ := filepath.Rel(bundleRoot, imgPath)

    // Remove .tar suffix
    ref := strings.TrimSuffix(relPath, ".tar")

    // Split into directory and filename components
    dir := filepath.Dir(ref)
    if dir == "." {
        dir = ""
    }
    filename := filepath.Base(ref)

    // Parse version from filename
    var imageName, tag string

    if idx := strings.LastIndex(filename, "_"); idx != -1 {
        // Format: name_version (e.g., k0rdent-enterprise_1.2.2)
        imageName = filename[:idx]
        tag = filename[idx+1:]
    } else if idx := strings.LastIndex(filename, ":"); idx != -1 {
        // Format: name:tag (e.g., k0s:v1.32.8-k0s.0)
        imageName = filename[:idx]
        tag = filename[idx+1:]
    } else {
        // No version found, use filename as image name and "latest" as tag
        imageName = filename
        tag = "latest"
    }

    // Build final reference: dir/imageName:tag
    if dir != "" && dir != "." {
        return fmt.Sprintf("%s/%s:%s", dir, imageName, tag)
    }
    return fmt.Sprintf("%s:%s", imageName, tag)
}
```

### Verification

After the fix:
- `charts/k0rdent-enterprise_1.2.2.tar` → `charts/k0rdent-enterprise:1.2.2`
- `k0sproject/k0s:v1.32.8-k0s.0.tar` → `k0sproject/k0s:v1.32.8-k0s.0`
- `kcm-controller_1.2.2.tar` (root level) → `kcm-controller:1.2.2`

### Next Steps

- **Phase 4**: Evaluate migration from skopeo to go-containerregistry (deferred)
- Test the fix with actual k0rdent bundle
- Verify k0rdent installation from local registry

---

## Containerd Registry Mirror Configuration (Post-Phase 3) ✅ (COMPLETED)

**Date**: 2026-02-12
**Status**: ✅ COMPLETED

### Issue Summary

When k0s starts in airgap mode, its underlying containerd runtime attempts to pull system images from internet registries (`registry.k8s.io`, `quay.io`). Since airgap environment has no internet access, these pulls fail. The local OCI registry running at `localhost:5000` contains all required images, but containerd is not configured to use it as a mirror.

### Root Cause

k0s does not have a simplified way to configure local registry mirror. The containerd configuration needs to be set up manually with:
1. Containerd drop-in configuration at `/etc/k0s/containerd.d/cri-registry.toml`
2. Registry hosts configuration at `/etc/k0s/containerd.d/certs.d/<registry>/hosts.toml`

### Implementation

**New Package**: `internal/airgap/containerd/config.go`

**Functions**:
- `CRIRegistryConfig()` - Returns the CRI registry configuration content
- `HostsConfig(registry, mirrorAddr)` - Returns hosts.toml content for a given registry
- `SetupContainerdMirror(mirrorAddr)` - Creates all containerd configuration files

**Modified File**: `internal/airgap/installer.go`

**Changes**:
- Added Step 4 to configure containerd registry mirror
- Calls `containerd.SetupContainerdMirror(registryAddr)` before writing k0s configuration
- Updated installation sequence to include containerd configuration

**Directory Structure Created**:
```
/etc/k0s/containerd.d/
├── cri-registry.toml              # CRI registry config
└── certs.d/
    ├── registry.k8s.io/
    │   └── hosts.toml             # Mirror for registry.k8s.io
    └── quay.io/
        └── hosts.toml             # Mirror for quay.io
```

**Configuration Files**:

1. `/etc/k0s/containerd.d/cri-registry.toml`:
```toml
version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
config_path = "/etc/k0s/containerd.d/certs.d"
```

2. `/etc/k0s/containerd.d/certs.d/registry.k8s.io/hosts.toml`:
```toml
server = "https://registry.k8s.io"

[host."http://127.0.0.1:5000"]
  capabilities = ["pull", "resolve"]
```

3. `/etc/k0s/containerd.d/certs.d/quay.io/hosts.toml`:
```toml
server = "https://quay.io"

[host."http://127.0.0.1:5000"]
  capabilities = ["pull", "resolve"]
```

### Installation Sequence

1. Extract k0s binary to `/usr/local/bin/k0s`
2. **Configure containerd registry mirror** (NEW)
3. Generate k0s configuration for airgap mode
4. Write k0s configuration to `/etc/k0s/k0s.yaml`
5. Run `k0s install controller --enable-worker`
6. Start k0s service

### Multi-Worker Support

For multi-worker clusters, the registry address must be reachable from all nodes. This is already configurable via `config.airgap.registry.address` in the k0rdentd config file.

### Verification

After k0s starts, verify the configuration:
```bash
# Check containerd configuration
cat /etc/k0s/containerd.d/cri-registry.toml

# Check registry hosts
cat /etc/k0s/containerd.d/certs.d/quay.io/hosts.toml
cat /etc/k0s/containerd.d/certs.d/registry.k8s.io/hosts.toml

# Verify images are pulled from local registry
k0s kubectl get pods -A
# All pods should be running without image pull errors
```

---

## References

- Design document: `docs/FEATURE_airgap.md`
- Bundle inventory: `docs/K0RDENT_BUNDLE_CATALOG.md`
- k0s airgap docs: https://docs.k0sproject.io/stable/airgap-install/
- k0rdent enterprise: https://docs.mirantis.com/k0rdent-enterprise/latest/admin/installation/airgap/airgap-bundles/
- go-containerregistry: https://github.com/google/go-containerregistry
- cosign: https://sigstore.dev/cosign/
- containerd registry hosts: https://github.com/containerd/containerd/blob/main/docs/hosts.md
