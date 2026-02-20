# Airgap Installation

This guide covers installing K0rdent in airgap (offline) environments using K0rdentd.

## Overview

Airgap installation allows you to deploy K0s and K0rdent in environments without internet access. The process involves:

1. **Online Machine**: Download artifacts
2. **Airgap Machine**: Transfer and install

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              VM-online (Internet)                            │
│                                                                              │
│  1. Download k0rdentd-airgap binary from GitHub                             │
│  2. Download K0rdent Enterprise Airgap Bundle from Mirantis                 │
│  3. Transfer artifacts to airgap VMs via SSH                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     │ SSH/SCP
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     k0rdent-airgap-* (No Internet)                          │
│                                                                              │
│  1. Receive artifacts from VM-online                                        │
│  2. Start local OCI registry daemon                                         │
│  3. Run k0rdentd install (airgap mode)                                      │
│  4. K0s cluster with K0rdent ready                                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### On Online Machine

- Internet access
- SSH client
- `curl` or `wget`
- `cosign` (optional, for bundle verification)

### On Airgap Machine

- No internet access required
- SSH server running
- Linux AMD64 or ARM64
- **Recommended disk**: 100 GB
- ~50 GB for bundle extraction and registry storage
- ~10 GB additional for K0s and K0rdent
- `sudo` privileges

## Phase 1: Prepare on Online Machine

### Step 1.1: Download K0rdentd Airgap Binary

```bash
# Set the version
K0RDENTD_VERSION="v0.2.0"

# Download for AMD64
curl -L "https://github.com/belgaied2/k0rdentd/releases/download/${K0RDENTD_VERSION}/k0rdentd-airgap-${K0RDENTD_VERSION}-linux-amd64.tar.gz" -o k0rdentd-airgap.tar.gz

# Or for ARM64
curl -L "https://github.com/belgaied2/k0rdentd/releases/download/${K0RDENTD_VERSION}/k0rdentd-airgap-${K0RDENTD_VERSION}-linux-arm64.tar.gz" -o k0rdentd-airgap.tar.gz
```

### Step 1.2: Download K0rdent Enterprise Airgap Bundle

```bash
# Set the K0rdent version
K0RDENT_VERSION="1.2.2"

# Download the bundle (~22 GB)
curl -L "https://get.mirantis.com/k0rdent-enterprise/${K0RDENT_VERSION}/airgap-bundle-${K0RDENT_VERSION}.tar.gz" -o airgap-bundle-${K0RDENT_VERSION}.tar.gz
```

### Step 1.3: Transfer Artifacts

```bash
# Set target VM details
TARGET_USER="ubuntu"
TARGET_VM="k0rdent-airgap-01"  # Replace with actual hostname/IP

# Transfer all artifacts
scp k0rdentd-airgap.tar.gz ${TARGET_USER}@${TARGET_VM}:/tmp/
scp airgap-bundle-${K0RDENT_VERSION}.tar.gz ${TARGET_USER}@${TARGET_VM}:/tmp/
```

## Phase 2: Install on Bootstrap Node

SSH into the bootstrap node:

```bash
ssh ${TARGET_USER}@k0rdent-airgap-01
```

### Step 2.1: Install K0rdentd Binary

```bash
# Extract and install
tar -xzf /tmp/k0rdentd-airgap.tar.gz
sudo mkdir -p /usr/local/bin
sudo cp k0rdentd-airgap /usr/local/bin/k0rdentd
sudo chmod +x /usr/local/bin/k0rdentd

# Verify installation
k0rdentd version
k0rdentd show-flavor  # Should output "airgap"
```

### Step 2.2: Create Configuration

```bash
sudo mkdir -p /etc/k0rdentd

sudo tee /etc/k0rdentd/k0rdentd.yaml > /dev/null <<EOF
k0s:
  version: "v1.32.4+k0s.0"

k0rdent:
  version: "1.2.2"

airgap:
  registry:
    address: localhost:5000
  bundlePath: /opt/k0rdent/airgap-bundle-1.2.2.tar.gz

debug: false
logLevel: "info"
EOF

# Move bundle to configured location
sudo mkdir -p /opt/k0rdent
sudo mv /tmp/airgap-bundle-1.2.2.tar.gz /opt/k0rdent/
```

