# K0rdentd Documentation

Welcome to the K0rdentd documentation! K0rdentd is a CLI tool that automates the deployment of K0s and K0rdent on a VM. It follows a similar pattern to RancherD but simplifies the architecture by directly configuring K0s and using its built-in helm extension mechanism.

## What is K0rdentd?

K0rdentd simplifies the installation and management of K0rdent, a Kubernetes cluster management platform. It handles:

- **K0s Installation**: Automatically downloads and installs K0s binary
- **K0rdent Deployment**: Deploys K0rdent using K0s's Helm extension mechanism
- **Cloud Provider Credentials**: Creates credentials for AWS, Azure, and OpenStack
- **Airgap Support**: Full offline installation support for isolated environments
- **UI Exposure**: Automatically exposes the K0rdent UI for easy access

## Key Features

### 🚀 Simple Installation

```bash
# Download K0rdentd
curl -sfL https://k0rdentd.belgai.de | sudo bash

# Install K0s and K0rdent in one command
sudo k0rdentd install
```

### 🔌 Cloud Provider Support

Configure credentials for multiple cloud providers:

```yaml
k0rdent:
  credentials:
    aws:
      - name: aws-prod
        accessKeyID: AKIA...
        secretAccessKey: secret...
    azure:
      - name: azure-prod
        clientID: client-id...
        clientSecret: secret...
        tenantID: tenant-id...
```

### 🌐 Airgap Installation

Full offline installation support for air-gapped environments:

```bash
# Start local registry
sudo k0rdentd registry --bundle-path /path/to/bundle.tar.gz

# Install in airgap mode
sudo k0rdentd install --airgap
```

## Quick Links

<div class="grid cards" markdown>

-   :rocket:{ .lg .middle } __Getting Started__

    ---

    Get up and running with K0rdentd in minutes

    [:octicons-arrow-right-24: Installation](getting-started/installation.md)

-   :book:{ .lg .middle } __Architecture__

    ---

    Understand how K0rdentd works under the hood

    [:octicons-arrow-right-24: Overview](architecture/overview.md)

-   :wrench:{ .lg .middle } __User Guide__

    ---

    Learn how to use all K0rdentd features

    [:octicons-arrow-right-24: CLI Reference](user-guide/cli-reference.md)

-   :airplane:{ .lg .middle } __Airgap Installation__

    ---

    Install K0rdent in offline environments

    [:octicons-arrow-right-24: Airgap Guide](user-guide/airgap-installation.md)

</div>

## Project Status

K0rdentd is under active development. Current version: **v0.2.0**

### Build Flavors

| Flavor | Description | Binary Size |
|--------|-------------|-------------|
| Online | Downloads k0s and k0rdent from internet | ~60 MB |
| Airgap | Embeds k0s binary, uses external bundle | ~300 MB |

## Contributing

We welcome contributions! See our [Contributing Guide](development/contributing.md) for details.

## License

K0rdentd is open source software licensed under the Apache License 2.0.

## Support

- **GitHub Issues**: [belgaied2/k0rdentd/issues](https://github.com/belgaied2/k0rdentd/issues)
- **Documentation**: You're reading it!
- **K0rdent Docs**: [docs.k0rdent.io](https://docs.k0rdent.io)
