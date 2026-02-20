# Quick Start

Get K0rdent up and running in minutes with this quick start guide.

## Prerequisites

- Linux VM with sudo access
- Internet access (for online installation)
- At least 4 GB RAM and 10 GB disk space

## Step 1: Install K0rdentd

```bash
curl -sfL https://k0rdentd.belgai.de | sudo bash
```

## Step 2: Run Installation

```bash
# Install K0s and K0rdent
sudo k0rdentd install
```

The installation process will:

1. Check for and install K0s if needed
2. Generate K0s configuration
3. Start the K0s cluster
4. Deploy K0rdent using Helm
5. Wait for all components to be ready
6. Display access information

## Step 3: Check out K0rdent Enterprise UI
The previous step would have shown options to access the K0rdent UI. If one of these options is available from your workstation, 

Expected output:

```
NAME                                              READY   STATUS    RESTARTS   AGE
k0rdent-cert-manager-xxx                          1/1     Running   0          5m
k0rdent-cert-manager-cainjector-xxx               1/1     Running   0          5m
k0rdent-cert-manager-webhook-xxx                  1/1     Running   0          5m
k0rdent-datasource-controller-manager-xxx         1/1     Running   0          5m
k0rdent-k0rdent-enterprise-controller-manager-xxx 1/1     Running   0          5m
k0rdent-k0rdent-ui-xxx                            1/1     Running   0          5m
k0rdent-rbac-manager-xxx                          1/1     Running   0          5m
k0rdent-regional-telemetry-xxx                    1/1     Running   0          5m
```

## Step 5: Access K0rdent UI

After installation, K0rdentd will display URLs to access the K0rdent UI. You can also expose it manually:

```bash
# Expose the UI
sudo k0rdentd expose-ui
```

This will:

1. Create an ingress for the UI
2. Detect available IP addresses
3. Display access URLs

## Next Steps

Now that you have K0rdent running, you can:

- **Configure Cloud Credentials**: Add AWS, Azure, or OpenStack credentials
- **Create Clusters**: Use K0rdent to manage Kubernetes clusters
- **Explore the UI**: Access the web interface for cluster management

### Configure Cloud Credentials

Add credentials to your configuration:

```yaml
k0rdent:
  credentials:
    aws:
      - name: aws-prod
        region: us-east-1
        accessKeyID: AKIA...
        secretAccessKey: secret...
```

Then run:

```bash
sudo k0rdentd install
```

### Create Your First Cluster

Use K0rdent to create a cluster:

```bash
# Get kubeconfig
sudo k0s kubeconfig admin > ~/.kube/config

# Create a cluster using k0rdent CLI or UI
```

## Common Issues

### Installation Stuck

If installation seems stuck:

```bash
# Check logs
sudo journalctl -u k0scontroller -f

# Check pod status
sudo k0s kubectl get pods -A
```

### Pod Image Pull Errors

If pods fail to pull images:

```bash
# Check if images are being pulled
sudo k0s kubectl describe pod <pod-name> -n kcm-system

# Check network connectivity
curl -I https://registry.k8s.io
```

### UI Not Accessible

If you can't access the UI:

```bash
# Check ingress
sudo k0s kubectl get ingress -A

# Check service
sudo k0s kubectl get svc -n kcm-system

# Try port-forwarding
sudo k0s kubectl port-forward -n kcm-system svc/k0rdent-k0rdent-ui 8080:80
```

## Clean Up

To uninstall K0rdent and K0s:

```bash
# Uninstall everything
sudo k0rdentd uninstall
```

This will remove K0rdent and K0s from your system.
