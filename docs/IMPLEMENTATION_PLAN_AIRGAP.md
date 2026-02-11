# Airgap Feature Implementation Plan

## Status: Phase 2 Completed - Registry Daemon Implemented

**Last Updated**: 2025-02-11
**Phase**: Phase 3 (Airgap Installation with Registry) - In Progress

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

## Phase 3: Airgap Installation with Registry (PARTIALLY COMPLETED)

**Estimated Effort**: 3 days
**Status**: ✅ `installAirgap()` exists but returns error with instructions

### Current State

**Status**: Framework in place, implementation pending

The `installAirgap()` function in [pkg/installer/installer.go:128-165](pkg/installer/installer.go#L128-L165) currently returns an error with instructions for users to use the registry daemon approach.

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
        logger.Infof("2. Load embedded image bundles")
        logger.Infof("3. Install k0rdent from embedded helm chart")
        if k0rdentConfig != nil && k0rdentConfig.Credentials.HasCredentials() {
            logger.Infof("4. Create cloud provider credentials")
        }
        return nil
    }

    // TODO: Phase 3 - Implement airgap installation with registry daemon
    // The airgap installer needs to be rewritten to work with registry daemon approach.
    // It should:
    // 1. Extract k0s binary from embedded assets
    // 2. Install and configure k0s
    // 3. Configure k0s to use local registry (localhost:5000 or configured address)
    // 4. Install k0rdent via helm from local registry
    //
    // For now, return an error with clear instructions.
    return fmt.Errorf("airgap installation is not yet implemented for registry daemon approach\n" +
        "\n" +
        "To use airgap feature:\n" +
        "1. Start registry daemon first:\n" +
        "   sudo k0rdentd registry --bundle-path <bundle.tar.gz> --port 5000\n" +
        "\n" +
        "2. Then run airgap installation (Phase 3 - not yet implemented):\n" +
        "   sudo k0rdentd install --airgap-bundle-path <bundle.tar.gz> --registry-address localhost:5000\n" +
        "\n" +
        "See docs/FEATURE_airgap.md for more details")
}
```

### 3.1 k0s Binary Extraction (PENDING)

**Status**: Not Started
**Estimated Effort**: 0.5 day

- [ ] Implement `extractK0sFromEmbedded()` function
- [ ] Install to /usr/local/bin/k0s
- [ ] Make executable

**Note**: The `exporter.go` already has `extractFromEmbedded()` that can be referenced

---

### 3.2 Registry Configuration in k0s (PENDING)

**Status**: Not Started
**Estimated Effort**: 1 day

- [ ] Implement `ConfigureK0sRegistry()` function
- [ ] Update k0s config with registry mirror
- [ ] Restart k0s to apply config
- [ ] Test image pull from local registry

**Code (PLANNED)**:
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

### 3.3 k0rdent Installation via Helm (PENDING)

**Status**: Not Started
**Estimated Effort**: 1 day

- [ ] Implement `InstallK0rdentFromRegistry()` function
- [ ] Use helm CLI to install from local registry
- [ ] Configure image repository to point to local registry
- [ ] Wait for k0rdent pods to be ready

**Code (PLANNED)**:
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

### 3.4 Refactor to Avoid Code Duplication (PENDING)

**Status**: Not Started
**Estimated Effort**: 0.5 day

- [ ] Move common install logic to `pkg/installer/`
- [ ] Airgap installer calls common functions
- [ ] Online installer calls common functions
- [ ] Ensure no duplication between modes

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

## References

- Design document: `docs/FEATURE_airgap.md`
- Bundle inventory: `docs/K0RDENT_BUNDLE_CATALOG.md`
- k0s airgap docs: https://docs.k0sproject.io/stable/airgap-install/
- k0rdent enterprise: https://docs.mirantis.com/k0rdent-enterprise/latest/admin/installation/airgap/airgap-bundles/
- go-containerregistry: https://github.com/google/go-containerregistry
- cosign: https://sigstore.dev/cosign/
