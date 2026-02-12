# Airgap Assets

This directory contains embedded assets for air-gapped installations.

## Structure

```
assets/
├── k0s/               # Embedded k0s binaries
│   ├── k0s-v1.xx.x-linux-amd64
│   └── k0s-v1.xx.x-linux-arm64
├── skopeo/            # Embedded skopeo binaries
│   ├── skopeo-v1.xx.x-linux-amd64
│   └── skopeo-v1.xx.x-linux-arm64
├── helm/              # Embedded helm charts
│   └── k0rdent-*.tgz
├── images/            # Embedded OCI image bundles
│   ├── k0s-bundle-*.tar
│   └── k0rdent-bundle-*.tar
└── metadata.json      # Build metadata
```

## Embedding

Assets are embedded using Go's `//go:embed` directive in `assets.go`:

```go
//go:embed k0s/*
var K0sBinary embed.FS

//go:embed skopeo/*
var SkopeoBinary embed.FS

//go:embed helm/*
var HelmCharts embed.FS

//go:embed images/*
var ImageBundles embed.FS
```

## Asset Sources

- **k0s binaries**: Downloaded from https://github.com/k0sproject/k0s/releases
- **skopeo binaries**: Downloaded from https://github.com/lework/skopeo-binary/releases
- **k0rdent helm chart**: Pulled from OCI registry (ghcr.io/k0rdent/charts/kcm)
- **Image bundles**: Downloaded from k0rdent enterprise or created via scripts

## Build Process

Assets are populated during the `make build-airgap` target:

1. Download k0s binaries
2. Download skopeo binaries
3. Pull k0rdent helm chart
4. Download/convert k0rdent enterprise bundle
5. Create k0s image bundle
6. Generate metadata.json
7. Build with `-tags airgap`

## Platform Support

Phase 1: linux/amd64 only
Phase 3: Multi-platform (amd64, arm64)
