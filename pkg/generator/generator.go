package generator

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"gopkg.in/yaml.v3"
)

const defaultK0rdentHelmReleaseName = "kcm"

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
	API        *K0sAPISpec       `yaml:"api,omitempty"`
	Network    *K0sNetworkSpec   `yaml:"network,omitempty"`
	Storage    *K0sStorageSpec   `yaml:"storage,omitempty"`
	Images     *K0sImagesSpec    `yaml:"images,omitempty"`
	Extensions K0sExtensionsSpec `yaml:"extensions"`
}

// K0sAPISpec represents K0s API specification
type K0sAPISpec struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// K0sNetworkSpec represents K0s network specification
type K0sNetworkSpec struct {
	Provider    string `yaml:"provider"`
	PodCIDR     string `yaml:"podCIDR"`
	ServiceCIDR string `yaml:"serviceCIDR"`
}

// K0sStorageSpec represents K0s storage specification
type K0sStorageSpec struct {
	Type string      `yaml:"type"`
	Etcd K0sEtcdSpec `yaml:"etcd,omitempty"`
}

// K0sEtcdSpec represents K0s etcd specification
type K0sEtcdSpec struct {
	PeerAddress string `yaml:"peerAddress"`
}

// K0sImagesSpec represents K0s images specification
type K0sImagesSpec struct {
	DefaultPullPolicy string `yaml:"default_pull_policy,omitempty"`
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
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Insecure *bool  `yaml:"insecure,omitempty"`
	CAFile   string `yaml:"caFile,omitempty"`
}

