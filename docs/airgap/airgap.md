# Airgap Support

K0rdentd provides full support for airgap (offline) installations, allowing you to deploy K0s and K0rdent in environments without internet access.

## Overview

### Build Flavors

| Flavor | Description | Binary Size | Use Case |
|--------|-------------|-------------|----------|
| Online | Downloads K0s and K0rdent from internet | ~60 MB | Standard installations |
| Airgap | Embeds K0s binary, uses external bundle | ~300 MB | Offline installations |

### What's Embedded

The airgap binary includes:

- **K0s binary** (~100 MB) - Embedded directly
- **Skopeo binary** (~40 MB) - For image pushing

### What's External

The K0rdent Enterprise Airgap Bundle (downloaded separately):

- **Size**: ~22 GB (compressed)
- **Contents**: 233 OCI images, 83 Helm charts
- **Source**: Mirantis download portal

## Registry Daemon

The registry daemon is a key component for airgap installations.

### Starting the Daemon

```bash
sudo k0rdentd registry \
  --bundle-path /opt/k0rdent/airgap-bundle-1.2.2.tar.gz \
  --port 5000 \
  --storage /var/lib/k0rdentd/registry
```

### Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--bundle-path, -b` | (required) | Path to airgap bundle |
| `--port, -p` | 5000 | Registry port |
| `--host` | 0.0.0.0 | Host to bind to |
| `--storage, -s` | /var/lib/k0rdentd/registry | Storage directory |
| `--verify` | false | Verify bundle signature |
| `--cosignKey` | - | Cosign public key URL/path |

## Containerd Mirror Configuration

For airgap mode, containerd is configured to use the local registry as a mirror.

### Directory Structure

```
/etc/k0s/
├── k0s.yaml                           # k0s cluster configuration
└── containerd.d/                      # Drop-in configuration directory
    ├── cri-registry.toml              # Containerd CRI registry config
    └── certs.d/                       # Registry hosts configuration
        ├── registry.k8s.io/
        │   └── hosts.toml             # Mirror config for registry.k8s.io
        └── quay.io/
            └── hosts.toml             # Mirror config for quay.io
```

## Multi-Worker Support

Workers must be configured to use the bootstrap node's registry:

```bash
BOOTSTRAP_IP="192.168.1.10"

sudo mkdir -p /etc/k0s/containerd.d/certs.d/registry.k8s.io
cat <<EOF | sudo tee /etc/k0s/containerd.d/certs.d/registry.k8s.io/hosts.toml
server = "https://registry.k8s.io"
[host."http://${BOOTSTRAP_IP}:5000"]
  capabilities = ["pull", "resolve"]
EOF
```

## Building Airgap Binary

```bash
# Build airgap binary (downloads K0s and Skopeo)
make build-airgap

# Binary location
ls bin/k0rdentd-airgap
```

## See Also

- [Airgap Installation Guide](../user-guide/airgap-installation.md) - Step-by-step installation
- [Bundle Catalog](bundle-catalog.md) - Complete bundle contents
