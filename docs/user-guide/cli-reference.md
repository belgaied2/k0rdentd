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
| `--k0s-version, -k` | - | Override K0s version from config |
| `--k0rdent-version, -r` | - | Override K0rdent version from config |
| `--join` | `false` | Join an existing cluster (requires --mode) |
| `--mode` | - | Node mode: controller or worker (required if --join is set) |
| `--replace-k0s, -R` | `false` | Replace existing k0s binary without prompting (only if not running) |
| `--debug` | `false` | Enable debug logging |
| `--dry-run` | `false` | Show what would be done |

### Examples

```bash
# Basic installation (first controller, implicit cluster-init)
sudo k0rdentd install

# With custom config file
sudo k0rdentd install -c /path/to/config.yaml

# Install specific k0s version
sudo k0rdentd install --k0s-version v1.32.4+k0s.0

# Replace existing k0s binary (if not running)
sudo k0rdentd install --replace-k0s

# Or via environment variable
K0RDENTD_REPLACE_K0S=true sudo k0rdentd install

# Join as additional controller
sudo k0rdentd install --join --mode=controller

# Join as worker node
sudo k0rdentd install --join --mode=worker

# Join using config file (join section in k0rdentd.yaml)
sudo k0rdentd install

# Debug mode
sudo k0rdentd install --debug

# Dry run
sudo k0rdentd install --dry-run
```

### What It Does

**First Controller (cluster-init):**
1. Checks if K0s binary exists, installs if missing
2. Checks for k0s version conflicts (online mode only)
3. Generates K0s configuration from k0rdentd.yaml
4. Installs K0s controller with worker enabled
5. Starts K0s service
6. Waits for K0s to be ready
7. Waits for K0rdent Helm chart to be installed
8. Creates cloud provider credentials (if configured)
9. Exposes K0rdent UI

**Joining Node (controller or worker):**
1. Reads join configuration from k0rdentd.yaml or CLI flags
2. Configures containerd mirrors for airgap (if applicable)
3. Writes K0s join configuration
4. Installs K0s with join token
5. Starts K0s service
6. Waits for node to be ready

### K0s Version Management

K0rdentd provides comprehensive k0s version management that handles conflicts between:

- **Bundled k0s version** (airgap mode) - Always used, warns if config differs
- **Config file `k0s.version`** - Specifies desired version
- **Already installed k0s binary** - Checked for conflicts

#### Decision Matrix

| Mode | k0s Exists? | Config Version? | Running? | Action |
|------|-------------|-----------------|----------|--------|
| Airgap | N/A | Any | N/A | Use bundled version, warn if config differs |
| Online | No | Specified | N/A | Download specified version |
| Online | No | Not specified | N/A | Use latest stable (via get.k0s.sh) |
| Online | Yes | Same as installed | Any | Proceed with existing |
| Online | Yes | Different | No | Use `--replace-k0s` to replace |
| Online | Yes | Different | Yes | **Fail - require manual intervention** |

#### Airgap Mode Behavior

In airgap mode, the bundled k0s binary is always used. If `k0s.version` in the config differs from the bundled version, a warning is logged but installation continues:

```
⚠️  Config specifies k0s version v1.31.0+k0s.0, but bundled version is v1.32.4+k0s.0.
   Using bundled version for airgap installation.
```

#### Online Mode Behavior

**No existing k0s:**
- If `k0s.version` is specified in config: Download that specific version
- If `k0s.version` is NOT specified: Download latest stable version

**Existing k0s - Same version:**
- Proceed with installation using existing binary

**Existing k0s - Different version (not running):**
- Use `--replace-k0s` flag to replace the binary
- Or fail with conflict message

**Existing k0s - Different version (running):**
- **Always fail** with manual intervention instructions:
  ```
  ❌ Cannot replace k0s while it's running!
     The installed k0s (v1.30.0+k0s.0) is currently running as a service.
     Config specifies: v1.32.4+k0s.0

     To proceed, you must manually stop and reset k0s:
       sudo k0s stop

     Then run k0rdentd install again.
  ```

#### Version Format

k0s versions follow the format: `v{KUBERNETES_VERSION}+k0s.{K0S_PATCH}`

Examples:
- `v1.32.4+k0s.0`
- `v1.31.0+k0s.0`
- `v1.30.0+k0s.0`

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

## export-join-config

Export join configurations for additional nodes in multi-node deployments.

### Usage

```bash
k0rdentd export-join-config [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--output, -o` | `./join-configs` | Output directory for join config files |
| `--controller-ip` | (auto-detected) | Override auto-detected controller IP address |
| `--expiry, -e` | `24h` | Token expiry time (e.g., 24h, 168h for 7 days) |
| `--registry-port` | `5000` | Registry port for airgap mode |
| `--overwrite, -f` | `false` | Overwrite existing files |
| `--debug` | `false` | Enable debug logging |

### Examples

```bash
# Export join configs with defaults
sudo k0rdentd export-join-config

# Custom output directory
sudo k0rdentd export-join-config -o /path/to/configs

# Override controller IP (useful if auto-detection fails)
sudo k0rdentd export-join-config --controller-ip 192.168.1.100

# Longer token expiry (7 days)
sudo k0rdentd export-join-config --expiry 168h

# Overwrite existing files
sudo k0rdentd export-join-config --overwrite
```

### What It Does

1. Loads current k0rdentd configuration
2. Auto-detects controller IP address
3. Creates controller join token via `k0s token create --role=controller`
4. Creates worker join token via `k0s token create --role=worker`
5. Generates `controller-join.yaml` with join config for additional controllers
6. Generates `worker-join.yaml` with join config for worker nodes
7. Includes airgap registry settings if applicable

### Output Files

The command creates two files in the output directory:

**controller-join.yaml** - For joining additional controller nodes:
```yaml
join:
  mode: controller
  server: 192.168.1.10
  token: "k0s-controller-token..."

k0s:
  version: "v1.32.4+k0s.0"

# Airgap settings (if applicable)
airgap:
  registry:
    address: 192.168.1.10:5000
    insecure: true
```

**worker-join.yaml** - For joining worker nodes:
```yaml
join:
  mode: worker
  server: 192.168.1.10
  token: "k0s-worker-token..."

k0s:
  version: "v1.32.4+k0s.0"
```

### Usage After Export

```bash
# On additional controller node
scp controller-join.yaml user@controller2:/etc/k0rdentd/k0rdentd.yaml
sudo k0rdentd install

# On worker node
scp worker-join.yaml user@worker1:/etc/k0rdentd/k0rdentd.yaml
sudo k0rdentd install
```

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
| `K0RDENTD_REPLACE_K0S` | `--replace-k0s` |
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
