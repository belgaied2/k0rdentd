# Architecture: OCI Registry Image Push Fix

## Problem Summary

Three interconnected issues have been identified in the airgap registry implementation:

### Issue 1: Incorrect Tag Format (CRITICAL)
**Current Behavior:**
- Bundle file: `charts/k0rdent-enterprise_1.2.2.tar`
- Converted to: `charts/k0rdent-enterprise_1.2.2` (no explicit tag)
- Stored as: `charts/k0rdent-enterprise_1.2.2:latest`

**Expected Behavior:**
- Bundle file: `charts/k0rdent-enterprise_1.2.2.tar`
- Should be: `charts/k0rdent-enterprise:1.2.2`
- Stored as: `charts/k0rdent-enterprise:1.2.2`

**Root Cause:** The `pathToImageRef()` function doesn't convert the underscore (`_`) version separator to OCI tag format (`:`).

### Issue 2: Incorrect Path Structure (CRITICAL)
**Current Behavior:**
- Images extracted to temp dir: `/tmp/k0rdent-bundle-xxx/extracted/charts/k0rdent-enterprise_1.2.2.tar`
- Pushed to: `localhost:5000/extracted/charts/k0rdent-enterprise_1.2.2:latest`

**Expected Behavior:**
- Should be pushed to: `localhost:5000/charts/k0rdent-enterprise:1.2.2`

**Root Cause:** The `pathToImageRef()` function uses `filepath.Base(filepath.Dir(path))` which captures the temp directory name (`extracted`) instead of the logical bundle structure.

### Issue 3: External Dependency on Skopeo (NICE-TO-HAVE)
- Current implementation requires `skopeo` binary to be available
- User prefers native Go implementation using `go-containerregistry`

---

## Proposed Solution

### Option A: Fix pathToImageRef() + Keep Skopeo (RECOMMENDED - Short Term)

**Rationale:**
- Minimal code changes
- Proven skopeo behavior
- Can be implemented quickly

**Changes Required:**

```go
// pathToImageRef converts a file path to an OCI image reference
// It handles:
// 1. Converting underscore version separator to OCI tag format
// 2. Mapping bundle directory structure to registry paths (ignoring temp dirs)
// 3. Preserving images that already have proper tag format
func pathToImageRef(imgPath string, bundleRoot string) string {
    // Get relative path from bundle root to maintain logical structure
    relPath, _ := filepath.Rel(bundleRoot, imgPath)
    
    // Remove .tar suffix
    ref := strings.TrimSuffix(relPath, ".tar")
    
    // Split into directory and filename components
    dir := filepath.Dir(relPath)
    if dir == "." {
        dir = ""
    }
    filename := filepath.Base(relPath)
    filename = strings.TrimSuffix(filename, ".tar")
    
    // Parse version from filename
    // Handle formats like: "k0rdent-enterprise_1.2.2" or "k0s:v1.32.8-k0s.0"
    var imageName, tag string
    
    if idx := strings.LastIndex(filename, "_"); idx != -1 {
        // Format: name_version (e.g., k0rdent-enterprise_1.2.2)
        imageName = filename[:idx]
        tag = filename[idx+1:]
    } else if idx := strings.LastIndex(filename, ":"); idx != -1 {
        // Format: name:tag (e.g., k0s:v1.32.8-k0s.0)
        imageName = filename[:idx]
        tag = filename[idx+1:]
    } else {
        // No version found, use filename as image name and "latest" as tag
        imageName = filename
        tag = "latest"
    }
    
    // Build final reference: dir/imageName:tag
    if dir != "" && dir != "." {
        return fmt.Sprintf("%s/%s:%s", dir, imageName, tag)
    }
    return fmt.Sprintf("%s:%s", imageName, tag)
}
```

**Update pushSingleImage to pass bundleRoot:**

```go
func pushSingleImage(imgPath, registryAddr string, bundleRoot string) error {
    imgRef := pathToImageRef(imgPath, bundleRoot)
    dest := fmt.Sprintf("docker://%s/%s", registryAddr, imgRef)
    
    // Use skopeo to copy the image
    cmd := exec.Command("skopeo", "copy", "--insecure-policy", "--dest-tls-verify=false", 
        "oci-archive:"+imgPath, dest)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("%w, output: %s", err, string(output))
    }
    return nil
}
```

