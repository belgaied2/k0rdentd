# K0rdent Enterprise Airgap Bundle Catalog

**Bundle**: `airgap-bundle-1.2.3.tar.gz`
**Source**: https://get.mirantis.com/k0rdent-enterprise/1.2.3/airgap-bundle-1.2.3.tar.gz
**Download Size**: 21 GB (compressed)
**Extracted Size**: 22 GB
**Total Images**: 233 files
**Date Analyzed**: 2025-02-05

---

## Bundle Structure

The bundle contains:
- **Helm Charts**: Pre-packaged `.tar` files for all k0rdent components
- **Container Images**: OCI image archives organized by vendor/purpose
- **K0s Components**: Core k0s images (v1.32.8)

---

## Helm Charts (`charts/`)

### Core K0rdent Charts
| Chart | Version | Size |
|-------|---------|------|
| k0rdent-enterprise | 1.2.3 | 558 KB |
| kcm-regional | 1.2.3 | 348 KB |
| kcm-templates | 1.2.3 | 10 KB |
| k0rdent-ui | 1.1.1 | 13 KB |
| datasource-controller | 1.2.3 | 28 MB (image) |
| kcm-controller | 1.2.3 | 46 MB (image) |
| kcm-telemetry | 1.2.3 | 24 MB (image) |

### Cluster Infrastructure Providers

#### AWS
| Chart | Version |
|-------|---------|
| aws-eks | 1.0.4 |
| aws-hosted-cp | 1.0.21 |
| aws-standalone-cp | 1.0.20 |
| aws-cloud-controller-manager | 0.0.9 |
| aws-ebs-csi-driver | 2.33.0 |

#### Azure
| Chart | Version |
|-------|---------|
| azure-aks | 1.0.1 |
| azure-hosted-cp | 1.0.22 |
| azure-standalone-cp | 1.0.19 |
| cloud-provider-azure | 1.31.2 |
| azuredisk-csi-driver | v1.30.3 |

#### GCP
| Chart | Version |
|-------|---------|
| gcp-gke | 1.0.6 |
| gcp-hosted-cp | 1.0.19 |
| gcp-standalone-cp | 1.0.17 |
| gcp-cloud-controller-manager | 0.0.1 |
| gcp-compute-persistent-disk-csi-driver | 0.0.2 |

#### vSphere
| Chart | Version |
|-------|---------|
| vsphere-hosted-cp | 1.0.18 |
| vsphere-standalone-cp | 1.0.17 |
| vsphere-cpi | v1.31.0 |
| vsphere-csi-driver | 0.0.3 |

#### OpenStack
| Chart | Version |
|-------|---------|
| openstack-hosted-cp | 1.0.12 |
| openstack-standalone-cp | 1.0.21 |
| openstack-cloud-controller-manager | 2.31.1 |
| openstack-cinder-csi | 2.31.2 |

#### Docker
| Chart | Version |
|-------|---------|
| docker-hosted-cp | 1.0.4 |

#### K0smotron (K0s)
| Chart | Version |
|-------|---------|
| cluster-api-provider-k0sproject-k0smotron | 1.0.12 |

### Cluster API Core
| Chart | Version |
|-------|---------|
| cluster-api | 1.0.7 |
| cluster-api-operator | 0.24.0 |
| cluster-api-visualizer | 1.4.0 |

### Observability & Monitoring

**OpenCost**
| Chart | Version |
|-------|---------|
| opencost | 2.4.0 |

**OpenTelemetry**
| Chart | Version |
|-------|---------|
| opentelemetry-operator | 0.84.2 |
| opentelemetry-operator | 0.98.0 |
| opentelemetry-kube-stack | 0.5.3 |

**Monitoring Stack**
| Chart | Version |
|-------|---------|
| metrics-server | (included via image) |
| kube-state-metrics | 5.21.0 |
| prometheus-node-exporter | 4.37.3 |
| prometheus-operator-crds | 15.0.0 |

**Grafana & VictoriaMetrics**
| Chart | Version |
|-------|---------|
| grafana-operator | v5.18.0 |
| victoria-metrics-operator | 0.43.1 |
| victoria-logs-cluster | 0.0.2 |

**Jaeger**
| Chart | Version |
|-------|---------|
| jaeger-operator | 2.50.1 |

