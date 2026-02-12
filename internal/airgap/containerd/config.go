// Package containerd provides containerd registry mirror configuration for airgap installations
package containerd

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// ContainerdDropInDir is the directory for containerd drop-in configs
	ContainerdDropInDir = "/etc/k0s/containerd.d"

	// CertsDir is the directory for registry host configurations
	CertsDir = "/etc/k0s/containerd.d/certs.d"

	// CRIRegistryConfigPath is the path to the CRI registry config
	CRIRegistryConfigPath = "/etc/k0s/containerd.d/cri-registry.toml"
)

// CRIRegistryConfig returns the content of the CRI registry config
func CRIRegistryConfig() string {
	return `version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
config_path = "/etc/k0s/containerd.d/certs.d"
`
}

// HostsConfig returns the content of a hosts.toml for a given registry
// registry: the upstream registry (e.g., "quay.io", "registry.k8s.io")
// mirrorAddr: the local registry address (e.g., "127.0.0.1:5000")
func HostsConfig(registry, mirrorAddr string) string {
	return fmt.Sprintf(`server = "https://%s"

[host."http://%s"]
  capabilities = ["pull", "resolve"]
`, registry, mirrorAddr)
}

// SetupContainerdMirror configures containerd to use the local registry as a mirror
func SetupContainerdMirror(mirrorAddr string) error {
	// Create directories
	if err := os.MkdirAll(CertsDir, 0755); err != nil {
		return fmt.Errorf("failed to create certs.d directory: %w", err)
	}

	// Write CRI registry config
	if err := os.WriteFile(CRIRegistryConfigPath, []byte(CRIRegistryConfig()), 0644); err != nil {
		return fmt.Errorf("failed to write CRI registry config: %w", err)
	}

	// Configure mirrors for known registries
	registries := []string{"registry.k8s.io", "quay.io"}
	for _, registry := range registries {
		regDir := filepath.Join(CertsDir, registry)
		if err := os.MkdirAll(regDir, 0755); err != nil {
			return fmt.Errorf("failed to create registry dir for %s: %w", registry, err)
		}

		hostsPath := filepath.Join(regDir, "hosts.toml")
		hostsContent := HostsConfig(registry, mirrorAddr)
		if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
			return fmt.Errorf("failed to write hosts.toml for %s: %w", registry, err)
		}
	}

	return nil
}
