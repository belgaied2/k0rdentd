# CLI Reference

Complete reference for all K0rdentd CLI commands.

## Global Flags

These flags apply to all commands:

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--config-file, -c` | `K0RDENTD_CONFIG_FILE` | `/etc/k0rdentd/k0rdentd.yaml` | Path to configuration file |
| `--debug` | `K0RDENTD_DEBUG` | `false` | Enable debug logging |
| `--dry-run` | - | `false` | Show what would be done without making changes |
| `--help, -h` | - | - | Show help for command |

## install

Install K0s and K0rdent on the VM.

### Usage

```bash
k0rdentd install [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config-file, -c` | `/etc/k0rdentd/k0rdentd.yaml` | Path to configuration file |
| `--airgap` | `false` | Force airgap mode installation |
| `--debug` | `false` | Enable debug logging |
| `--dry-run` | `false` | Show what would be done |

### Examples

```bash
# Basic installation
sudo k0rdentd install

# With custom config file
sudo k0rdentd install -c /path/to/config.yaml

# Airgap installation
sudo k0rdentd install --airgap

# Debug mode
sudo k0rdentd install --debug

# Dry run
sudo k0rdentd install --dry-run
```

### What It Does

1. Checks if K0s binary exists, installs if missing
2. Generates K0s configuration from k0rdentd.yaml
3. Installs K0s controller with worker enabled
4. Starts K0s service
5. Waits for K0s to be ready
6. Waits for K0rdent Helm chart to be installed
7. Creates cloud provider credentials (if configured)
8. Exposes K0rdent UI

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Installation error |

---

## uninstall

Uninstall K0s and K0rdent from the VM.

### Usage

```bash
k0rdentd uninstall [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config-file, -c` | `/etc/k0rdentd/k0rdentd.yaml` | Path to configuration file |
| `--force` | `false` | Force uninstall without confirmation |
| `--debug` | `false` | Enable debug logging |

### Examples

```bash
# Basic uninstall
sudo k0rdentd uninstall

# Force uninstall
sudo k0rdentd uninstall --force

# With custom config
sudo k0rdentd uninstall -c /path/to/config.yaml
```

### What It Does

1. Stops K0s service
2. Uninstalls K0s
3. Removes configuration files
4. Cleans up data directories (optional)

---

## version

Show version information.

### Usage

```bash
k0rdentd version [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | `false` | Output in JSON format |
| `--short` | `false` | Show only version number |

### Examples

```bash
# Show version
k0rdentd version

# Short format
k0rdentd version --short

# JSON format
k0rdentd version --json
```

### Output

```
k0rdentd version v0.2.0
A CLI tool to deploy K0s and K0rdent
Copyright © 2024 belgaied2
```

---

## expose-ui

Expose the K0rdent UI via ingress and display access URLs.

### Usage

```bash
k0rdentd expose-ui [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config-file, -c` | `/etc/k0rdentd/k0rdentd.yaml` | Path to configuration file |
| `--timeout` | `5m` | Timeout for waiting for UI |
| `--path` | `/k0rdent-ui` | Ingress path |
| `--debug` | `false` | Enable debug logging |

### Examples

```bash
# Expose UI with defaults
sudo k0rdentd expose-ui

# Custom timeout
sudo k0rdentd expose-ui --timeout 10m

# Custom path
sudo k0rdentd expose-ui --path /ui
```

### What It Does

1. Waits for K0rdent UI deployment to be ready
2. Checks K0rdent UI service
3. Creates ingress for UI
4. Detects VM IP addresses
5. Filters out Calico and Docker interfaces
6. Tests UI accessibility
7. Displays access URLs

---

## registry

Start the OCI registry daemon for airgap installations.

### Usage

```bash
k0rdentd registry [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--bundle-path, -b` | - | Path to k0rdent airgap bundle (required) |
| `--port, -p` | `5000` | Registry port |
| `--host` | `0.0.0.0` | Host to bind to |
| `--storage, -s` | `/var/lib/k0rdentd/registry` | Storage directory |
| `--verify` | `false` | Verify bundle signature |
| `--cosignKey` | - | Cosign public key URL or path |
| `--daemon, -d` | `false` | Run as background daemon |
| `--debug` | `false` | Enable debug logging |

### Examples

```bash
# Start registry with bundle
sudo k0rdentd registry -b /path/to/bundle.tar.gz

# Custom port
sudo k0rdentd registry -b /path/to/bundle.tar.gz -p 5001

# With signature verification
sudo k0rdentd registry -b /path/to/bundle.tar.gz \
  --verify true \
  --cosignKey https://get.mirantis.com/cosign.pub

# Run as daemon
sudo k0rdentd registry -b /path/to/bundle.tar.gz -d
```

### What It Does

1. Verifies bundle signature (if enabled)
2. Extracts k0rdent version from bundle
3. Initializes local OCI registry
4. Pushes images from bundle to registry
5. Starts HTTP server
6. Handles graceful shutdown

---

## export-worker-artifacts

Export artifacts needed for worker nodes in airgap installations.

### Usage

```bash
k0rdentd export-worker-artifacts [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--output, -o` | `/tmp/worker-bundle` | Output directory |
| `--bundle-path, -b` | - | Path to k0rdent airgap bundle |
| `--debug` | `false` | Enable debug logging |

### Examples

```bash
# Export to default location
k0rdentd export-worker-artifacts

# Custom output location
k0rdentd export-worker-artifacts -o /path/to/output

# With bundle reference
k0rdentd export-worker-artifacts -b /path/to/bundle.tar.gz
```

### What It Does

1. Extracts embedded K0s binary
2. Creates bundle reference file
3. Generates helper scripts
4. Creates README with instructions

---

## show-flavor

Show the build flavor (online or airgap).

### Usage

```bash
k0rdentd show-flavor [flags]
```

### Examples

```bash
k0rdentd show-flavor
```

### Output

```
airgap
```

or

```
online
```

---

## config

Manage configuration (placeholder for future use).

### Usage

```bash
k0rdentd config [command]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `view` | View current configuration |
| `validate` | Validate configuration file |

---

## Environment Variables

All CLI flags can be set via environment variables:

| Variable | Equivalent Flag |
|----------|----------------|
| `K0RDENTD_CONFIG_FILE` | `--config-file` |
| `K0RDENTD_DEBUG` | `--debug` |
| `K0RDENTD_LOG_LEVEL` | (sets log level) |
| `K0RDENTD_K0S_VERSION` | (sets k0s.version in config) |
| `K0RDENTD_K0RDENT_VERSION` | (sets k0rdent.version in config) |
| `K0RDENTD_AIRGAP_BUNDLE_PATH` | (sets airgap.bundlePath in config) |
| `K0RDENTD_REGISTRY_ADDRESS` | (sets airgap.registry.address in config) |
| `K0RDENTD_REGISTRY_PORT` | `--port` (for registry command) |
| `K0RDENTD_REGISTRY_STORAGE` | `--storage` (for registry command) |

---

## Logging

K0rdentd uses logrus for logging with the following levels:

| Level | Usage |
|-------|-------|
| DEBUG | Intermediate steps during processing |
| INFO | Finished steps relevant to user |
| WARN | Unexpected behavior, ignored errors |
| ERROR | Critical errors |

### Enabling Debug Logging

```bash
# Via flag
sudo k0rdentd install --debug

# Via environment variable
export K0RDENTD_DEBUG=true
sudo k0rdentd install

# Via config file
debug: true
```

---

## Exit Codes

Common exit codes across all commands:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Installation/uninstallation error |
| 4 | Runtime error |
| 130 | Interrupted (Ctrl+C) |
