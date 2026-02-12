// Package registry provides OCI registry functionality for airgap installations
package registry

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/belgaied2/k0rdentd/internal/airgap"
	"github.com/belgaied2/k0rdentd/internal/airgap/assets"
	"github.com/google/go-containerregistry/pkg/registry"
)

// RegistryDaemon runs a local OCI registry for airgap installations
type RegistryDaemon struct {
	port       string
	host       string
	storageDir string
	bundlePath string
	server     *http.Server
	verifySig  bool
	cosignKey  string
}

// NewRegistryDaemon creates a new registry daemon instance
func NewRegistryDaemon(port, storageDir, bundlePath string, verifySig bool, cosignKey string) *RegistryDaemon {
	return &RegistryDaemon{
		port:       port,
		host:       "0.0.0.0", // Listen on all interfaces
		storageDir: storageDir,
		bundlePath: bundlePath,
		verifySig:  verifySig,
		cosignKey:  cosignKey,
	}
}

// Start starts the registry daemon
func (r *RegistryDaemon) Start(ctx context.Context) error {
	logger := getLogger()

	logger.Infof("Starting k0rdentd registry daemon...")

	// Step 1: Extract skopeo binary if this is an airgap build
	if airgap.IsAirGap() {
		logger.Infof("Extracting skopeo binary from embedded assets...")
		if err := ExtractSkopeoBinary(); err != nil {
			return fmt.Errorf("failed to extract skopeo binary: %w", err)
		}
		logger.Infof("✓ Skopeo binary extracted to /usr/bin/skopeo")
	}

	// Step 2: Verify bundle with cosign if requested
	if r.verifySig {
		logger.Infof("Verifying bundle signature...")
		if err := r.verifyBundle(); err != nil {
			return fmt.Errorf("bundle verification failed: %w", err)
		}
		logger.Infof("✓ Bundle verified with cosign")
	}

	// Step 3: Extract k0rdent version from bundle
	// logger.Infof("Extracting k0rdent version from bundle...")
	// k0rdentVersion, err := bundle.ExtractK0rdentVersion(r.bundlePath)
	// if err != nil {
	// 	return fmt.Errorf("failed to extract version: %w", err)
	// }
	// logger.Infof("✓ K0rdent version from bundle: %s", k0rdentVersion)

	// Step 4: Initialize disk-based registry
	if err := os.MkdirAll(r.storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage dir: %w", err)
	}

	logger.Infof("Initializing registry with disk-based storage: %s", r.storageDir)
	// Create disk-based blob handler
	blobHandler := registry.NewDiskBlobHandler(r.storageDir)
	// Create registry handler
	reg := registry.New(registry.WithBlobHandler(blobHandler))

	// Step 5: Start HTTP server
	r.server = &http.Server{
		Addr:         r.host + ":" + r.port,
		Handler:      reg,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Infof("Registry server listening on %s", r.Addr())
		logger.Infof("Storage directory: %s", r.storageDir)
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Step 6: Push images from bundle to local registry
	logger.Infof("Pushing images from bundle to local registry...")
	if err := r.pushImagesToRegistry(ctx); err != nil {
		return fmt.Errorf("failed to push images: %w", err)
	}

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		logger.Infof("Shutting down registry server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := r.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("registry shutdown failed: %w", err)
		}
		logger.Infof("✓ Registry server stopped gracefully")
		return nil
	case err := <-errChan:
		return err
	}
}

// verifyBundle verifies the bundle signature
func (r *RegistryDaemon) verifyBundle() error {
	// If cosignKey is a URL, download it first
	cosignKey := r.cosignKey
	if strings.HasPrefix(cosignKey, "http://") || strings.HasPrefix(cosignKey, "https://") {
		logger := getLogger()
		logger.Infof("Downloading cosign key from %s", cosignKey)

		tempDir, err := os.MkdirTemp("", "cosign-key-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir for cosign key: %w", err)
		}
		defer os.RemoveAll(tempDir)

		downloadedKey, err := DownloadCosignKey(cosignKey, tempDir)
		if err != nil {
			return err
		}
		cosignKey = downloadedKey
		defer os.Remove(downloadedKey)
	}

	return VerifyBundle(r.bundlePath, cosignKey)
}

// pushImagesToRegistry pushes images from the bundle to the local registry
func (r *RegistryDaemon) pushImagesToRegistry(ctx context.Context) error {
	registryAddr := "localhost:" + r.port
	return PushImages(r.bundlePath, registryAddr)
}

// Addr returns the registry address
func (r *RegistryDaemon) Addr() string {
	return r.host + ":" + r.port
}

// WaitForSignal waits for shutdown signals
func (r *RegistryDaemon) WaitForSignal(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigChan:
		// Signal received, context will be cancelled
	case <-ctx.Done():
		// Context already cancelled
	}
}

// IsRunning checks if the registry port is in use
func (r *RegistryDaemon) IsRunning() bool {
	ln, err := net.Listen("tcp", r.host+":"+r.port)
	if err != nil {
		return true // Port is in use
	}
	ln.Close()
	return false
}

// GetStorageSize returns the total size of the registry storage
func (r *RegistryDaemon) GetStorageSize() (int64, error) {
	var size int64

	err := filepath.Walk(r.storageDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ExtractSkopeoBinary extracts the embedded skopeo binary to /usr/bin/skopeo
// This is only available in airgap builds
func ExtractSkopeoBinary() error {
	if !airgap.IsAirGap() {
		return fmt.Errorf("not an airgap build, cannot extract embedded skopeo binary")
	}

	// Read the skopeo directory from embedded FS
	entries, err := fs.ReadDir(assets.SkopeoBinary, "skopeo")
	if err != nil {
		return fmt.Errorf("failed to read embedded skopeo directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no skopeo binary found in embedded assets")
	}

	// Find the skopeo binary (should be only one file)
	var skopeoBinaryName string
	for _, entry := range entries {
		if !entry.IsDir() {
			skopeoBinaryName = entry.Name()
			break
		}
	}

	if skopeoBinaryName == "" {
		return fmt.Errorf("no skopeo binary file found in embedded assets")
	}

	// Open the embedded skopeo binary
	srcPath := filepath.Join("skopeo", skopeoBinaryName)
	srcFile, err := assets.SkopeoBinary.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open embedded skopeo binary: %w", err)
	}
	defer srcFile.Close()

	// Ensure /usr/bin exists
	if err := os.MkdirAll("/usr/bin", 0755); err != nil {
		return fmt.Errorf("failed to create /usr/bin directory: %w", err)
	}

	// Create destination file at /usr/bin/skopeo
	dstPath := "/usr/bin/skopeo"
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create skopeo binary at %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy the binary
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy skopeo binary: %w", err)
	}

	return nil
}