### Step 2.3: Start the OCI Registry Daemon

!!! warning "Important"
    The registry daemon must be running before installation. Open a new terminal or run in background.

```bash
# Terminal 1: Start registry daemon
sudo k0rdentd registry \
  --bundle-path /opt/k0rdent/airgap-bundle-1.2.2.tar.gz \
  --port 5000 \
  --storage /var/lib/k0rdentd/registry
```

!!! note "No Feedback"
    As of v0.2.0, this step might seem like it's hanging because uncompressing the 22GB+ tar file takes time. You can verify tar is running with `top` or `ps -ef | grep tar`.

**Expected output:**

```
INFO[0000] Starting k0rdentd registry daemon...
INFO[0000] Configuration:
INFO[0000]   Bundle: /opt/k0rdent/airgap-bundle-1.2.2.tar.gz
INFO[0000]   Storage: /var/lib/k0rdentd/registry
INFO[0000]   Address: 0.0.0.0:5000
INFO[0000] Pushing images from bundle to local registry...
INFO[0001] [1/233] Pushing kcm-controller:1.2.2...
...
INFO[1235] Registry server listening on 0.0.0.0:5000
```

### Step 2.4: Run Installation

In a new terminal (or after starting registry in background):

```bash
# Terminal 2: Run installation
sudo k0rdentd install
```

The installation will:

1. Extract embedded K0s binary to `/usr/local/bin/k0s`
2. Configure containerd registry mirrors
3. Generate K0s configuration for airgap mode
4. Install and start K0s cluster
5. Wait for K0rdent to be installed via K0s Helm extension
6. Create cloud provider credentials (if configured)

### Step 2.5: Verify Installation

```bash
# Check K0s status
sudo k0s status

# Check K0rdent pods
sudo k0s kubectl get pods -n kcm-system
```

Expected output: All pods in Running state.

### Step 2.6: Access K0rdent UI

The CLI displays URLs to access the UI. For SSH port forwarding:

```bash
# Get NodePort
NODEPORT=$(sudo k0s kubectl get svc -n kcm-system k0rdent-k0rdent-ui -o jsonpath='{.spec.ports[0].nodePort}')

# On your local machine
ssh -J vm-online -L 8080:localhost:$NODEPORT $TARGET_USER@$TARGET_VM
```

Then access: `http://localhost:8080/k0rdent-ui`

## Phase 3: Add Worker Nodes (Optional)

### Step 3.1: Generate Join Token on Bootstrap

```bash
# On bootstrap node
sudo k0s token create --role=worker > /tmp/worker-token
```

### Step 3.2: Export Worker Artifacts

```bash
# On bootstrap node
k0rdentd export-worker-artifacts \
  --output /tmp/worker-bundle \
  --bundle-path /opt/k0rdent/airgap-bundle-1.2.2.tar.gz

# Transfer to workers
scp -r /tmp/worker-bundle ${TARGET_USER}@k0rdent-airgap-02:/tmp/
```

### Step 3.3: Join Worker Nodes

On each worker node:

```bash
# Install K0s binary
sudo cp /tmp/worker-bundle/k0s /usr/local/bin/k0s
sudo chmod +x /usr/local/bin/k0s

# Configure containerd to use bootstrap node's registry
BOOTSTRAP_IP="192.168.1.10"  # Replace with actual IP

sudo mkdir -p /etc/k0s/containerd.d/certs.d/registry.k8s.io
sudo mkdir -p /etc/k0s/containerd.d/certs.d/quay.io

cat <<EOF | sudo tee /etc/k0s/containerd.d/certs.d/registry.k8s.io/hosts.toml
server = "https://registry.k8s.io"
[host."http://${BOOTSTRAP_IP}:5000"]
  capabilities = ["pull", "resolve"]
EOF

cat <<EOF | sudo tee /etc/k0s/containerd.d/certs.d/quay.io/hosts.toml
server = "https://quay.io"
[host."http://${BOOTSTRAP_IP}:5000"]
  capabilities = ["pull", "resolve"]
EOF

# Join the cluster
sudo k0s worker $(cat /tmp/worker-token)
```

