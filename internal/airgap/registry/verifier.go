// Package registry provides OCI registry functionality for airgap installations
package registry

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VerifyBundle verifies the bundle signature using cosign
func VerifyBundle(bundlePath, cosignKey string) error {
	// Check if cosign is installed
	if _, err := exec.LookPath("cosign"); err != nil {
		return fmt.Errorf("cosign not found in PATH: %w", err)
	}

	// Check if signature file exists
	sigPath := bundlePath + ".sig"
	if _, err := os.Stat(sigPath); os.IsNotExist(err) {
		return fmt.Errorf("signature file not found: %s", sigPath)
	}

	// Verify the signature
	// cosign verify-blob --key <cosignKey> --signature <bundle>.sig <bundle>
	var cmd *exec.Cmd
	if strings.HasPrefix(cosignKey, "http://") || strings.HasPrefix(cosignKey, "https://") {
		// Download key from URL
		cmd = exec.Command("cosign", "verify-blob",
			"--key", cosignKey,
			"--signature", sigPath,
			bundlePath)
	} else {
		// Use local key file
		cmd = exec.Command("cosign", "verify-blob",
			"--key", cosignKey,
			"--signature", sigPath,
			bundlePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cosign verification failed: %w, output: %s", err, string(output))
	}

	return nil
}

// DownloadCosignKey downloads the cosign public key from a URL
func DownloadCosignKey(keyURL, destDir string) (string, error) {
	// Check if curl or wget is available
	var cmd *exec.Cmd
	tempFile := filepath.Join(destDir, "cosign.pub")

	if _, err := exec.LookPath("wget"); err == nil {
		cmd = exec.Command("wget", "-O", tempFile, keyURL)
	} else if _, err := exec.LookPath("curl"); err == nil {
		cmd = exec.Command("curl", "-o", tempFile, keyURL)
	} else {
		return "", fmt.Errorf("neither wget nor curl found in PATH")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to download cosign key: %w, output: %s", err, string(output))
	}

	return tempFile, nil
}
