# Configuration

K0rdentd is configured via a YAML configuration file, by default located at `/etc/k0rdentd/k0rdentd.yaml`.

## Configuration File Location

The configuration file location can be specified in several ways:

1. **Default**: `/etc/k0rdentd/k0rdentd.yaml`
2. **CLI Flag**: `--config-file /path/to/config.yaml`
3. **Environment Variable**: `K0RDENTD_CONFIG_FILE=/path/to/config.yaml`

## Complete Configuration Reference

Here's a complete example configuration file with all available options:

```yaml
# K0s Configuration
k0s:
  version: "v1.32.4+k0s.0"
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

# K0rdent Configuration
k0rdent:
  version: "v0.1.0"
  helm:
    chart: "k0rdent/k0rdent"
    namespace: "kcm-system"
    values:
      replicaCount: 1
      service:
        type: ClusterIP
        port: 80

# Cloud Provider Credentials (Optional)
k0rdent:
  credentials:
    aws:
      - name: aws-prod-credentials
        region: us-east-1
        accessKeyID: AKIA...
        secretAccessKey: secret...
        sessionToken: "..."  # Optional: for MFA or SSO
    azure:
      - name: azure-prod-credentials
        subscriptionID: sub-id...
        clientID: client-id...
        clientSecret: secret...
        tenantID: tenant-id...
    openstack:
      - name: openstack-prod-credentials
        authURL: https://auth.example.com:5000/v3
        region: RegionOne
        applicationCredentialID: app-cred-id...
        applicationCredentialSecret: app-cred-secret...

# Airgap Configuration (Optional)
airgap:
  registry:
    address: localhost:5000
  bundlePath: /opt/k0rdent/airgap-bundle-1.2.3.tar.gz

# Global Settings
debug: false
logLevel: "info"
```

## Configuration Sections

### K0s Configuration

The `k0s` section configures the underlying K0s cluster:

```yaml
k0s:
  # K0s version to install
  version: "v1.32.4+k0s.0"
  
  # API server configuration
  api:
    address: "0.0.0.0"  # API server bind address
    port: 6443          # API server port
  
  # Network configuration
  network:
    provider: "calico"           # CNI provider: calico, kuberouter
    podCIDR: "10.244.0.0/16"     # Pod CIDR range
    serviceCIDR: "10.96.0.0/12"  # Service CIDR range
  
  # Storage configuration
  storage:
    type: "etcd"  # Storage type: etcd, kine
    etcd:
      peerAddress: "127.0.0.1"  # etcd peer address
```

### K0rdent Configuration

The `k0rdent` section configures the K0rdent deployment:

```yaml
k0rdent:
  # K0rdent version
  version: "v0.1.0"
  
  # Helm chart configuration
  helm:
    chart: "k0rdent/k0rdent"  # Helm chart reference
    namespace: "kcm-system"   # Installation namespace
    
    # Custom Helm values
    values:
      replicaCount: 1
      service:
        type: ClusterIP
        port: 80
```

### Cloud Credentials Configuration

The `k0rdent.credentials` section configures cloud provider credentials:

#### AWS Credentials

```yaml
k0rdent:
  credentials:
    aws:
      - name: aws-prod-credentials
        region: us-east-1
        accessKeyID: AKIA...
        secretAccessKey: secret...
        sessionToken: "..."  # Optional: for MFA or SSO
```

#### Azure Credentials

```yaml
k0rdent:
  credentials:
    azure:
      - name: azure-prod-credentials
        subscriptionID: subscription-uuid
        clientID: client-uuid
        clientSecret: secret...
        tenantID: tenant-uuid
```

#### OpenStack Credentials

```yaml
k0rdent:
  credentials:
    openstack:
      - name: openstack-prod-credentials
        authURL: https://auth.example.com:5000/v3
        region: RegionOne
        # Option 1: Application Credentials
        applicationCredentialID: app-cred-id...
        applicationCredentialSecret: app-cred-secret...
        # Option 2: Username/Password (alternative)
        # username: admin
        # password: secret
        # projectName: myproject
        # domainName: default
```

### Airgap Configuration

The `airgap` section configures airgap (offline) installation:

```yaml
airgap:
  registry:
    address: localhost:5000  # Local registry address
    # For multi-worker: use reachable IP
    # address: 192.168.1.10:5000
  
  # Path to k0rdent airgap bundle
  bundlePath: /opt/k0rdent/airgap-bundle-1.2.3.tar.gz
```

### Global Settings

```yaml
# Enable debug logging
debug: false

# Log level: debug, info, warn, error
logLevel: "info"
```

## Environment Variables

All configuration options can be overridden with environment variables:

| Environment Variable | Configuration Path |
|---------------------|-------------------|
| `K0RDENTD_CONFIG_FILE` | Configuration file path |
| `K0RDENTD_DEBUG` | `debug` |
| `K0RDENTD_LOG_LEVEL` | `logLevel` |
| `K0RDENTD_K0S_VERSION` | `k0s.version` |
| `K0RDENTD_K0RDENT_VERSION` | `k0rdent.version` |
| `K0RDENTD_AIRGAP_BUNDLE_PATH` | `airgap.bundlePath` |
| `K0RDENTD_REGISTRY_ADDRESS` | `airgap.registry.address` |

Example:

```bash
export K0RDENTD_DEBUG=true
export K0RDENTD_K0S_VERSION=v1.32.4+k0s.0
sudo k0rdentd install
```

## Generated K0s Configuration

K0rdentd generates the K0s configuration file at `/etc/k0s/k0s.yaml` based on your `k0rdentd.yaml` settings. The generated configuration includes:

- API server settings
- Network configuration
- Storage configuration
- Helm extensions for K0rdent

Example generated `k0s.yaml`:

```yaml
apiVersion: k0s.k0sproject.io/v1beta1
kind: Cluster
metadata:
  name: k0s
spec:
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
  extensions:
    helm:
      repositories:
        - name: k0rdent
          url: https://charts.k0rdent.io
      charts:
        - name: k0rdent
          chartname: k0rdent/k0rdent
          version: "v0.1.0"
          namespace: kcm-system
          values: |
            replicaCount: 1
            service:
              type: ClusterIP
              port: 80
```

## Validation

K0rdentd validates the configuration before starting:

- Required fields are present
- Version formats are correct
- Network CIDRs are valid
- File paths exist (for airgap bundle)

If validation fails, K0rdentd will log an error and exit.

## Example Configurations

### Minimal Configuration

```yaml
k0s:
  version: "v1.32.4+k0s.0"

k0rdent:
  version: "v0.1.0"
```

### Production Configuration

```yaml
k0s:
  version: "v1.32.4+k0s.0"
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
    namespace: "kcm-system"
  credentials:
    aws:
      - name: aws-prod
        region: us-east-1
        accessKeyID: ${AWS_ACCESS_KEY_ID}
        secretAccessKey: ${AWS_SECRET_ACCESS_KEY}

debug: false
logLevel: "info"
```

### Airgap Configuration

```yaml
k0s:
  version: "v1.32.4+k0s.0"

k0rdent:
  version: "v0.1.0"

airgap:
  registry:
    address: localhost:5000
  bundlePath: /opt/k0rdent/airgap-bundle-1.2.3.tar.gz

debug: false
logLevel: "info"
```