---

### Option B: Use go-containerregistry (RECOMMENDED - Long Term)

**Rationale:**
- No external binary dependency
- More control over the push process
- Better error handling
- Already using go-containerregistry for the registry server

**Implementation Approach:**

```go
import (
    "github.com/google/go-containerregistry/pkg/name"
    "github.com/google/go-containerregistry/pkg/v1/tarball"
    "github.com/google/go-containerregistry/pkg/v1/remote"
)

// pushSingleImage pushes a single image using go-containerregistry
func pushSingleImage(imgPath, registryAddr string, bundleRoot string) error {
    // Parse the image reference
    imgRef := pathToImageRef(imgPath, bundleRoot)
    
    // Parse as OCI reference
    ref, err := name.ParseReference(imgRef, name.WithDefaultRegistry(registryAddr))
    if err != nil {
        return fmt.Errorf("failed to parse reference %s: %w", imgRef, err)
    }
    
    // Read image from OCI archive
    // Note: tarball.ImageFromPath expects a docker-style tarball
    // For OCI archives, we need to use a different approach
    img, err := readOCIArchive(imgPath)
    if err != nil {
        return fmt.Errorf("failed to read OCI archive: %w", err)
    }
    
    // Push to registry
    if err := remote.Write(ref, img, remote.WithInsecure()); err != nil {
        return fmt.Errorf("failed to push image: %w", err)
    }
    
    return nil
}
```

**Challenge with go-containerregistry:**

The `go-containerregistry` library's `tarball` package is designed for Docker-style tarballs (created by `docker save`), not OCI archives. OCI archives have a different internal structure:

```
OCI Archive Structure:
├── oci-layout
├── index.json
├── blobs/
│   └── sha256/
│       └── ...

Docker Tarball Structure:
├── manifest.json
├── ...layer tarballs...
```

**Solution for OCI Archives:**

Option B1: Use `github.com/google/go-containerregistry/pkg/v1/layout` to read OCI layout:

```go
import (
    "github.com/google/go-containerregistry/pkg/v1/layout"
)

func readOCIArchive(archivePath string) (v1.Image, error) {
    // Extract OCI archive to temp directory
    tempDir, err := os.MkdirTemp("", "oci-extract-*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tempDir)
    
    // Extract tar
    if err := extractTar(archivePath, tempDir); err != nil {
        return nil, err
    }
    
    // Read OCI layout
    p, err := layout.FromPath(tempDir)
    if err != nil {
        return nil, err
    }
    
    // Get the image from the layout
    // This requires knowing the digest or tag
    // ...
}
```

Option B2: Continue using skopeo for now, migrate later (HYBRID APPROACH)

---

## Recommended Implementation Plan

### Phase 1: Fix pathToImageRef (Immediate Priority)

**Goal:** Fix both the tag format and path structure issues

**Changes:**
1. Modify `pathToImageRef()` to accept `bundleRoot` parameter
2. Use relative path from bundle root instead of parent directory
3. Parse underscore version separator and convert to OCI tag format
4. Update `pushSingleImage()` to pass `bundleRoot`
5. Update `pushImagesWithProgress()` to pass `bundleRoot`

**Testing:**
```bash
# After fix, verify:
curl -s http://localhost:5000/v2/charts/k0rdent-enterprise/tags/list
# Expected: {"name":"charts/k0rdent-enterprise","tags":["1.2.2"]}

curl -s http://localhost:5000/v2/k0sproject/k0s/tags/list  
# Expected: {"name":"k0sproject/k0s","tags":["v1.32.8-k0s.0"]}
```

### Phase 2: Evaluate go-containerregistry Migration

**Research Tasks:**
1. Investigate if go-containerregistry has native OCI archive support
2. Compare performance: skopeo vs native Go
3. Assess complexity of implementing OCI archive reader

**Decision Criteria:**
- If OCI archive support is straightforward: migrate to go-containerregistry
- If complex: keep skopeo, consider embedding skopeo binary

---

## Bundle Directory Mapping

