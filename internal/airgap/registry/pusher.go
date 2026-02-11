// Package registry provides OCI registry functionality for airgap installations
package registry

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// PushImages pushes all images from the bundle to the local registry
func PushImages(bundlePath, registryAddr string) error {
	logger := getLogger()

	// Extract bundle if tar.gz
	bundleDir := bundlePath
	if strings.HasSuffix(bundlePath, ".tar.gz") || strings.HasSuffix(bundlePath, ".tgz") {
		// Extract to temp directory
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

	// Push images with progress reporting
	if err := pushImagesWithProgress(images, bundleDir, registryAddr); err != nil {
		return err
	}

	logger.Infof("✓ All images pushed to registry")
	return nil
}

// findOCIArchives finds all .tar files that look like OCI archives
func findOCIArchives(bundleDir string) ([]string, error) {
	var images []string

	err := filepath.Walk(bundleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Look for .tar files
		if strings.HasSuffix(path, ".tar") {
			// filtering out skopeo
			if !strings.Contains(path, "skopeo") {
				images = append(images, path)
			}
		}
		return nil
	})

	return images, err
}

// pushImagesWithProgress pushes images and reports progress
func pushImagesWithProgress(images []string, bundleDir, registryAddr string) error {
	logger := getLogger()
	total := len(images)

	// Use a wait group for concurrent pushes (limited concurrency)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Max 5 concurrent pushes
	errors := make(chan error, total)

	for i, imgPath := range images {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			relPath, _ := filepath.Rel(bundleDir, path)
			logger.Infof("[%d/%d] Pushing: %s", index+1, total, relPath)

			if err := pushSingleImage(path, registryAddr); err != nil {
				logger.Warnf("Failed to push %s: %v", relPath, err)
				errors <- err
			}
		}(i, imgPath)
	}

	// Wait for all pushes to complete
	wg.Wait()
	close(errors)

	// Check if there were any errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		return fmt.Errorf("failed to push %d out of %d images", len(errorList), total)
	}

	return nil
}

// pushSingleImage pushes a single image to the registry
func pushSingleImage(imgPath, registryAddr string) error {
	// Build image reference from path
	// Example: bundle/k0sproject/k0s:v1.32.8-k0s.0.tar -> localhost:5000/k0sproject/k0s:v1.32.8-k0s.0
	imgRef := pathToImageRef(imgPath)
	dest := fmt.Sprintf("docker://%s/%s", registryAddr, imgRef)

	// Use skopeo to copy the image
	cmd := exec.Command("skopeo", "copy", "--dest-tls-verify=false", "oci-archive:"+imgPath, dest) // TODO: Remove the tls-verify=false for production
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w, output: %s", err, string(output))
	}

	return nil
}

// pathToImageRef converts a file path to an OCI image reference
func pathToImageRef(path string) string {
	// Extract directory structure as image name
	// Example: /tmp/bundle/k0sproject/k0s:v1.32.8-k0s.0.tar -> k0sproject/k0s:v1.32.8-k0s.0
	filename := filepath.Base(path)
	// Remove .tar suffix
	ref := strings.TrimSuffix(filename, ".tar")

	// Get parent directory as image namespace
	dir := filepath.Base(filepath.Dir(path))
	if dir != "." && dir != "/" && !strings.HasPrefix(ref, dir) {
		ref = dir + "/" + ref
	}

	return ref
}

// extractTarGz extracts a tar.gz archive to a directory
func extractTarGz(src, dst string) error {
	cmd := exec.Command("tar", "-xzf", src, "-C", dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w, output: %s", err, string(output))
	}
	return nil
}

// logger interface for progress reporting
type logger interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
}

// getLogger returns a logger instance
func getLogger() logger {
	return defaultLogger{}
}

// defaultLogger is a minimal logger implementation
type defaultLogger struct{}

func (l defaultLogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l defaultLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}
