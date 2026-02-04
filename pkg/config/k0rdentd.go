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
	Version     string            `yaml:"version"`
	Helm        K0rdentHelmConfig `yaml:"helm"`
	Credentials CredentialsConfig `yaml:"credentials,omitempty"`
}

// CredentialsConfig holds credentials for all cloud providers
type CredentialsConfig struct {
	AWS       []AWSCredential       `yaml:"aws,omitempty"`
	Azure     []AzureCredential     `yaml:"azure,omitempty"`
	OpenStack []OpenStackCredential `yaml:"openstack,omitempty"`
}

// HasCredentials returns true if any credentials are configured
func (c CredentialsConfig) HasCredentials() bool {
	return len(c.AWS) > 0 || len(c.Azure) > 0 || len(c.OpenStack) > 0
}

// AWSCredential represents AWS credentials
type AWSCredential struct {
	Name            string `yaml:"name"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	SessionToken    string `yaml:"sessionToken,omitempty"` // Optional: for MFA or SSO
}

// AzureCredential represents Azure Service Principal credentials
type AzureCredential struct {
	Name           string `yaml:"name"`
	SubscriptionID string `yaml:"subscriptionID"`
	ClientID       string `yaml:"clientID"`
	ClientSecret   string `yaml:"clientSecret"`
	TenantID       string `yaml:"tenantID"`
}

// OpenStackCredential represents OpenStack credentials
type OpenStackCredential struct {
	Name                        string `yaml:"name"`
	AuthURL                     string `yaml:"authURL"`
	Region                      string `yaml:"region"`
	ApplicationCredentialID     string `yaml:"applicationCredentialID,omitempty"`
	ApplicationCredentialSecret string `yaml:"applicationCredentialSecret,omitempty"`
	Username                    string `yaml:"username,omitempty"`
	Password                    string `yaml:"password,omitempty"`
	ProjectName                 string `yaml:"projectName,omitempty"`
	DomainName                  string `yaml:"domainName,omitempty"`
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

// LoadConfigWithFallback loads configuration with fallback logic:
// 1. If explicitPath is provided and non-empty, use it (fail if invalid)
// 2. If defaultPath exists and is non-empty, use it
// 3. Otherwise, return DefaultConfig()
func LoadConfigWithFallback(explicitPath string, defaultPath string, explicitSet bool) (*K0rdentdConfig, error) {
	// Case 1: -c flag is explicitly provided - must use it and fail if anything is wrong
	if explicitSet && explicitPath != "" {
		return LoadConfig(explicitPath)
	}

	// Case 2: Check if default config file exists and is non-empty
	if defaultPath != "" {
		info, err := os.Stat(defaultPath)
		if err == nil && info.Size() > 0 {
			// File exists and is non-empty, try to load it
			return LoadConfig(defaultPath)
		}
	}

	// Case 3: No explicit path and no valid default file - use defaults
	return DefaultConfig(), nil
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
		K0s: K0sConfig{},
		K0rdent: K0rdentConfig{
			Version: "1.2.2",
			Helm: K0rdentHelmConfig{
				Chart:     "oci://registry.mirantis.com/k0rdent-enterprise/charts/k0rdent-enterprise",
				Namespace: "kcm-system",
				Values: map[string]interface{}{
					"replicaCount": 1,
					"k0rdent-ui": map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
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