### Current (Broken) Mapping

The bundle extracts to a temp directory (e.g., `/tmp/k0rdent-bundle-xxx/extracted/`). The current `pathToImageRef()` function uses `filepath.Base(filepath.Dir(path))` which causes different behavior depending on image location:

| Bundle File (in extracted dir) | Current Registry Path | Issues |
|-------------------------------|----------------------|--------|
| `extracted/charts/k0rdent-enterprise_1.2.2.tar` | `charts/k0rdent-enterprise_1.2.2:latest` | ✅ Path correct, ❌ Tag wrong |
| `extracted/k0sproject/k0s:v1.32.8-k0s.0.tar` | `k0sproject/k0s:v1.32.8-k0s.0:latest` | ✅ Path correct, ❌ Tag wrong |
| `extracted/capi/cluster-api-controller_v1.11.2.tar` | `capi/cluster-api-controller_v1.11.2:latest` | ✅ Path correct, ❌ Tag wrong |
| `extracted/kcm-controller_1.2.2.tar` (root level) | `extracted/kcm-controller_1.2.2:latest` | ❌ Path wrong (includes 'extracted'), ❌ Tag wrong |
| `extracted/k0rdent-operator_1.2.2.tar` (root level) | `extracted/k0rdent-operator_1.2.2:latest` | ❌ Path wrong (includes 'extracted'), ❌ Tag wrong |

**Key Insight:**
- Images in subdirectories (charts/, k0sproject/, capi/) get the correct path because `filepath.Base(filepath.Dir(path))` returns the subdirectory name
- Images at the root of `extracted/` get `extracted` as their "directory", which becomes part of the image name

### Fixed Mapping

| Bundle File (in extracted dir) | Registry Path | Fix Applied |
|--------------------------------|---------------|-------------|
| `extracted/charts/k0rdent-enterprise_1.2.2.tar` | `charts/k0rdent-enterprise:1.2.2` | Underscore → colon |
| `extracted/k0sproject/k0s:v1.32.8-k0s.0.tar` | `k0sproject/k0s:v1.32.8-k0s.0` | Remove :latest suffix |
| `extracted/capi/cluster-api-controller_v1.11.2.tar` | `capi/cluster-api-controller:v1.11.2` | Underscore → colon |
| `extracted/kcm-controller_1.2.2.tar` (root level) | `kcm-controller:1.2.2` | Strip 'extracted', underscore → colon |
| `extracted/k0rdent-operator_1.2.2.tar` (root level) | `k0rdent-operator:1.2.2` | Strip 'extracted', underscore → colon |
| `extracted/provider-aws/aws-controller_v2.10.0.tar` | `provider-aws/aws-controller:v2.10.0` | Underscore → colon |

---

## Code Changes Required

### File: `internal/airgap/registry/pusher.go`

