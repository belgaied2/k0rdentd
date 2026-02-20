# Cloud Credentials

K0rdentd can automatically create cloud provider credentials for K0rdent during installation. This allows K0rdent to provision and manage Kubernetes clusters on AWS, Azure, and OpenStack.

## Overview

When you configure cloud credentials in `k0rdentd.yaml`, K0rdentd creates the necessary Kubernetes objects:

1. **Secret**: Stores sensitive credential data
2. **Identity**: Cloud provider identity (AWS/Azure only)
3. **Credential**: K0rdent credential object

## Configuration

### AWS Credentials

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

#### Required Fields

| Field | Description |
|-------|-------------|
| `name` | Credential name (used for all created resources) |
| `region` | Default AWS region |
| `accessKeyID` | AWS Access Key ID |
| `secretAccessKey` | AWS Secret Access Key |

#### Optional Fields

| Field | Description |
|-------|-------------|
| `sessionToken` | Session token for MFA or SSO authentication |

#### Created Resources

1. **Secret**: `<name>-secret`
2. **AWSClusterStaticIdentity**: `<name>-identity`
3. **Credential**: `<name>`

### Azure Credentials

```yaml
k0rdent:
  credentials:
    azure:
      - name: azure-prod-credentials
        subscriptionID: subscription-uuid...
        clientID: client-uuid...
        clientSecret: secret...
        tenantID: tenant-uuid...
```

#### Required Fields

| Field | Description |
|-------|-------------|
| `name` | Credential name |
| `subscriptionID` | Azure Subscription ID |
| `clientID` | Service Principal Client ID |
| `clientSecret` | Service Principal Client Secret |
| `tenantID` | Azure Tenant ID |

#### Created Resources

1. **Secret**: `<name>-secret`
2. **AzureClusterIdentity**: `<name>-identity`
3. **Credential**: `<name>`

### OpenStack Credentials

```yaml
k0rdent:
  credentials:
    openstack:
      - name: openstack-prod-credentials
        authURL: https://auth.example.com:5000/v3
        region: RegionOne
        applicationCredentialID: app-cred-id...
        applicationCredentialSecret: app-cred-secret...
```

#### Application Credential Authentication

| Field | Description |
|-------|-------------|
| `name` | Credential name |
| `authURL` | Keystone authentication URL |
| `region` | OpenStack region |
| `applicationCredentialID` | Application Credential ID |
| `applicationCredentialSecret` | Application Credential Secret |

#### Username/Password Authentication (Alternative)

```yaml
k0rdent:
  credentials:
    openstack:
      - name: openstack-prod-credentials
        authURL: https://auth.example.com:5000/v3
        region: RegionOne
        username: admin
        password: secret...
        projectName: myproject
        domainName: default
```

#### Created Resources

1. **Secret**: `<name>-config` (contains clouds.yaml)
2. **Credential**: `<name>`

## Multiple Credentials

You can configure multiple credentials for the same or different providers:

```yaml
k0rdent:
  credentials:
    aws:
      - name: aws-prod
        region: us-east-1
        accessKeyID: AKIA...
        secretAccessKey: secret...
      - name: aws-dev
        region: us-west-2
        accessKeyID: AKIA...
        secretAccessKey: secret...
    azure:
      - name: azure-prod
        subscriptionID: sub-id...
        clientID: client-id...
        clientSecret: secret...
        tenantID: tenant-id...
    openstack:
      - name: openstack-prod
        authURL: https://auth.example.com:5000/v3
        region: RegionOne
        applicationCredentialID: app-cred...
        applicationCredentialSecret: secret...
```

## Idempotency

Credential creation is idempotent:

- Each resource is checked before creation
- Existing resources are skipped
- Partial state recovery is supported
- Safe to run multiple times

## Timing Considerations

!!! warning "CAPI Provider Timing"
    Credentials are created after K0rdent is installed, but CAPI infrastructure providers may not be ready yet. If credential creation fails, you may need to wait and retry.

The installer waits for CAPI provider Helm releases to be deployed before creating credentials. If this times out, credential creation will log warnings but won't fail the installation.

## Security Best Practices

### 1. Use Environment Variables

Don't hardcode credentials in configuration files:

```yaml
k0rdent:
  credentials:
    aws:
      - name: aws-prod
        region: us-east-1
        accessKeyID: ${AWS_ACCESS_KEY_ID}
        secretAccessKey: ${AWS_SECRET_ACCESS_KEY}
```

### 2. Restrict File Permissions

```bash
# Set restrictive permissions on config file
sudo chmod 600 /etc/k0rdentd/k0rdentd.yaml
```

### 3. Use Short-Lived Credentials

For AWS, use IAM roles with temporary credentials:

```yaml
k0rdent:
  credentials:
    aws:
      - name: aws-prod
        region: us-east-1
        accessKeyID: ${AWS_ACCESS_KEY_ID}
        secretAccessKey: ${AWS_SECRET_ACCESS_KEY}
        sessionToken: ${AWS_SESSION_TOKEN}  # Short-lived
```

### 4. Audit Credential Usage

```bash
# List credentials
sudo k0s kubectl get credentials -n kcm-system

# Check credential details
sudo k0s kubectl describe credential aws-prod -n kcm-system
```

## Troubleshooting

### Credential Creation Failed

**Symptom**: Warning messages during installation about credential creation failure.

**Cause**: CAPI infrastructure providers not yet installed.

**Solution**: Wait for providers to be ready and retry:

```bash
# Check CAPI providers
sudo k0s kubectl get deployments -n capi-system

# Manually create credentials
sudo k0rdentd install  # Re-run to retry credential creation
```

### Secret Already Exists

**Symptom**: "Secret already exists" error.

**Cause**: Previous installation attempt created resources.

**Solution**: This is handled by idempotency - existing resources are skipped.

### Invalid Credentials

**Symptom**: Credentials created but K0rdent can't use them.

**Cause**: Invalid credential values.

**Solution**: Verify credentials manually:

```bash
# For AWS
aws sts get-caller-identity --profile test-profile

# For Azure
az login --service-principal -u CLIENT_ID -p CLIENT_SECRET --tenant TENANT_ID
```

## Manual Credential Creation

If automatic creation fails, you can create credentials manually:

### AWS Example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-prod-secret
  namespace: kcm-system
type: Opaque
stringData:
  AccessKeyID: AKIA...
  SecretAccessKey: secret...
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: AWSClusterStaticIdentity
metadata:
  name: aws-prod-identity
  namespace: kcm-system
spec:
  secretRef: aws-prod-secret
  allowedNamespaces:
    selector:
      matchLabels: {}
---
apiVersion: k0rdent.mirantis.com/v1beta1
kind: Credential
metadata:
  name: aws-prod
  namespace: kcm-system
spec:
  description: "AWS production credentials"
  identityRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
    kind: AWSClusterStaticIdentity
    name: aws-prod-identity
    namespace: kcm-system
```

Apply with:

```bash
sudo k0s kubectl apply -f credentials.yaml
```

## References

- [K0rdent Credentials Documentation](https://docs.k0rdent.io)
- [Cluster API AWS Provider](https://cluster-api-aws.sigs.k8s.io/)
- [Cluster API Azure Provider](https://cluster-api-azure.sigs.k8s.io/)
- [Cluster API OpenStack Provider](https://cluster-api-openstack.sigs.k8s.io/)