**Logging**
| Chart | Version |
|-------|---------|
| vector | 0.40.0 |

### Networking & Services
| Chart | Version |
|-------|---------|
| ingress-nginx | 4.12.1 |
| istiod | 1.25.3 |
| k0rdent-istio | 0.1.0 |
| k0rdent-istio-base | 0.1.0 |
| cert-manager-istio-csr | v0.14.0 |
| kube-vip | 0.6.1 |
| external-dns | 1.15.2 |

### Security & IAM
| Chart | Version |
|-------|---------|
| dex | 0.23.0 |
| rbac-manager | 1.21.2 |
| cert-manager | v1.16.4 |
| cert-manager | v1.19.1 |

### Backup & Storage
| Chart | Version |
|-------|---------|
| velero | 11.0.0 |

### Flux CD
| Chart | Version |
|-------|---------|
| flux2 | 2.16.4 |

### KOF (Kubernetes Observability Framework)
| Chart | Version |
|-------|---------|
| kof-mothership | 1.5.0 |
| kof-storage | 1.5.0 |
| kof-operators | 1.5.0 |
| kof-collectors | 1.5.0 |
| kof-dashboards | 1.5.0 |
| kof-child | 1.5.0 |
| kof-regional | 1.5.0 |

### Sveltos
| Chart | Version |
|-------|---------|
| projectsveltos | 1.1.1 |
| sveltos-crds | 1.1.1 |

### Base Components
| Chart | Version |
|-------|---------|
| base | 1.25.3 |
| gateway | 1.25.5 |

### Other
| Chart | Version |
|-------|---------|
| adopted-cluster | 1.0.1 |
| remote-cluster | 1.0.18 |

---

## Container Images

### Cluster API Controllers (`capi/`)
| Image | Version | Size |
|-------|---------|------|
| cluster-api-controller | v1.11.2 | 191 MB |
| cluster-api-aws-controller | v2.10.0 | 354 MB |
| cluster-api-azure-controller | v1.21.0 | 265 MB |
| cluster-api-gcp-controller | v1.10.0 | 254 MB |
| cluster-api-vsphere-controller | v1.14.0 | 204 MB |
| capi-openstack-controller | v0.12.5-mirantis.0 | 166 MB |
| capd-manager (Docker) | v1.11.2 | 191 MB |
| k0smotron | v1.10.0 | 86 MB |

### K0s Components (`k0sproject/`)
| Image | Version | Purpose |
|-------|---------|---------|
| k0s | v1.32.8-k0s.0 | Main k0s binary |
| etcd | v3.5.13 | etcd datastore |
| coredns | 1.12.2 | DNS service |
| kube-proxy | v1.32.8 | Kubernetes network proxy |
| calico-node | v3.29.4-0 | Calico CNI |
| calico-cni | v3.29.4-0 | Calico CNI plugin |
| calico-kube-controllers | v3.29.4-0 | Calico controllers |
| kube-router | v2.4.1-iptables1.8.9-0 | Alternative CNI |
| cni-node | 1.3.0-k0s.0 | CNI plugins |
| apiserver-network-proxy-agent | v0.31.0 | Network proxy |
| envoy-distroless | v1.31.10 | Envoy proxy |
| pushgateway-ttl | 1.4.0-k0s.0 | Metrics push gateway |

### Observability Images

**Metrics Server** (`metrics-server/`)
- metrics-server: v0.7.2

**OpenCost** (`opencost/`)
- opencost: 1.118.0
- opencost-ui: 1.118.0

**OpenTelemetry** (`opentelemetry-operator/`)
- opentelemetry-operator: 0.120.0
- opentelemetry-operator: 0.137.0
- target-allocator: v0.140.0

**Other Observability** (`otel/`, `prom/`, `grafana/`, `jaegertracing/`, `victoriametrics/`, `timberio/`)
- Various OpenTelemetry collectors and instrumentation images
- Prometheus components
- Grafana images
- Jaeger components
- VictoriaMetrics components
- Vector logger

### Kubernetes Images (`k8s/`)
- pause: 3.9

### Cloud Provider Images

**AWS** (`provider-aws/`, `ebs-csi-driver/`)
- cloud-controller-manager: v1.30.3

