# Installation

This guide covers how to install K0rdentd on your system.

## Prerequisites

Before installing K0rdentd, ensure your system meets the following requirements:

### System Requirements

- **Operating System**: Linux (AMD64 or ARM64)
- **Disk Space**: 
  - Online mode: ~10 GB free
  - Airgap mode: ~100 GB free (for bundle and registry)
- **Memory**: Minimum 4 GB RAM (8 GB recommended)
- **Privileges**: `sudo` access required

### Software Dependencies

- `curl` or `wget` (for downloading)
- `tar` (for extracting archives)
- For airgap installations:
  - `skopeo` (included in airgap binary)
  - `cosign` (optional, for bundle verification)

## Installation Methods

### Download from GitHub Releases

The recommended way to install K0rdentd is to download the latest release from GitHub.

#### Online Flavor (Default) - one-liner installation

```bash
# The following commands installs the latest version of K0rdentd
curl -sfL https://k0rdentd.belgai.de | sudo bash
```

#### Airgap Flavor

For offline installations, download the airgap build:

```bash
# Set the version
K0RDENTD_VERSION="v0.2.0"

# Download for AMD64
curl -L "https://github.com/belgaied2/k0rdentd/releases/download/${K0RDENTD_VERSION}/k0rdentd-airgap-${K0RDENTD_VERSION}-linux-amd64.tar.gz" -o k0rdentd-airgap.tar.gz

# Extract
tar -xzf k0rdentd-airgap.tar.gz

# Install
sudo mv k0rdentd-airgap /usr/local/bin/k0rdentd
sudo chmod +x /usr/local/bin/k0rdentd
```

**NOTE:** Please be aware that the airgap version needs the official K0rdent Enterprise bundle for the actual installation of K0s and K0rdent.

### Build from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/belgaied2/k0rdentd.git
cd k0rdentd

# Build online flavor
make build

# Or build airgap flavor
make build-airgap

# Install
sudo cp bin/k0rdentd /usr/local/bin/
```

## Verify Installation

After installation, verify that K0rdentd is working:

```bash
# Check version
k0rdentd version

# Check build flavor
k0rdentd show-flavor
```

Expected output:

```
k0rdentd version v0.2.0
A CLI tool to deploy K0s and K0rdent
Copyright © 2024 belgaied2
```

## Next Steps

- [Quick Start Guide](quick-start.md) - Get K0rdent running quickly
- [Configuration](configuration.md) - Learn about configuration options
- [Airgap Installation](../user-guide/airgap-installation.md) - For offline environments

## Troubleshooting

### Permission Denied

If you get permission errors:

```bash
# Ensure binary is executable
sudo chmod +x /usr/local/bin/k0rdentd

# Ensure config directory has correct permissions
sudo chown -R root:root /etc/k0rdentd
sudo chmod 755 /etc/k0rdentd
```

### Binary Not Found

If the `k0rdentd` command is not found:

```bash
# Check if it's in your PATH
which k0rdentd

# If not, add the directory to PATH
export PATH=$PATH:/usr/local/bin

# Or move to a directory in PATH
sudo mv k0rdentd /usr/local/bin/
```

### K0s Installation Issues

K0rdentd will automatically install K0s if it's not present. If you encounter issues:

```bash
# Check if k0s is installed
which k0s

# Install k0s manually
curl -sSLf https://get.k0s.sh | sudo sh
```