```go
// PushImages pushes all images from the bundle to the local registry
func PushImages(bundlePath, registryAddr string) error {
    logger := getLogger()

    // Extract bundle if tar.gz
    bundleDir := bundlePath
    if strings.HasSuffix(bundlePath, ".tar.gz") || strings.HasSuffix(bundlePath, ".tgz") {
        tmpDir, err := os.MkdirTemp("", "k0rdent-bundle-*")
        if err != nil {
            return fmt.Errorf("failed to create temp dir: %w", err)
        }
        defer os.RemoveAll(tmpDir)

        logger.Infof("Extracting bundle to %s", tmpDir)
        if err := extractTarGz(bundlePath, tmpDir); err != nil {
            return err
        }
        bundleDir = tmpDir
    }

    // Find all OCI archives in bundle
    images, err := findOCIArchives(bundleDir)
    if err != nil {
        return fmt.Errorf("failed to find OCI archives: %w", err)
    }

    logger.Infof("Found %d images to push", len(images))

    // Check if skopeo is available
    if _, err := exec.LookPath("skopeo"); err != nil {
        return fmt.Errorf("skopeo not found in PATH: %w", err)
    }

    // Push images with progress reporting - PASS bundleDir as root
    if err := pushImagesWithProgress(images, bundleDir, registryAddr); err != nil {
        return err
    }

    logger.Infof("✓ All images pushed to registry")
    return nil
}

// pushImagesWithProgress pushes images and reports progress
func pushImagesWithProgress(images []string, bundleRoot, registryAddr string) error {
    // ... existing code ...
    
    for i, imgPath := range images {
        wg.Add(1)
        go func(index int, path string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            relPath, _ := filepath.Rel(bundleRoot, path)
            logger.Infof("[%d/%d] Pushing: %s", index+1, total, relPath)

            // PASS bundleRoot to pushSingleImage
            if err := pushSingleImage(path, registryAddr, bundleRoot); err != nil {
                logger.Warnf("Failed to push %s: %v", relPath, err)
                errors <- err
            }
        }(i, imgPath)
    }
    // ... rest of function ...
}

// pushSingleImage pushes a single image to the registry
func pushSingleImage(imgPath, registryAddr string, bundleRoot string) error {
    // Build image reference from path using bundleRoot
    imgRef := pathToImageRef(imgPath, bundleRoot)
    dest := fmt.Sprintf("docker://%s/%s", registryAddr, imgRef)

    cmd := exec.Command("skopeo", "copy", "--insecure-policy", "--dest-tls-verify=false", 
        "oci-archive:"+imgPath, dest)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("%w, output: %s", err, string(output))
    }
    return nil
}

// pathToImageRef converts a file path to an OCI image reference
// bundleRoot is the root directory of the extracted bundle
func pathToImageRef(imgPath string, bundleRoot string) string {
    // Get relative path from bundle root to maintain logical structure
    relPath, _ := filepath.Rel(bundleRoot, imgPath)
    
    // Remove .tar suffix
    ref := strings.TrimSuffix(relPath, ".tar")
    
    // Split into directory and filename components
    dir := filepath.Dir(relPath)
    if dir == "." {
        dir = ""
    }
    filename := filepath.Base(ref)
    
    // Parse version from filename
    var imageName, tag string
    
    if idx := strings.LastIndex(filename, "_"); idx != -1 {
        // Format: name_version (e.g., k0rdent-enterprise_1.2.2)
        imageName = filename[:idx]
        tag = filename[idx+1:]
    } else if idx := strings.LastIndex(filename, ":"); idx != -1 {
        // Format: name:tag (e.g., k0s:v1.32.8-k0s.0)
        imageName = filename[:idx]
        tag = filename[idx+1:]
    } else {
        // No version found, use filename as image name and "latest" as tag
        imageName = filename
        tag = "latest"
    }
    
    // Build final reference: dir/imageName:tag
    if dir != "" && dir != "." {
        return fmt.Sprintf("%s/%s:%s", dir, imageName, tag)
    }
    return fmt.Sprintf("%s:%s", imageName, tag)
}
```

---

## Verification Steps

1. **Start registry daemon:**
   ```bash
   sudo ./k0rdentd registry -d /path/to/bundle.tar.gz -p 5000
   ```

2. **Verify image tags:**
   ```bash
   # Charts should have proper tags
   curl -s http://localhost:5000/v2/charts/k0rdent-enterprise/tags/list
   # Expected: {"name":"charts/k0rdent-enterprise","tags":["1.2.2"]}
   
   # Images should be at root, not in 'extracted' path
   curl -s http://localhost:5000/v2/k0sproject/k0s/tags/list
   # Expected: {"name":"k0sproject/k0s","tags":["v1.32.8-k0s.0"]}
   
   # Should NOT find images at 'extracted' path
   curl -s http://localhost:5000/v2/extracted/charts/k0rdent-enterprise/tags/list
   # Expected: 404 Not Found
   ```

3. **Test k0rdent installation:**
   ```bash
   sudo ./k0rdentd install --airgap
   # Should successfully pull charts from localhost:5000
   ```

---

## References

- [Skopeo Documentation](https://github.com/containers/skopeo/blob/main/docs/skopeo.1.md)
- [OCI Image Layout Spec](https://github.com/opencontainers/image-spec/blob/main/image-layout.md)
- [go-containerregistry](https://github.com/google/go-containerregistry)
- [k0s Helm Extensions](https://github.com/k0sproject/k0s/blob/main/docs/helm-charts.md)
