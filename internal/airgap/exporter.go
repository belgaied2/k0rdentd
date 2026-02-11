package airgap

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/belgaied2/k0rdentd/internal/airgap/assets"
)

// Exporter handles export of embedded assets for worker nodes
// Option B: Only k0s binary is embedded; k0rdent bundle is external
type Exporter struct {
	// bundlePath is the path to the k0rdent airgap bundle
	// This will be referenced in the worker bundle (not copied)
	bundlePath string
}

// NewExporter creates a new exporter instance
func NewExporter(bundlePath string) *Exporter {
	return &Exporter{
		bundlePath: bundlePath,
	}
}

// ExtractK0sBinary extracts the embedded k0s binary to the output directory
func (e *Exporter) ExtractK0sBinary(outputDir string) error {
	fmt.Printf("Extracting k0s binary to: %s\n", outputDir)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// In airgap builds, extract from embedded assets
	if IsAirGap() {
		return e.extractFromEmbedded(outputDir)
	}

	// Fallback: check file system (for development/testing)
	k0sSrcDir := filepath.Join(".", "internal", "airgap", "assets", "k0s")

	// Check if k0s binaries exist
	entries, err := os.ReadDir(k0sSrcDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No k0s binaries available
			fmt.Println("  No k0s binaries found")
			return fmt.Errorf("k0s binaries not found in embedded assets or filesystem")
		}
		return fmt.Errorf("failed to read k0s directory: %w", err)
	}

	// Copy each k0s binary
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(k0sSrcDir, entry.Name())
		dstPath := filepath.Join(outputDir, entry.Name())

		if err := copyFile(srcPath, dstPath, 0755); err != nil {
			return fmt.Errorf("failed to copy k0s binary: %w", err)
		}

		fmt.Printf("  ✓ Copied: %s\n", entry.Name())
	}

	return nil
}

// extractFromEmbedded extracts k0s binary from embedded assets
func (e *Exporter) extractFromEmbedded(outputDir string) error {
	fmt.Println("  Extracting from embedded assets...")

	// Read the k0s directory from embedded FS
	entries, err := fs.ReadDir(assets.K0sBinary, "k0s")
	if err != nil {
		return fmt.Errorf("failed to read embedded k0s directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no files found in embedded k0s directory")
	}

	// Copy each k0s binary from embed FS to output
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Open the embedded file
		srcPath := filepath.Join("k0s", entry.Name())
		srcFile, err := assets.K0sBinary.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open embedded file %s: %w", entry.Name(), err)
		}
		defer srcFile.Close()

		// Create destination file
		dstPath := filepath.Join(outputDir, entry.Name())
		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", entry.Name(), err)
		}
		defer dstFile.Close()

		// Copy content
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}

		fmt.Printf("  ✓ Extracted: %s\n", entry.Name())
	}

	return nil
}

// ExtractImageBundles creates a reference to the external bundle
// For Option B, we don't copy images - we reference the bundle location
func (e *Exporter) ExtractImageBundles(outputDir string) error {
	fmt.Printf("Creating bundle reference for: %s\n", e.bundlePath)

	// Create a marker file with bundle path
	markerPath := filepath.Join(outputDir, "BUNDLE_LOCATION.txt")
	content := fmt.Sprintf("# k0rdent Airgap Bundle Location\n\n")
	content += fmt.Sprintf("BUNDLE_PATH=%s\n\n", e.bundlePath)
	content += fmt.Sprintf("# This is the k0rdent enterprise airgap bundle.\n")
	content += fmt.Sprintf("# Download from: https://get.mirantis.com/k0rdent-enterprise/\n")
	content += fmt.Sprintf("#\n")
	content += fmt.Sprintf("# On the worker node, extract images from this bundle:\n")
	content += fmt.Sprintf("#   1. Copy the bundle to this worker node\n")
	content += fmt.Sprintf("#   2. Extract: tar -xf %s -C /var/lib/k0s/images/\n", filepath.Base(e.bundlePath))
	content += fmt.Sprintf("#      (if bundle is tar.gz, first: tar -xzf bundle.tar.gz)\n")

	if err := os.WriteFile(markerPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write bundle marker: %w", err)
	}

	fmt.Printf("  ✓ Created bundle reference: BUNDLE_LOCATION.txt\n")
	return nil
}

