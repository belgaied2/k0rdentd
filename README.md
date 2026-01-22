# K0rdentd

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

K0rdentd is a CLI tool that automates the deployment of K0s and K0rdent on a VM. It follows a similar pattern to RancherD but simplifies the architecture by directly configuring K0s and using its built-in helm extension mechanism to install K0rdent.

## Overview

K0rdentd provides a streamlined approach to deploying Kubernetes clusters with K0s and managing them with K0rdent. It handles the entire deployment process from configuration to cluster initialization, making it easy to set up production-ready Kubernetes environments.

## Architecture

### High-Level Architecture

```mermaid
graph TD
    A[User] -->|CLI Commands| B[k0rdentd]
    B -->|Read Config| C[/etc/k0rdentd/k0rdentd.yaml]
    B -->|Generate Config| D[/etc/k0s/k0s.yaml]
    B -->|Execute| E[k0s binary]
    E -->|Deploy| F[K0s Cluster]
    E -->|Helm Extension| G[K0rdent]
```

### Key Components

1. **CLI Interface** ([`cmd/k0rdentd/main.go`](cmd/k0rdentd/main.go))
   - Built with [urfave/cli](https://github.com/urfave/cli) for a user-friendly command-line interface
   - Supports install, uninstall, version, and configuration commands

2. **Configuration Management** ([`pkg/config/k0rdentd.go`](pkg/config/k0rdentd.go))
   - YAML-based configuration file at `/etc/k0rdentd/k0rdentd.yaml`
   - Configures both K0s and K0rdent settings
   - Supports environment variables for configuration override

3. **K0s Configuration Generator** ([`pkg/generator/generator.go`](pkg/generator/generator.go))
   - Transforms k0rdentd configuration into K0s-compatible format
   - Generates `/etc/k0s/k0s.yaml` for cluster initialization
   - Includes helm extensions for K0rdent installation

4. **Installation Logic** ([`pkg/installer/installer.go`](pkg/installer/installer.go))
   - Executes k0s binary with generated configuration
   - Handles cluster initialization and K0rdent deployment
   - Provides progress tracking and error handling

5. **Logging Utilities** ([`pkg/utils/logging.go`](pkg/utils/logging.go))
   - Uses [logrus](https://github.com/sirupsen/logrus) for structured logging
   - Configurable log levels (DEBUG, INFO, WARN)
   - Automatic log level adjustment based on flags

## Installation

### Prerequisites

- Go 1.21 or higher
- k0s binary (will be installed by k0rdentd)
- Helm 3.x (required for K0rdent installation)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/belgaied2/k0rdentd.git
cd k0rdentd

# Build the binary
go build -o k0rdentd ./cmd/k0rdentd

# Make it executable (Linux/Mac)
chmod +x k0rdentd
```

### Installing k0rdentd

After building, you can install k0rdentd to your system:

```bash
# Copy to /usr/local/bin
sudo cp k0rdentd /usr/local/bin/

# Verify installation
k0rdentd version
```

## Usage

### Basic Usage

#### 1. Install K0s and K0rdent

```bash
# Using default configuration file
k0rdentd install

# Using custom configuration file
k0rdentd install --config-file /path/to/config.yaml

# Dry-run mode (shows what would be done)
k0rdentd install --dry-run

# With debug logging
k0rdentd install --debug
```

#### 2. Uninstall K0s and K0rdent

```bash
k0rdentd uninstall
```

#### 3. View Version Information

```bash
k0rdentd version
```

#### 4. Configure k0rdentd

```bash
k0rdentd config
```

### Configuration

#### Configuration File Structure

The main configuration file is located at `/etc/k0rdentd/k0rdentd.yaml` by default. You can specify a custom path using the `--config-file` flag or `K0RDENTD_CONFIG_FILE` environment variable.

```yaml
# K0s Configuration
k0s:
  version: "v1.34.3+k0s.0"
  api:
    address: 172.31.3.124
    port: 6443
  network:
    provider: calico
    podCIDR: 10.244.0.0/16
    serviceCIDR: 10.96.0.0/12
  storage:
    type: etcd
    etcd:
      peerAddress: 127.0.0.1

# K0rdent Configuration
k0rdent:
  version: "1.6.0"
  helm:
    chart: oci://ghcr.io/k0rdent/kcm/charts/kcm
    namespace: kcm-system
    values:
      replicas: 1

# Global Settings
debug: false
logLevel: "info"
```

#### Configuration Options

**K0s Configuration:**
- `version`: K0s version to install
- `api.address`: API server address
- `api.port`: API server port (default: 6443)
- `network.provider`: Network provider (e.g., calico, flannel)
- `network.podCIDR`: Pod network CIDR
- `network.serviceCIDR`: Service network CIDR
- `storage.type`: Storage type (e.g., etcd)
- `storage.etcd.peerAddress`: etcd peer address

**K0rdent Configuration:**
- `version`: K0rdent version to install
- `helm.chart`: Helm chart location (OCI or URL)
- `helm.namespace`: Namespace for K0rdent
- `helm.values`: Helm values for customization

**Global Settings:**
- `debug`: Enable debug logging
- `logLevel`: Log level (debug, info, warn)

#### Environment Variables

You can override configuration using environment variables:

```bash
export K0RDENTD_CONFIG_FILE=/custom/path/config.yaml
export K0RDENTD_DEBUG=true
export K0RDENTD_DRY_RUN=true
```

### Command-Line Flags

| Flag | Alias | Description | Environment Variable |
|------|-------|-------------|---------------------|
| `--config-file` | `-c` | Path to configuration file | `K0RDENTD_CONFIG_FILE` |
| `--debug` | `-d` | Enable debug logging | `K0RDENTD_DEBUG` |
| `--dry-run` | `-n` | Show what would be done without changes | `K0RDENTD_DRY_RUN` |

## Examples

### Example 1: Basic Installation

```bash
# Create a configuration file
cat > /etc/k0rdentd/k0rdentd.yaml <<EOF
k0s:
  version: "v1.27.4+k0s.0"
  api:
    address: "0.0.0.0"
    port: 6443
  network:
    provider: "calico"
    podCIDR: "10.244.0.0/16"
    serviceCIDR: "10.96.0.0/12"
  storage:
    type: "etcd"
    etcd:
      peerAddress: "127.0.0.1"

k0rdent:
  version: "v0.1.0"
  helm:
    chart: "k0rdent/k0rdent"
    namespace: "k0rdent-system"
    values:
      replicaCount: 1
      service:
        type: ClusterIP
        port: 80

debug: false
logLevel: "info"
EOF

# Install
k0rdentd install
```

### Example 2: Custom Network Configuration

```bash
cat > /etc/k0rdentd/k0rdentd.yaml <<EOF
k0s:
  version: "v1.27.4+k0s.0"
  api:
    address: "192.168.1.100"
    port: 6443
  network:
    provider: "calico"
    podCIDR: "10.244.0.0/16"
    serviceCIDR: "10.96.0.0/12"
  storage:
    type: "etcd"
    etcd:
      peerAddress: "127.0.0.1"

k0rdent:
  version: "v0.1.0"
  helm:
    chart: "k0rdent/k0rdent"
    namespace: "k0rdent-system"
    values:
      replicaCount: 1

debug: false
logLevel: "info"
EOF

k0rdentd install --config-file /etc/k0rdentd/k0rdentd.yaml
```

### Example 3: Development and Testing

```bash
# Enable debug logging
k0rdentd install --debug

# Dry-run to see what would happen
k0rdentd install --dry-run

# Check version
k0rdentd version
```

## Development

### Project Structure

```
.
├── cmd/
│   └── k0rdentd/
│       └── main.go          # CLI entry point
├── pkg/
│   ├── cli/                # CLI command implementations
│   │   ├── install.go
│   │   ├── uninstall.go
│   │   ├── version.go
│   │   └── config.go
│   ├── config/             # Configuration management
│   │   ├── k0rdentd.go
│   │   └── config_test.go
│   ├── generator/          # K0s config generation
│   │   └── generator.go
│   ├── installer/          # Installation logic
│   │   └── installer.go
│   └── utils/              # Utility functions
│       ├── logging.go
│       └── validation.go
├── internal/
│   └── test/               # Test utilities
├── examples/               # Example configurations
│   └── k0rdentd.yaml
├── scripts/                # Build and deployment scripts
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies
└── README.md               # Project documentation
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with ginkgo/gomega
ginkgo -v ./...
```

### Code Style

- Follow Go best practices and conventions
- Use `gofmt` for code formatting
- Maintain >80% code coverage
- Write unit tests for all business logic

## Logging

K0rdentd uses [logrus](https://github.com/sirupsen/logrus) for structured logging with the following levels:

- **DEBUG**: Detailed information for intermediate steps
- **INFO**: General information about completed tasks
- **WARN**: Unexpected behavior or warnings

Log levels can be configured via the `logLevel` setting in the configuration file or via the `--debug` flag.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [K0s](https://k0sproject.io/) - Kubernetes distribution
- [K0rdent](https://docs.k0rdent.io/) - Kubernetes management platform
- [RancherD](https://github.com/harvester/rancherd) - Inspiration for the architecture
- [urfave/cli](https://github.com/urfave/cli) - CLI framework
- [logrus](https://github.com/sirupsen/logrus) - Logging library

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/belgaied2/k0rdentd).