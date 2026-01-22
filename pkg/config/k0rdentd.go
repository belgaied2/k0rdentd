package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// K0rdentdConfig represents the main configuration structure
type K0rdentdConfig struct {
	K0s      K0sConfig     `yaml:"k0s"`
	K0rdent  K0rdentConfig `yaml:"k0rdent"`
	Debug    bool          `yaml:"debug,omitempty"`
	LogLevel string        `yaml:"logLevel,omitempty"`
}

// K0sConfig represents K0s-specific configuration
type K0sConfig struct {
	Version string        `yaml:"version"`
	API     APIConfig     `yaml:"api"`
	Network NetworkConfig `yaml:"network"`
	Storage StorageConfig `yaml:"storage"`
}

// APIConfig represents API server configuration
type APIConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Provider    string `yaml:"provider"`
	PodCIDR     string `yaml:"podCIDR"`
	ServiceCIDR string `yaml:"serviceCIDR"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Type string     `yaml:"type"`
	Etcd EtcdConfig `yaml:"etcd,omitempty"`
}

// EtcdConfig represents etcd-specific configuration
type EtcdConfig struct {
	PeerAddress string `yaml:"peerAddress"`
}

// K0rdentConfig represents K0rdent-specific configuration
type K0rdentConfig struct {
	Version string            `yaml:"version"`
	Helm    K0rdentHelmConfig `yaml:"helm"`
}

// K0rdentHelmConfig represents K0rdent helm chart configuration
type K0rdentHelmConfig struct {
	Chart     string                 `yaml:"chart"`
	Namespace string                 `yaml:"namespace"`
	Values    map[string]interface{} `yaml:"values,omitempty"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*K0rdentdConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg K0rdentdConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults if not provided
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}

// MarshalConfig marshals configuration to YAML
func MarshalConfig(cfg *K0rdentdConfig) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// WriteConfigFile writes configuration to file
func WriteConfigFile(path string, data []byte) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(getConfigDir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file with appropriate permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *K0rdentdConfig {
	return &K0rdentdConfig{
		K0s: K0sConfig{
			API: APIConfig{
				Address: "0.0.0.0",
				Port:    6443,
			},
			Network: NetworkConfig{
				Provider:    "calico",
				PodCIDR:     "10.244.0.0/16",
				ServiceCIDR: "10.96.0.0/12",
			},
			Storage: StorageConfig{
				Type: "etcd",
				Etcd: EtcdConfig{
					PeerAddress: "127.0.0.1",
				},
			},
		},
		K0rdent: K0rdentConfig{
			Version: "v0.1.0",
			Helm: K0rdentHelmConfig{
				Chart:     "k0rdent/k0rdent",
				Namespace: "k0rdent-system",
				Values: map[string]interface{}{
					"replicaCount": 1,
					"service": map[string]interface{}{
						"type": "ClusterIP",
						"port": 80,
					},
				},
			},
		},
		Debug:    false,
		LogLevel: "info",
	}
}

// getConfigDir extracts directory from file path
func getConfigDir(path string) string {
	// Simple implementation - in production, use filepath.Dir
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