## CLI Commands Reference

### Registry Daemon

```bash
# Start registry (foreground)
sudo k0rdentd registry \
  --bundle-path /opt/k0rdent/airgap-bundle-1.2.2.tar.gz \
  --port 5000 \
  --storage /var/lib/k0rdentd/registry

# With signature verification
sudo k0rdentd registry \
  --bundle-path /opt/k0rdent/airgap-bundle-1.2.2.tar.gz \
  --verify true \
  --cosignKey https://get.mirantis.com/cosign.pub

# For multi-worker: bind to all interfaces
sudo k0rdentd registry \
  --bundle-path /opt/k0rdent/airgap-bundle-1.2.2.tar.gz \
  --host 0.0.0.0 \
  --port 5000
```

### Installation

```bash
# Install with config file
sudo k0rdentd install --config-file /etc/k0rdentd/k0rdentd.yaml

# Debug mode
sudo k0rdentd install --debug

# Dry-run
sudo k0rdentd install --dry-run
```

## File Locations

| File | Location | Purpose |
|------|----------|---------|
| K0rdentd binary | `/usr/local/bin/k0rdentd` | Main CLI tool |
| K0s binary | `/usr/local/bin/k0s` | Kubernetes distribution |
| K0rdentd config | `/etc/k0rdentd/k0rdentd.yaml` | Main configuration |
| K0s config | `/etc/k0s/k0s.yaml` | K0s cluster configuration |
| Bundle location | `/opt/k0rdent/airgap-bundle-*.tar.gz` | K0rdent airgap bundle |
| Registry storage | `/var/lib/k0rdentd/registry` | OCI registry data |
| Containerd config | `/etc/k0s/containerd.d/` | Containerd drop-in configs |

## Troubleshooting

### Registry Issues

**Port already in use:**

```bash
# Check what's using port 5000
sudo lsof -i :5000

# Use a different port
sudo k0rdentd registry --port 5001 ...
```

**Images fail to push:**

```bash
# Check disk space
df -h /var/lib/k0rdentd/registry

# Check bundle integrity
tar -tzf /opt/k0rdent/airgap-bundle-*.tar.gz | head
```

### Installation Issues

**Pods stuck in ImagePullBackOff:**

```bash
# Verify containerd mirror configuration
cat /etc/k0s/containerd.d/cri-registry.toml
cat /etc/k0s/containerd.d/certs.d/quay.io/hosts.toml

# Verify registry is accessible
curl http://localhost:5000/v2/_catalog

# Check K0s logs
sudo journalctl -u k0scontroller -f
```

**K0rdent pods not starting:**

```bash
# Check K0rdent Helm release status
sudo k0s kubectl get helmchart -n kube-system

# Check K0rdent logs
sudo k0s kubectl logs -n kcm-system -l app.kubernetes.io/name=kcm
```

### Verification Commands

```bash
# Check K0s cluster status
sudo k0s status
sudo k0s kubectl get nodes

# Check K0rdent installation
sudo k0s kubectl get pods -n kcm-system
sudo k0s kubectl get clustermanagement -A

# Check local registry
curl http://localhost:5000/v2/_catalog | jq
curl http://localhost:5000/v2/kcm-controller/tags/list | jq
```

## References

- [K0rdentd GitHub Repository](https://github.com/belgaied2/k0rdentd)
- [K0rdent Enterprise Documentation](https://docs.mirantis.com/k0rdent-enterprise/)
- [K0s Airgap Installation](https://docs.k0sproject.io/stable/airgap-install/)
- [K0rdent Airgap Installation](https://docs.mirantis.com/k0rdent-enterprise/latest/admin/installation/airgap/airgap-install/)
- [Containerd Registry Configuration](https://github.com/containerd/containerd/blob/main/docs/hosts.md)
