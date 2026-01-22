package generator

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"gopkg.in/yaml.v3"
)

// K0sClusterConfig represents the K0s cluster configuration structure
type K0sClusterConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec K0sClusterSpec `yaml:"spec"`
}

// K0sClusterSpec represents the K0s cluster specification
type K0sClusterSpec struct {
	API      K0sAPISpec      `yaml:"api"`
	Network  K0sNetworkSpec  `yaml:"network"`
	Storage  K0sStorageSpec  `yaml:"storage"`
	Extensions K0sExtensionsSpec `yaml:"extensions"`
}

// K0sAPISpec represents K0s API specification
type K0sAPISpec struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// K0sNetworkSpec represents K0s network specification
type K0sNetworkSpec struct {
	Provider  string `yaml:"provider"`
	PodCIDR   string `yaml:"podCIDR"`
	ServiceCIDR string `yaml:"serviceCIDR"`
}

// K0sStorageSpec represents K0s storage specification
type K0sStorageSpec struct {
	Type string `yaml:"type"`
	Etcd K0sEtcdSpec `yaml:"etcd,omitempty"`
}

// K0sEtcdSpec represents K0s etcd specification
type K0sEtcdSpec struct {
	PeerAddress string `yaml:"peerAddress"`
}

// K0sExtensionsSpec represents K0s extensions specification
type K0sExtensionsSpec struct {
	Helm K0sHelmExtensions `yaml:"helm"`
}

// K0sHelmExtensions represents K0s helm extensions
type K0sHelmExtensions struct {
	Repositories []K0sHelmRepository `yaml:"repositories"`
	Charts       []K0sHelmChart      `yaml:"charts"`
}

// K0sHelmRepository represents a helm repository
type K0sHelmRepository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// K0sHelmChart represents a helm chart to install
type K0sHelmChart struct {
	Name      string                 `yaml:"name"`
	Chartname string                 `yaml:"chartname"`
	Version   string                 `yaml:"version"`
	Namespace string                 `yaml:"namespace"`
	Values    string                 `yaml:"values"`
}

// GenerateK0sConfig generates K0s configuration from k0rdentd configuration
func GenerateK0sConfig(cfg *config.K0rdentdConfig) ([]byte, error) {
	// Create K0s cluster configuration
	k0sConfig := K0sClusterConfig{
		APIVersion: "k0s.k0sproject.io/v1beta1",
		Kind:       "Cluster",
		Metadata: struct {
			Name string `yaml:"name"`
		}{
			Name: "k0s",
		},
		Spec: K0sClusterSpec{
			API: K0sAPISpec{
				Address: cfg.K0s.API.Address,
				Port:    cfg.K0s.API.Port,
			},
			Network: K0sNetworkSpec{
				Provider:  cfg.K0s.Network.Provider,
				PodCIDR:   cfg.K0s.Network.PodCIDR,
				ServiceCIDR: cfg.K0s.Network.ServiceCIDR,
			},
			Storage: K0sStorageSpec{
				Type: cfg.K0s.Storage.Type,
				Etcd: K0sEtcdSpec{
					PeerAddress: cfg.K0s.Storage.Etcd.PeerAddress,
				},
			},
			Extensions: K0sExtensionsSpec{
				Helm: K0sHelmExtensions{
					Repositories: []K0sHelmRepository{
						{
							Name: "k0rdent",
							URL:  "https://charts.k0rdent.io",
						},
					},
					Charts: []K0sHelmChart{
						{
							Name:      "k0rdent",
							Chartname: cfg.K0rdent.Helm.Chart,
							Version:   cfg.K0rdent.Version,
							Namespace: cfg.K0rdent.Helm.Namespace,
							Values:    formatHelmValues(cfg.K0rdent.Helm.Values),
						},
					},
				},
			},
		},
	}

	// Marshal to YAML
	configBytes, err := yaml.Marshal(k0sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal K0s config: %w", err)
	}

	return configBytes, nil
}

// formatHelmValues formats helm values as YAML string
func formatHelmValues(values map[string]interface{}) string {
	if values == nil {
		return ""
	}
	valuesBytes, _ := yaml.Marshal(values)
	return string(valuesBytes)
}