// GenerateScripts creates helper scripts for worker setup
func (e *Exporter) GenerateScripts(outputDir string) error {
	fmt.Printf("Generating helper scripts in: %s\n", outputDir)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate install.sh script that references the external bundle
	installScript := fmt.Sprintf(`#!/bin/bash
# k0rdentd worker installation script
set -e

BUNDLE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
K0RDENT_BUNDLE="%s"

echo "=== K0rdentd Worker Installation ==="
echo ""

# Step 1: Install k0s binary
echo "[1/3] Installing k0s from bundle..."
if [ -d "${BUNDLE_DIR}/k0s-binary" ]; then
    K0S_BIN=$(find "${BUNDLE_DIR}/k0s-binary" -type f -name "k0s-*" | head -n 1)
    if [ -n "$K0S_BIN" ]; then
        sudo install -m 755 "$K0S_BIN" /usr/local/bin/k0s
        echo "✓ k0s installed to /usr/local/bin/k0s"
    else
        echo "ERROR: k0s binary not found in ${BUNDLE_DIR}/k0s-binary/"
        exit 1
    fi
else
    echo "ERROR: k0s-binary directory not found"
    exit 1
fi

# Step 2: Check for k0rdent bundle
echo ""
echo "[2/3] Checking for k0rdent airgap bundle..."
if [ ! -f "$K0RDENT_BUNDLE" ]; then
    echo "⚠ WARNING: k0rdent bundle not found at: $K0RDENT_BUNDLE"
    echo ""
    echo "You need to download the k0rdent enterprise airgap bundle:"
    echo "  https://get.mirantis.com/k0rdent-enterprise/"
    echo ""
    echo "After downloading, extract images:"
    echo "  tar -xzf airgap-bundle-*.tar.gz -C /var/lib/k0s/images/"
    echo ""
    echo "For now, continuing without k0rdent images..."
else
    echo "✓ k0rdent bundle found: $K0RDENT_BUNDLE"
    echo ""
    echo "Extracting bundle to /var/lib/k0s/images/..."
    sudo mkdir -p /var/lib/k0s/images/

    # Extract the bundle
    if [[ "$K0RDENT_BUNDLE" == *.tar.gz ]]; then
        # If it's a tar.gz, extract to temp dir first, then copy images
        TEMP_DIR=$(mktemp -d)
        tar -xzf "$K0RDENT_BUNDLE" -C "$TEMP_DIR"

        # Load all OCI archives into k0s containerd
        echo "Loading images into k0s..."
        for img in "$TEMP_DIR"/*.tar "$TEMP_DIR"/*/*.tar 2>/dev/null; do
            if [ -f "$img" ]; then
                echo "  Loading: $(basename "$img")"
                sudo k0s ctr images import "$img" || echo "  Warning: Failed to load $(basename "$img")"
            fi
        done

        rm -rf "$TEMP_DIR"
    elif [ -d "$K0RDENT_BUNDLE" ]; then
        # If it's already extracted, load directly
        echo "Loading images from extracted bundle..."
        for img in "$K0RDENT_BUNDLE"/*.tar "$K0RDENT_BUNDLE"/*/*.tar 2>/dev/null; do
            if [ -f "$img" ]; then
                echo "  Loading: $(basename "$img")"
                sudo k0s ctr images import "$img" || echo "  Warning: Failed to load $(basename "$img")"
            fi
        done
    fi

    echo "✓ Bundle images loaded"
fi

# Step 3: Ready to join
echo ""
echo "[3/3] Installation complete!"
echo ""
echo "Next steps:"
echo "1. Get worker token from controller:"
echo "   sudo k0s token create --role worker"
echo ""
echo "2. Join the cluster:"
echo "   sudo k0s worker <token-from-controller>"
`, e.bundlePath)

	installScriptPath := filepath.Join(outputDir, "install.sh")
	if err := os.WriteFile(installScriptPath, []byte(installScript), 0755); err != nil {
		return fmt.Errorf("failed to write install.sh: %w", err)
	}
	fmt.Printf("  ✓ Created: install.sh\n")

	return nil
}

// GenerateReadme creates a README with instructions for worker setup
func (e *Exporter) GenerateReadme(outputDir string) error {
	fmt.Printf("Generating README in: %s\n", outputDir)

	// Generate README with external bundle reference
	readmeContent := fmt.Sprintf(`# K0rdentd Worker Artifacts

This directory contains the k0s binary and references needed to join a worker node to an air-gapped k0s cluster with k0rdent.

## Contents

- k0s-binary/: k0s binary for this architecture
- BUNDLE_LOCATION.txt: Reference to the k0rdent enterprise airgap bundle
- scripts/: Helper scripts for installation

## Quick Start

1. Copy this directory AND the k0rdent airgap bundle to the worker node:
   scp -r worker-bundle user@worker:/tmp/
   scp airgap-bundle-*.tar.gz user@worker:/tmp/

2. On the worker node, run the install script:
   cd /tmp/worker-bundle/scripts
   sudo ./install.sh

3. Join the cluster:
   sudo k0s worker <token-from-controller>

## K0rdent Airgap Bundle

The k0rdent bundle is NOT included in this worker bundle to avoid redistributing enterprise content.

Bundle location: %s

### Downloading the Bundle

If you don't have the bundle, download it from Mirantis:
  https://get.mirantis.com/k0rdent-enterprise/

Example:
  wget https://get.mirantis.com/k0rdent-enterprise/1.2.3/airgap-bundle-1.2.3.tar.gz

### Extracting Images from Bundle

After downloading, extract images to /var/lib/k0s/images/:

If bundle is a tar.gz archive:
  # First extract the archive
  tar -xzf airgap-bundle-1.2.3.tar.gz -C /tmp/bundle

  # Then load all images into k0s
  cd /tmp/bundle
  for img in *.tar */*.tar; do
    [ -f "$img" ] && sudo k0s ctr images import "$img"
  done

Or extract directly to /var/lib/k0s/images/:
  tar -xzf airgap-bundle-1.2.3.tar.gz -C /var/lib/k0s/images/

## Getting the Worker Token

On the controller node, run:
  sudo k0s token create --role worker

## Manual Installation

1. Install k0s:
   sudo install -m 755 k0s-binary/k0s-* /usr/local/bin/k0s

2. Extract k0rdent bundle images (see above)

3. Join the cluster:
   sudo k0s worker <token>

## Support

- k0s documentation: https://docs.k0sproject.io/
- k0rdent documentation: https://docs.k0rdent.io/
- k0rdent enterprise airgap docs: https://docs.mirantis.com/k0rdent-enterprise/latest/admin/installation/airgap/airgap-bundles/
`, e.bundlePath)

	readmePath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}
	fmt.Printf("  ✓ Created: README.md\n")

	return nil
}

// copyFile copies a file from src to dst with specified permissions
func copyFile(src, dst string, perm os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
