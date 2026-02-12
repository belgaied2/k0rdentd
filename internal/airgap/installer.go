package airgap

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/belgaied2/k0rdentd/internal/airgap/assets"
	"github.com/belgaied2/k0rdentd/internal/airgap/bundle"
	"github.com/belgaied2/k0rdentd/internal/airgap/containerd"
	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/generator"
	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// Installer handles airgap installation
type Installer struct {
	config *config.K0rdentdConfig
	debug  bool
}

// NewInstaller creates a new airgap installer
func NewInstaller(cfg *config.K0rdentdConfig, debug bool) *Installer {
	return &Installer{
		config: cfg,
		debug:  debug,
	}
}

// ExtractK0sBinary extracts the embedded k0s binary to /usr/local/bin/k0s
func (i *Installer) ExtractK0sBinary() error {
	if !IsAirGap() {
		return fmt.Errorf("not an airgap build, cannot extract embedded k0s binary")
	}

	// Read the k0s directory from embedded FS
	entries, err := fs.ReadDir(assets.K0sBinary, "k0s")
	if err != nil {
		return fmt.Errorf("failed to read embedded k0s directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no k0s binary found in embedded assets")
	}

	// Find the k0s binary (should be only one file)
	var k0sBinaryName string
	for _, entry := range entries {
		if !entry.IsDir() {
			k0sBinaryName = entry.Name()
			break
		}
	}

	if k0sBinaryName == "" {
		return fmt.Errorf("no k0s binary file found in embedded assets")
	}

	// Open the embedded k0s binary
	srcPath := filepath.Join("k0s", k0sBinaryName)
	srcFile, err := assets.K0sBinary.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open embedded k0s binary: %w", err)
	}
	defer srcFile.Close()

	// Ensure /usr/local/bin exists
	if err := os.MkdirAll("/usr/local/bin", 0755); err != nil {
		return fmt.Errorf("failed to create /usr/local/bin directory: %w", err)
	}

	// Create destination file at /usr/local/bin/k0s
	dstPath := "/usr/local/bin/k0s"
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create k0s binary at %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy the binary
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy k0s binary: %w", err)
	}

	if i.debug {
		fmt.Printf("✓ Extracted k0s binary: %s → %s\n", k0sBinaryName, dstPath)
	}

	return nil
}

// GetRegistryAddress returns the configured registry address or default
func (i *Installer) GetRegistryAddress() string {
	if i.config.Airgap.Registry.Address != "" {
		return i.config.Airgap.Registry.Address
	}
	// Default to localhost:5000
	return "localhost:5000"
}

// GetBundlePath returns the configured bundle path
func (i *Installer) GetBundlePath() string {
	return i.config.Airgap.BundlePath
}

// IsRegistryInsecure returns whether to use insecure registry connections
func (i *Installer) IsRegistryInsecure() bool {
	// Default to true for localhost registries
	if i.config.Airgap.Registry.Address == "" ||
		i.config.Airgap.Registry.Address == "localhost:5000" ||
		i.config.Airgap.Registry.Address == "127.0.0.1:5000" {
		return true
	}
	return i.config.Airgap.Registry.Insecure
}

// Install performs the complete airgap installation
// This is called by pkg/installer.installAirgap()
func (i *Installer) Install(ctx context.Context) error {
	logger := utils.GetLogger()

	logger.Info("🚀 Starting airgap installation")

	// Step 1: Extract k0rdent version from bundle (if bundle path is configured)
	var k0rdentVersion string
	if i.GetBundlePath() != "" {
		fmt.Printf("Extracting k0rdent version from bundle: %s", i.GetBundlePath())
		version, err := bundle.ExtractK0rdentVersion(i.GetBundlePath())
		if err != nil {
			logger.Warnf("⚠️ Failed to extract version from bundle: %v. Using config version.", err)
			k0rdentVersion = i.config.K0rdent.Version
		} else {
			k0rdentVersion = version
			fmt.Printf("\r✅ K0rdent version from bundle: %s\033[K\n", k0rdentVersion)
			// Update config with extracted version
			i.config.K0rdent.Version = k0rdentVersion
		}
	} else {
		k0rdentVersion = i.config.K0rdent.Version
		fmt.Printf("Using k0rdent version from config: %s", k0rdentVersion)
	}

	// Step 2: Extract k0s binary from embedded assets
	fmt.Printf("Extracting k0s binary from embedded assets...")
	if err := i.ExtractK0sBinary(); err != nil {
		return fmt.Errorf("failed to extract k0s binary: %w", err)
	}
	fmt.Printf("\r✅ K0s binary extracted to /usr/local/bin/k0s\033[K\n")

	// Step 3: Generate k0s configuration for airgap mode
	fmt.Printf("Generating k0s configuration for airgap mode...")
	registryAddr := i.GetRegistryAddress()
	insecure := i.IsRegistryInsecure()

	k0sConfigBytes, err := generator.GenerateAirgapK0sConfig(i.config, registryAddr, insecure)
	if err != nil {
		return fmt.Errorf("failed to generate airgap k0s config: %w", err)
	}

	if i.debug {
		logger.Debugf("Generated k0s config:\n%s", string(k0sConfigBytes))
	}

	fmt.Printf("\r✅ K0s configuration generated (registry: %s, insecure: %t)\033[K\n", registryAddr, insecure)

	// Step 4: Configure containerd registry mirror
	fmt.Printf("Configuring containerd registry mirror...")
	if err := containerd.SetupContainerdMirror(registryAddr); err != nil {
		return fmt.Errorf("failed to configure containerd mirror: %w", err)
	}
	fmt.Printf("\r✅ Containerd registry mirror configured (local registry: %s)\033[K\n", registryAddr)

	// Step 5: Write k0s configuration
	fmt.Printf("Writing k0s configuration to /etc/k0s/k0s.yaml...")
	if err := i.writeK0sConfig(k0sConfigBytes); err != nil {
		return fmt.Errorf("failed to write k0s config: %w", err)
	}

	fmt.Printf("\r✅ K0s configuration written\033[K\n")

	logger.Info("\r✅ Airgap installation preparation complete")

	return nil
}

// writeK0sConfig writes K0s configuration to file
func (i *Installer) writeK0sConfig(config []byte) error {
	configPath := "/etc/k0s/k0s.yaml"

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write configuration file
	if err := os.WriteFile(configPath, config, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if i.debug {
		utils.GetLogger().Debugf("📄 Wrote K0s configuration to %s", configPath)
	}

	return nil
}
