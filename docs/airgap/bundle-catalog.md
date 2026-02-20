# K0rdent Enterprise Airgap Bundle Catalog

This document describes the contents of the K0rdent Enterprise Airgap Bundle.

## Bundle Overview

| Property | Value |
|----------|-------|
| **Bundle** | `airgap-bundle-1.2.3.tar.gz` |
| **Source** | https://get.mirantis.com/k0rdent-enterprise/1.2.3/airgap-bundle-1.2.3.tar.gz |
| **Download Size** | 21 GB (compressed) |
| **Extracted Size** | 22 GB |
| **Total Images** | 233 files |

## Bundle Structure

```
airgap-bundle-1.2.3/
├── charts/                    # Helm charts (83 .tar files)
├── capi/                      # Cluster API provider images
├── k0sproject/                # K0s component images
├── provider-aws/              # AWS cloud provider
├── provider-os/               # OpenStack cloud provider
├── metrics-server/            # Metrics server images
├── opencost/                  # OpenCost images
├── opentelemetry-operator/    # OpenTelemetry images
└── [40+ vendor directories]   # Other component images
```

## Helm Charts

### Core K0rdent Charts

| Chart | Version | Size |
|-------|---------|------|
| k0rdent-enterprise | 1.2.3 | 558 KB |
| kcm-regional | 1.2.3 | 348 KB |
| kcm-templates | 1.2.3 | 10 KB |
| k0rdent-ui | 1.1.1 | 13 KB |

### Cluster API Providers

#### AWS
- aws-eks (1.0.4)
- aws-hosted-cp (1.0.21)
- aws-standalone-cp (1.0.20)

#### Azure
- azure-aks (1.0.1)
- azure-hosted-cp (1.0.22)
- azure-standalone-cp (1.0.19)

#### GCP
- gcp-gke (1.0.6)
- gcp-hosted-cp (1.0.19)
- gcp-standalone-cp (1.0.17)

#### vSphere
- vsphere-hosted-cp (1.0.18)
- vsphere-standalone-cp (1.0.17)

#### OpenStack
- openstack-hosted-cp (1.0.12)
- openstack-standalone-cp (1.0.21)

#### K0smotron
- cluster-api-provider-k0sproject-k0smotron (1.0.12)

### Observability Stack

- OpenCost (2.4.0)
- OpenTelemetry Operator (0.84.2, 0.98.0)
- OpenTelemetry Kube Stack (0.5.3)
- Prometheus Operator CRDs (15.0.0)
- Grafana Operator (v5.18.0)
- VictoriaMetrics Operator (0.43.1)
- Jaeger Operator (2.50.1)

### Networking

- Ingress NGINX (4.12.1)
- Istio (1.25.3)
- Kube-VIP (0.6.1)
- External-DNS (1.15.2)

### Security

- Cert-Manager (v1.16.4, v1.19.1)
- Dex (0.23.0)
- RBAC Manager (1.21.2)

### Backup

- Velero (11.0.0)

### GitOps

- Flux2 (2.16.4)

## Container Images

### Cluster API Controllers

| Image | Version | Size |
|-------|---------|------|
| cluster-api-controller | v1.11.2 | 191 MB |
| cluster-api-aws-controller | v2.10.0 | 354 MB |
| cluster-api-azure-controller | v1.21.0 | 265 MB |
| cluster-api-gcp-controller | v1.10.0 | 254 MB |
| cluster-api-vsphere-controller | v1.14.0 | 204 MB |
| capi-openstack-controller | v0.12.5 | 166 MB |
| k0smotron | v1.10.0 | 86 MB |

### K0s Components

| Image | Version | Purpose |
|-------|---------|---------|
| k0s | v1.32.8-k0s.0 | Main K0s binary |
| etcd | v3.5.13 | etcd datastore |
| coredns | 1.12.2 | DNS service |
| kube-proxy | v1.32.8 | Network proxy |
| calico-node | v3.29.4-0 | Calico CNI |
| calico-cni | v3.29.4-0 | Calico CNI plugin |

### Observability Images

- Metrics Server (v0.7.2)
- OpenCost (1.118.0)
- OpenTelemetry Collector (various)
- Prometheus components
- Grafana components
- Jaeger components
- VictoriaMetrics components

## What's Included

### ✅ Included

- All Cluster API infrastructure providers
- All Helm chart dependencies
- All observability stack components
- K0s container images
- CSI drivers for all providers

### ❌ Not Included

- **K0s binary executable** - Must be downloaded separately
  - Embedded in k0rdentd-airgap binary
  - Or downloaded from K0s releases

## Downloading the Bundle

```bash
# Set version
K0RDENT_VERSION="1.2.2"

# Download
curl -L "https://get.mirantis.com/k0rdent-enterprise/${K0RDENT_VERSION}/airgap-bundle-${K0RDENT_VERSION}.tar.gz" -o airgap-bundle.tar.gz
```

## Verifying the Bundle

```bash
# With cosign
cosign verify-blob \
  --key https://get.mirantis.com/cosign.pub \
  --signature airgap-bundle-1.2.2.tar.gz.sig \
  airgap-bundle-1.2.2.tar.gz
```

## Using with K0rdentd

```bash
# Start registry with bundle
sudo k0rdentd registry --bundle-path /path/to/airgap-bundle.tar.gz

# Install K0rdent
sudo k0rdentd install
```

## References

- [Airgap Installation Guide](../user-guide/airgap-installation.md)
- [Airgap Feature Documentation](airgap.md)
- [K0rdent Enterprise Documentation](https://docs.mirantis.com/k0rdent-enterprise/)
