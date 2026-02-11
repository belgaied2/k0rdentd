// Package bundle handles operations on the k0rdent airgap bundle
package bundle

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractK0rdentVersion extracts the k0rdent version from the bundle
// It searches for the k0rdent-enterprise helm chart and parses Chart.yaml
func ExtractK0rdentVersion(bundlePath string) (string, error) {
	// Check if it's a directory
	if info, err := os.Stat(bundlePath); err == nil && info.IsDir() {
		return extractVersionFromDir(bundlePath)
	}

	// Otherwise treat as tar.gz archive
	return extractVersionFromArchive(bundlePath)
}

// extractVersionFromArchive extracts version from a tar.gz archive
func extractVersionFromArchive(bundlePath string) (string, error) {
	f, err := os.Open(bundlePath)
	if err != nil {
		return "", fmt.Errorf("failed to open bundle: %w", err)
	}
	defer f.Close()

	// Create gzip reader
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gz.Close()

	// Create tar reader
	tr := tar.NewReader(gz)

	// Iterate through tar entries
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Look for Chart.yaml in k0rdent-enterprise chart
		if strings.Contains(header.Name, "k0rdent-enterprise") && strings.HasSuffix(header.Name, "Chart.yaml") {
			// Read Chart.yaml content
			content, err := io.ReadAll(tr)
			if err != nil {
				return "", fmt.Errorf("failed to read Chart.yaml: %w", err)
			}

			// Parse version field
			return parseVersionFromChartYaml(content)
		}
	}

	return "", fmt.Errorf("k0rdent Chart.yaml not found in bundle")
}

// extractVersionFromDir extracts version from an extracted directory
func extractVersionFromDir(bundleDir string) (string, error) {
	// Find k0rdent-enterprise chart directory
	chartsDir := filepath.Join(bundleDir, "charts")
	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read charts directory: %w", err)
	}

	// Look for k0rdent-enterprise chart
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "k0rdent-enterprise") {
			chartYamlPath := filepath.Join(chartsDir, entry.Name(), "Chart.yaml")
			content, err := os.ReadFile(chartYamlPath)
			if err != nil {
				return "", fmt.Errorf("failed to read Chart.yaml: %w", err)
			}

			return parseVersionFromChartYaml(content)
		}

		// Also check for .tar files (compressed charts)
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "k0rdent-enterprise") && strings.HasSuffix(entry.Name(), ".tar") {
			// Extract Chart.yaml from the tar file
			version, err := extractVersionFromTarFile(filepath.Join(chartsDir, entry.Name()))
			if err != nil {
				continue // Try next file
			}
			return version, nil
		}
	}

	return "", fmt.Errorf("k0rdent Chart.yaml not found in bundle directory")
}

// extractVersionFromTarFile extracts version from a single tar file
func extractVersionFromTarFile(tarPath string) (string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	tr := tar.NewReader(f)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if filepath.Base(header.Name) == "Chart.yaml" {
			content, err := io.ReadAll(tr)
			if err != nil {
				return "", err
			}
			return parseVersionFromChartYaml(content)
		}
	}

	return "", fmt.Errorf("Chart.yaml not found in tar file")
}

// parseVersionFromChartYaml parses the version field from Chart.yaml content
func parseVersionFromChartYaml(content []byte) (string, error) {
	// Simple line-based parsing for version field
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "version:") {
			version := strings.TrimSpace(strings.TrimPrefix(trimmed, "version:"))
			// Remove quotes if present
			version = strings.Trim(version, `"`)
			version = strings.Trim(version, `'`)
			if version == "" {
				return "", fmt.Errorf("version field is empty")
			}
			return version, nil
		}
	}

	return "", fmt.Errorf("version field not found in Chart.yaml")
}
