# Release Process

This document describes the release process for K0rdentd.

## Overview

Releases are automated via GitHub Actions. When a tag is pushed, the workflow:

1. Builds binaries for AMD64 and ARM64
2. Creates online and airgap flavors
3. Generates release notes
4. Uploads artifacts to GitHub Releases

## Release Types

| Type | Format | Example |
|------|--------|---------|
| Stable | `vX.Y.Z` | `v0.2.0` |
| Pre-release | `vX.Y.Z-rc.N` | `v0.2.0-rc.1` |
| Beta | `vX.Y.Z-beta.N` | `v0.2.0-beta.1` |

## Release Steps

### 1. Prepare Release

#### Update Version

Update version in code if needed:

```go
// pkg/cli/version.go
var Version = "dev"
```

Version is injected at build time via ldflags.

#### Update Changelog

Create or update `CHANGELOG.md`:

```markdown
## v0.2.0 (2024-01-15)

### New Features
- Add airgap installation support
- Add cloud provider credentials

### Bug Fixes
- Fix K0rdent install check
- Fix OCI registry image tags

### Improvements
- Improve error messages
- Add more logging
```

#### Update Documentation

- Update README.md if needed
- Update docs for new features
- Update installation guide

### 2. Create Tag

```bash
# Ensure you're on main
git checkout main
git pull origin main

# Create tag
git tag v0.2.0

# Push tag
git push origin v0.2.0
```

### 3. Monitor Workflow

1. Go to [GitHub Actions](https://github.com/belgaied2/k0rdentd/actions)
2. Find the "Release" workflow
3. Monitor progress

### 4. Verify Release

Once complete:

1. Go to [Releases](https://github.com/belgaied2/k0rdentd/releases)
2. Verify all artifacts are present:
   - `k0rdentd-v0.2.0-linux-amd64.tar.gz`
   - `k0rdentd-v0.2.0-linux-arm64.tar.gz`
   - `k0rdentd-airgap-v0.2.0-linux-amd64.tar.gz`
   - `k0rdentd-airgap-v0.2.0-linux-arm64.tar.gz`
   - SHA256 checksums
3. Verify release notes are correct

### 5. Post-Release

#### Announce

- Update GitHub release description
- Announce on relevant channels

#### Update Documentation Site

```bash
# Build and deploy docs
mkdocs build
mkdocs deploy
```

## Build Artifacts

### Online Flavor

| Architecture | Binary | Archive |
|-------------|--------|---------|
| AMD64 | `k0rdentd` | `k0rdentd-v0.2.0-linux-amd64.tar.gz` |
| ARM64 | `k0rdentd` | `k0rdentd-v0.2.0-linux-arm64.tar.gz` |

### Airgap Flavor

| Architecture | Binary | Archive |
|-------------|--------|---------|
| AMD64 | `k0rdentd-airgap` | `k0rdentd-airgap-v0.2.0-linux-amd64.tar.gz` |
| ARM64 | `k0rdentd-airgap` | `k0rdentd-airgap-v0.2.0-linux-arm64.tar.gz` |

### Archive Contents

Each archive contains:

- Binary
- `README.md`
- `LICENSE`

## Rollback

If a release has critical issues:

### 1. Mark as Pre-release

```bash
# Via GitHub UI or API
# Edit release → Mark as pre-release
```

### 2. Delete Tag (if needed)

```bash
# Delete local tag
git tag -d v0.2.0

# Delete remote tag
git push origin :refs/tags/v0.2.0
```

### 3. Delete Release

Via GitHub UI:

1. Go to Releases
2. Find the release
3. Click Delete

### 4. Fix and Re-release

1. Fix the issue
2. Create new tag (e.g., `v0.2.1`)
3. Push tag

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Examples

| Version | Change Type |
|---------|-------------|
| `v1.0.0` → `v2.0.0` | Breaking changes |
| `v1.0.0` → `v1.1.0` | New features |
| `v1.0.0` → `v1.0.1` | Bug fixes |

## Release Notes Template

```markdown
## K0rdentd v0.2.0

### New Features

- Feature 1
- Feature 2

### Bug Fixes

- Fix 1
- Fix 2

### Improvements

- Improvement 1
- Improvement 2

### Breaking Changes

- Breaking change 1 (if any)

### Installation

Installation instructions here.

### Airgap Installation

Download the airgap build for offline installations.

### Changelog

See [CHANGELOG.md](CHANGELOG.md) for full details.
```

## Manual Release

If GitHub Actions is unavailable:

```bash
# Build all flavors
make build
make build-airgap

# Create archives
tar -czf k0rdentd-v0.2.0-linux-amd64.tar.gz k0rdentd README.md LICENSE
tar -czf k0rdentd-airgap-v0.2.0-linux-amd64.tar.gz k0rdentd-airgap README.md LICENSE

# Create checksums
sha256sum *.tar.gz > checksums.sha256

# Upload manually via GitHub UI
```

## References

- [Semantic Versioning](https://semver.org/)
- [GitHub Releases API](https://docs.github.com/en/rest/releases/releases)