// K0sHelmChart represents a helm chart to install
type K0sHelmChart struct {
	Name      string `yaml:"name"`
	Chartname string `yaml:"chartname"`
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"`
	Values    string `yaml:"values"`
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
							Name:      defaultK0rdentHelmReleaseName,
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

	// Only populate API spec if fields are set
	if cfg.K0s.API.Address != "" || cfg.K0s.API.Port != 0 {
		k0sConfig.Spec.API = &K0sAPISpec{
			Address: cfg.K0s.API.Address,
			Port:    cfg.K0s.API.Port,
		}
	}

	// Only populate Network spec if fields are set
	if cfg.K0s.Network.Provider != "" || cfg.K0s.Network.PodCIDR != "" || cfg.K0s.Network.ServiceCIDR != "" {
		k0sConfig.Spec.Network = &K0sNetworkSpec{
			Provider:    cfg.K0s.Network.Provider,
			PodCIDR:     cfg.K0s.Network.PodCIDR,
			ServiceCIDR: cfg.K0s.Network.ServiceCIDR,
		}
	}

	// Only populate Storage spec if fields are set
	if cfg.K0s.Storage.Type != "" || cfg.K0s.Storage.Etcd.PeerAddress != "" {
		storageSpec := &K0sStorageSpec{
			Type: cfg.K0s.Storage.Type,
		}
		if cfg.K0s.Storage.Etcd.PeerAddress != "" {
			storageSpec.Etcd = K0sEtcdSpec{
				PeerAddress: cfg.K0s.Storage.Etcd.PeerAddress,
			}
		}
		k0sConfig.Spec.Storage = storageSpec
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

// GenerateAirgapK0sConfig generates K0s configuration for airgap installation
// registryAddr: local registry address (e.g., "localhost:5000")
// insecure: whether to skip TLS verification for the local registry
func GenerateAirgapK0sConfig(cfg *config.K0rdentdConfig, registryAddr string, insecure bool) ([]byte, error) {
	// Get k0rdent version from build metadata or config
	k0rdentVersion := cfg.K0rdent.Version
	if k0rdentVersion == "" {
		k0rdentVersion = "1.2.2" // default fallback
	}

	// Build OCI chart URL for local registry
	chartURL := fmt.Sprintf("oci://%s/charts/k0rdent-enterprise", registryAddr)

	// Create K0s cluster configuration
	k0sConfig := K0sClusterConfig{
		APIVersion: "k0s.k0sproject.io/v1beta1",
		Kind:       "ClusterConfig",
		Metadata: struct {
			Name string `yaml:"name"`
		}{
			Name: "k0s",
		},
		Spec: K0sClusterSpec{
			Extensions: K0sExtensionsSpec{
				Helm: K0sHelmExtensions{
					Repositories: []K0sHelmRepository{},
					Charts: []K0sHelmChart{
						{
							Name:      defaultK0rdentHelmReleaseName,
							Chartname: chartURL,
							Version:   k0rdentVersion,
							Namespace: cfg.K0rdent.Helm.Namespace,
							Values:    formatAirgapHelmValues(cfg.K0rdent.Helm.Values, registryAddr),
						},
					},
				},
			},
		},
	}

	// Add registry repository if needed (for OCI with TLS)
	if !insecure {
		// For secure registries, we might need to add a repository entry
		// For now, we skip this as local registries are typically insecure
	}

	// Only populate API spec if fields are set
	if cfg.K0s.API.Address != "" || cfg.K0s.API.Port != 0 {
		k0sConfig.Spec.API = &K0sAPISpec{
			Address: cfg.K0s.API.Address,
			Port:    cfg.K0s.API.Port,
		}
	}

	// Only populate Network spec if fields are set
	if cfg.K0s.Network.Provider != "" || cfg.K0s.Network.PodCIDR != "" || cfg.K0s.Network.ServiceCIDR != "" {
		k0sConfig.Spec.Network = &K0sNetworkSpec{
			Provider:    cfg.K0s.Network.Provider,
			PodCIDR:     cfg.K0s.Network.PodCIDR,
			ServiceCIDR: cfg.K0s.Network.ServiceCIDR,
		}
	}

	// Only populate Storage spec if fields are set
	if cfg.K0s.Storage.Type != "" || cfg.K0s.Storage.Etcd.PeerAddress != "" {
		storageSpec := &K0sStorageSpec{
			Type: cfg.K0s.Storage.Type,
		}
		if cfg.K0s.Storage.Etcd.PeerAddress != "" {
			storageSpec.Etcd = K0sEtcdSpec{
				PeerAddress: cfg.K0s.Storage.Etcd.PeerAddress,
			}
		}
		k0sConfig.Spec.Storage = storageSpec
	}

	// Marshal to YAML
	configBytes, err := yaml.Marshal(k0sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal airgap K0s config: %w", err)
	}

	return configBytes, nil
}

// formatAirgapHelmValues formats helm values for airgap installation
// According to k0rdent airgap installation documentation, each component's image
// repository must be explicitly configured to point to the local registry.
func formatAirgapHelmValues(values map[string]interface{}, registryAddr string) string {
	if values == nil {
		values = make(map[string]interface{})
	}

	// Build the airgap configuration structure
	airgapValues := buildAirgapValues(registryAddr)

	// Merge with any user-provided values (user values take precedence)
	mergedValues := mergeValues(airgapValues, values)

	// Marshal to YAML
	valuesBytes, _ := yaml.Marshal(mergedValues)
	return string(valuesBytes)
}

// buildAirgapValues creates the complete airgap values structure
// Based on k0rdent enterprise airgap installation documentation:
// https://docs.mirantis.com/k0rdent-enterprise/latest/admin/installation/airgap/airgap-install/
func buildAirgapValues(registryAddr string) map[string]interface{} {
	return map[string]interface{}{
		// Controller configuration for template repository and global registry
		"controller": map[string]interface{}{
			"templatesRepoURL": fmt.Sprintf("oci://%s/charts", registryAddr),
			"globalRegistry":   registryAddr,
		},
		// KCM controller image
		"image": map[string]interface{}{
			"repository": fmt.Sprintf("%s/kcm-controller", registryAddr),
		},
		// Flux2 images
		"flux2": map[string]interface{}{
			"helmController": map[string]interface{}{
				"image": fmt.Sprintf("%s/fluxcd/helm-controller", registryAddr),
			},
			"sourceController": map[string]interface{}{
				"image": fmt.Sprintf("%s/fluxcd/source-controller", registryAddr),
			},
			"cli": map[string]interface{}{
				"image": fmt.Sprintf("%s/fluxcd/flux-cli", registryAddr),
			},
		},
		// Regional components configuration
		"regional": map[string]interface{}{
			"telemetry": map[string]interface{}{
				"mode": "disabled", // Disable telemetry in airgap mode
				"controller": map[string]interface{}{
					"image": map[string]interface{}{
						"repository": fmt.Sprintf("%s/kcm-telemetry", registryAddr),
					},
				},
			},
			"cert-manager": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": fmt.Sprintf("%s/jetstack/cert-manager-controller", registryAddr),
				},
				"webhook": map[string]interface{}{
					"image": map[string]interface{}{
						"repository": fmt.Sprintf("%s/jetstack/cert-manager-webhook", registryAddr),
					},
				},
				"cainjector": map[string]interface{}{
					"image": map[string]interface{}{
						"repository": fmt.Sprintf("%s/jetstack/cert-manager-cainjector", registryAddr),
					},
				},
				"startupapicheck": map[string]interface{}{
					"image": map[string]interface{}{
						"repository": fmt.Sprintf("%s/jetstack/cert-manager-startupapicheck", registryAddr),
					},
				},
			},
			"cluster-api-operator": map[string]interface{}{
				"image": map[string]interface{}{
					"manager": map[string]interface{}{
						"repository": fmt.Sprintf("%s/capi-operator/cluster-api-operator", registryAddr),
					},
				},
			},
			"velero": map[string]interface{}{
				"image": map[string]interface{}{
					"repository": fmt.Sprintf("%s/velero/velero", registryAddr),
				},
			},
		},
		// RBAC manager
		"rbac-manager": map[string]interface{}{
			"enabled": true,
			"image": map[string]interface{}{
				"repository": fmt.Sprintf("%s/reactiveops/rbac-manager", registryAddr),
			},
		},
		// K0rdent UI
		"k0rdent-ui": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": fmt.Sprintf("%s/k0rdent-ui", registryAddr),
			},
		},
		// Datasource controller
		"datasourceController": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": fmt.Sprintf("%s/datasource-controller", registryAddr),
			},
		},
	}
}

// mergeValues recursively merges source into target (target values take precedence)
func mergeValues(source, target map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy all from source first
	for k, v := range source {
		result[k] = v
	}

	// Merge target values (overwriting or deep merging)
	for k, v := range target {
		if srcVal, exists := result[k]; exists {
			// Both exist, try to deep merge if both are maps
			if srcMap, ok := srcVal.(map[string]interface{}); ok {
				if tgtMap, ok := v.(map[string]interface{}); ok {
					result[k] = mergeValues(srcMap, tgtMap)
					continue
				}
			}
		}
		// Otherwise target value takes precedence
		result[k] = v
	}

	return result
}