**Azure** (root level)
- azure-cloud-controller-manager: v1.32.4
- azure-cloud-node-manager: v1.32.4

**OpenStack** (`provider-os/`)
- openstack-cloud-controller-manager: v1.31.1
- cinder-csi-plugin: v1.31.0

**GCP** (`cloud-provider-gcp/`, `k8s-staging-cloud-provider-gcp/`)
- Various GCP cloud provider images

**vSphere** (`csi-vsphere/`, `cloud-pv-vsphere/`)
- vSphere CSI and cloud provider images

### CSI Drivers
- AWS EBS CSI driver
- Azure Disk CSI driver
- GCP Persistent Disk CSI driver
- OpenStack Cinder CSI
- vSphere CSI
- Various external-snapshotter and Kubernetes CSI images

### Networking & Ingress
- ingress-nginx images (`ingress-nginx/`)
- istio images (`istio/`)
- kube-vip images (`kube-vip/`)
- external-dns images (`external-dns/`)

### Security
- cert-manager images (`jetstack/`)
- dex images (`dexidp/`)

### Storage & Backup
- velero images (`velero/`)

### Flux CD
- flux2 images (`fluxcd/`)

### Core Infrastructure
- etcd images
- various alpine and busybox images

---

## Key Findings for K0rdentd Implementation

### ✅ What IS Included in the Bundle

1. **Cluster API Providers**: ALL major infrastructure providers are included
   - AWS (EKS)
   - Azure (AKS)
   - GCP (GKE)
   - OpenStack
   - vSphere
   - Docker
   - K0smotron (K0s-native)

2. **Helm Chart Dependencies**: ALL dependencies are bundled
   - OpenCost ✅
   - OpenTelemetry Operator ✅
   - OpenTelemetry Kube Stack ✅
   - Metrics Server ✅
   - Prometheus Operator ✅
   - Grafana Operator ✅
   - cert-manager ✅
   - And many more...

3. **K0s Images**: Core k0s v1.32.8 components included
   - k0s binary container
   - etcd, CoreDNS, kube-proxy
   - Calico CNI
   - All standard k0s components

4. **CSI Drivers**: All major cloud storage drivers included
   - AWS EBS
   - Azure Disk
   - GCP Persistent Disk
   - OpenStack Cinder
   - vSphere

### ❌ What is NOT Included

1. **K0s Binary Executable**: The standalone `k0s` binary for the host OS
   - Must be downloaded separately from k0s releases
   - Expected to be hosted on HTTP server per enterprise docs

2. **Multi-Architecture Images**: Bundle appears to be single-architecture
   - Likely amd64 only (needs verification)
   - arm64 would require separate bundles

### Implications for K0rdentd

1. **Single Bundle Sufficient**: The k0rdent enterprise bundle contains everything needed
   - No need to source CAPI provider images separately
   - No need to download Helm chart dependencies separately
   - Just need to add the k0s binary

2. **Architecture Decision**:
   - For Phase 1, focus on amd64 (enterprise bundle is likely amd64)
   - For multi-arch, investigate if k0rdent provides arm64 bundles

3. **Distribution Strategy**:
   - Download k0rdent bundle from Mirantis
   - Download k0s binary for target architecture
   - Bundle both in k0rdentd airgap binary via `//go:embed`

---

## Directory Organization

```
airgap-bundle-1.2.3/
├── charts/                    # Helm charts (83 .tar files)
├── capi/                      # Cluster API provider controllers
├── k0sproject/                # K0s component images
├── provider-aws/              # AWS cloud provider
├── provider-os/               # OpenStack cloud provider
├── metrics-server/            # Metrics server images
├── opencost/                  # OpenCost images
├── opentelemetry-operator/    # OpenTelemetry images
├── [40+ vendor directories]   # Other component images
└── [standalone .tar files]    # Additional images
```

---

## Summary

The k0rdent enterprise airgap bundle is comprehensive and includes:
- ✅ All Cluster API infrastructure providers
- ✅ All Helm chart dependencies
- ✅ All observability stack components
- ✅ K0s container images
- ❌ Host k0s binary (must be added separately)

**Conclusion**: For k0rdentd, we only need to:
1. Download the k0rdent airgap bundle
2. Download the appropriate k0s binary
3. Embed both in the k0rdentd-airgap binary
